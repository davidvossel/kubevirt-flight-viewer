/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	flightviewerv1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	clientset "k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned"
	flightviewerscheme "k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned/scheme"
	informers "k8s.io/kubevirt-flight-viewer/pkg/generated/informers/externalversions/kubevirtflightviewer/v1alpha1"
	listers "k8s.io/kubevirt-flight-viewer/pkg/generated/listers/kubevirtflightviewer/v1alpha1"
)

const controllerAgentName = "kubevirt-flight-viewer"

const (
	// SuccessSynced is used as part of the Event 'reason' when a InFlightOperation is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a InFlightOperation fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by InFlightOperation"
	// MessageResourceSynced is the message used for an Event fired when a InFlightOperation
	// is synced successfully
	MessageResourceSynced = "InFlightOperation synced successfully"
	// FieldManager distinguishes this controller from other things writing to API objects
	FieldManager = controllerAgentName
)

type InFlightOperationRegistration interface {
	ProcessOperation(context.Context, interface{}, []*flightviewerv1alpha1.InFlightOperation)
}

// TODO sync map or somehow prevent writes after init.
var registrations map[string]registrationObj

type registrationObj struct {
	operationName string
	resourceType  string
	registration  InFlightOperationRegistration
}

func RegisterOperation(registration InFlightOperationRegistration, operationName string, resourceType string) error {

	if registrations == nil {
		registrations = map[string]registrationObj{}
	}

	registrations[operationName] = registrationObj{
		operationName: operationName,
		resourceType:  resourceType,
		registration:  registration,
	}
	return nil
}

type queueKey struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ResourceType string `json:"resourceType"`
}

func queueKeyToString(qk *queueKey) (string, error) {
	jsonData, err := json.Marshal(qk)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func queueStringToKey(qkString string) (*queueKey, error) {

	var qk queueKey
	err := json.Unmarshal([]byte(qkString), &qk)
	if err != nil {
		return nil, err
	}
	return &qk, nil
}

func (c *Controller) getInformer(resourceType string) (cache.SharedIndexInformer, error) {

	informer, ok := c.resourceInformers[resourceType]
	if !ok {
		return nil, fmt.Errorf("unknown shared index informer for resource type [%s]", resourceType)
	}

	return informer, nil
}

// Controller is the controller implementation for InFlightOperation resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// flightviewerclientset is a clientset for our own API group
	flightviewerclientset clientset.Interface

	inflightOperationsLister listers.InFlightOperationLister
	inflightOperationsSynced cache.InformerSynced

	resourceInformers map[string]cache.SharedIndexInformer

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder

	logger klog.Logger
}

func (c *Controller) genericAddHandler(obj interface{}, resourceType string) {
	o := obj.(metav1.Object)

	key, err := queueKeyToString(&queueKey{
		Name:         o.GetName(),
		Namespace:    o.GetNamespace(),
		ResourceType: resourceType,
	})
	if err != nil {
		c.logger.Error(err, fmt.Sprintf("failed to process add handler for resource type %s", resourceType))
		return
	}
	c.workqueue.Add(key)
	c.logger.Info(fmt.Sprintf("Add event for resource type %s", resourceType))
}

func (c *Controller) genericUpdateHandler(old, cur interface{}, resourceType string) {
	curObj := cur.(metav1.Object)
	oldObj := old.(metav1.Object)
	if curObj.GetResourceVersion() == oldObj.GetResourceVersion() {
		return
	}

	key, err := queueKeyToString(&queueKey{
		Name:         curObj.GetName(),
		Namespace:    curObj.GetNamespace(),
		ResourceType: resourceType,
	})
	if err != nil {
		c.logger.Error(err, fmt.Sprintf("failed to process update handler for resource type %s", resourceType))
		return
	}
	c.workqueue.Add(key)
	c.logger.Info(fmt.Sprintf("Update event for resource type %s", resourceType))

}

func validateDeleteObject(obj interface{}) (metav1.Object, error) {
	var o metav1.Object
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		o, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, fmt.Errorf("tombstone contained object that is not a k8s object %#v", obj)
		}
	} else if o, ok = obj.(metav1.Object); !ok {
		return nil, fmt.Errorf("couldn't get object from %+v", obj)
	}
	return o, nil
}

func (c *Controller) genericDeleteHandler(obj interface{}, resourceType string) {
	o, err := validateDeleteObject(obj)
	if err != nil {
		c.logger.Error(err, "Failed to process delete notification")
		return
	}

	key, err := queueKeyToString(&queueKey{
		Name:         o.GetName(),
		Namespace:    o.GetNamespace(),
		ResourceType: resourceType,
	})
	if err != nil {
		c.logger.Error(err, fmt.Sprintf("failed to process delete handler for resource type %s", resourceType))
		return
	}
	c.workqueue.Add(key)
	c.logger.Info(fmt.Sprintf("Delete event for resource type %s", resourceType))
}

// NewController returns a new flightviewer controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	flightviewerclientset clientset.Interface,
	inflightOperationInformer informers.InFlightOperationInformer,
	resourceInformers map[string]cache.SharedIndexInformer) (*Controller, error) {
	logger := klog.FromContext(ctx)

	// Create event broadcaster
	// Add kubevirt-flight-viewer types to the default Kubernetes Scheme so Events can be
	// logged for kubevirt-flight-viewer types.
	utilruntime.Must(flightviewerscheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Every(5*time.Second), 1)},
	)

	controller := &Controller{
		kubeclientset:            kubeclientset,
		flightviewerclientset:    flightviewerclientset,
		inflightOperationsLister: inflightOperationInformer.Lister(),
		inflightOperationsSynced: inflightOperationInformer.Informer().HasSynced,
		workqueue:                workqueue.NewNamedRateLimitingQueue(ratelimiter, "kubevirt-inflight-operation"),
		recorder:                 recorder,
		logger:                   logger,
		resourceInformers:        resourceInformers,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when InFlightOperation resources change
	/*
		inflightOperationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: controller.enqueueInFlightOperation,
			UpdateFunc: func(old, new interface{}) {
				controller.enqueueInFlightOperation(new)
			},
		})
	*/

	for resourceType, informer := range resourceInformers {
		_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				controller.genericAddHandler(obj, resourceType)
			},
			DeleteFunc: func(obj interface{}) {
				controller.genericDeleteHandler(obj, resourceType)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				controller.genericUpdateHandler(oldObj, newObj, resourceType)
			},
		})

		if err != nil {
			return nil, err
		}
	}

	return controller, nil
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	c.logger = logger

	// Start the informer factories to begin populating the informer caches
	c.logger.Info("Starting InFlightOperation controller")

	// Wait for the caches to be synced before starting workers
	c.logger.Info("Waiting for informer caches to sync")

	// TODO sync resource informers

	if ok := cache.WaitForCacheSync(ctx.Done(), c.inflightOperationsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	// Launch two workers to process InFlightOperation resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	c.logger.Info("Started workers")
	<-ctx.Done()
	c.logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the reconcile.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	key, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We call Done at the end of this func so the workqueue knows we have
	// finished processing this item. We also must remember to call Forget
	// if we do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off
	// period.
	defer c.workqueue.Done(key)

	// Run the reconcile, passing it the structured reference to the object to be synced.
	err := c.reconcile(ctx, key.(string))
	if err == nil {
		// If no error occurs then we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(key)
		c.logger.Info("Successfully synced", "objectName", key)
		return true
	}
	// there was a failure so be sure to report it.  This method allows for
	// pluggable error handling which can be used for things like
	// cluster-monitoring.
	utilruntime.HandleErrorWithContext(ctx, err, "Error syncing; requeuing for later retry", "objectReference", key)
	// since we failed, we should requeue the item to work on later.  This
	// method will add a backoff to avoid hotlooping on particular items
	// (they're probably still not going to work right away) and overall
	// controller protection (everything I've done is broken, this controller
	// needs to calm down or it can starve other useful work) cases.
	c.workqueue.AddRateLimited(key)
	return true
}

// reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the InFlightOperation resource
// with the current status of the resource.
func (c *Controller) reconcile(ctx context.Context, keyJSONStr string) error {
	//logger := klog.LoggerWithValues(klog.FromContext(ctx), "objectRef", objectRef)

	c.logger.Info(fmt.Sprintf("processing queue key: %s", keyJSONStr))

	key, err := queueStringToKey(keyJSONStr)
	if err != nil {
		return err
	}

	resourceInformer, err := c.getInformer(key.ResourceType)
	if err != nil {
		return err
	}

	objKey := key.Namespace + "/" + key.Name
	if key.Namespace == "" {
		objKey = key.Name
	}

	obj, exists, _ := resourceInformer.GetStore().GetByKey(objKey)
	if !exists {
		c.logger.Info(fmt.Sprintf("no obect found for [%s] of type [%s]", objKey, key.ResourceType))
		return nil
	}

	c.logger.Info(fmt.Sprintf("registration count [%d]", len(registrations)))

	for _, regObj := range registrations {
		if regObj.resourceType == key.ResourceType {
			c.logger.Info(fmt.Sprintf("processing registration: %s for resource: %s", regObj.operationName, key.ResourceType))
			regObj.registration.ProcessOperation(ctx, obj, nil)
		}
	}

	/*
		// Get the InFlightOperation resource with this namespace/name
		inflightOperation, err := c.inflightOperationsLister.InFlightOperations(objectRef.Namespace).Get(objectRef.Name)
		if err != nil {
			// The InFlightOperation resource may no longer exist, in which case we stop
			// processing.
			if errors.IsNotFound(err) {
				utilruntime.HandleErrorWithContext(ctx, err, "InFlightOperation referenced by item in work queue no longer exists", "objectReference", objectRef)
				return nil
			}

			return err
		}

		err = c.updateInFlightOperationStatus(ctx, inflightOperation)
		if err != nil {
			return err
		}

		c.recorder.Event(inflightOperation, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	*/
	return nil
}
