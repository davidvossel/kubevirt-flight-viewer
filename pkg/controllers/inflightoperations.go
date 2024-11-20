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
	"strings"
	"time"

	"golang.org/x/time/rate"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	v1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
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
	ProcessOperation(context.Context, interface{}, []metav1.Condition) []metav1.Condition
}

// TODO sync map or somehow prevent writes after init.
var registrations map[string]registrationObj

type registrationObj struct {
	operationType            string
	resourceType             string
	resourceGroupVersionKind schema.GroupVersionKind
	registration             InFlightOperationRegistration
}

func RegisterOperation(registration InFlightOperationRegistration, operationType string, resourceType string, resourceGroupVersionKind schema.GroupVersionKind) error {

	if registrations == nil {
		registrations = map[string]registrationObj{}
	}

	registrations[operationType] = registrationObj{
		operationType:            operationType,
		resourceType:             resourceType,
		registration:             registration,
		resourceGroupVersionKind: resourceGroupVersionKind,
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
	// fvclientset is a clientset for our own API group
	fvclientset clientset.Interface

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

func (c *Controller) processOldAndNewOperations(ctx context.Context, curOp *v1alpha1.InFlightOperation,
	oldOps []v1alpha1.InFlightOperation) error {

	var err error
	var origOp *v1alpha1.InFlightOperation = nil
	deleteOps := []v1alpha1.InFlightOperation{}

	if curOp == nil || curOp.Name == "" {
		deleteOps = oldOps
	} else if curOp.Name != "" {
		origOp = curOp.DeepCopy()
		for _, op := range oldOps {
			if op.Name != curOp.Name {
				deleteOps = append(deleteOps, op)
			}
		}
	}

	// Create or Update
	if curOp != nil && curOp.Name == "" {
		curOp, err = c.fvclientset.KubevirtflightviewerV1alpha1().InFlightOperations(curOp.Namespace).Create(ctx, curOp, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("error during creating operation: %v", err)
		}
		/*
			// TODO remove spec section entirely if not needed
				} else if !equality.Semantic.DeepEqual(origOp.Spec, curOp.Spec) {
					curOp, err = c.fvclientset.KubevirtflightviewerV1alpha1().InFlightOperations(curOp.Namespace).Update(ctx, curOp, metav1.UpdateOptions{})
					if err != nil {
						return fmt.Errorf("error during updating operation: %v", err)
					}
		*/
	}

	// Update Status
	if origOp != nil && !equality.Semantic.DeepEqual(origOp.Status, curOp.Status) {
		curOp, err = c.fvclientset.KubevirtflightviewerV1alpha1().InFlightOperations(origOp.Namespace).UpdateStatus(ctx, origOp, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error during updating operation status: %v", err)
		}
	}

	for _, op := range deleteOps {
		err = c.fvclientset.KubevirtflightviewerV1alpha1().InFlightOperations(op.Namespace).Delete(ctx, op.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("error during deleting operation: %v", err)
		}
	}

	return nil

}

func currentOperation(ops []v1alpha1.InFlightOperation) *v1alpha1.InFlightOperation {
	curIdx := -1
	for i, op := range ops {
		if curIdx == -1 || ops[curIdx].CreationTimestamp.Before(&op.CreationTimestamp) {
			curIdx = i
		}
	}
	if curIdx == -1 {
		return nil
	}
	return ops[curIdx].DeepCopy()
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
	fvclientset clientset.Interface,
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
		fvclientset:              fvclientset,
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

func (c *Controller) getOperationsForResource(ownerRef *metav1.OwnerReference, namespace string, operationType string) ([]v1alpha1.InFlightOperation, error) {
	// TODO make custom indexer for this
	allInFlightOperations, err := c.inflightOperationsLister.InFlightOperations(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	relatedInFlightOperations := []v1alpha1.InFlightOperation{}

	for _, op := range allInFlightOperations {
		if op.Status.OperationType != operationType {
			continue
		} else if !equality.Semantic.DeepEqual(&op.OwnerReferences[0], ownerRef) {
			continue
		}
		relatedInFlightOperations = append(relatedInFlightOperations, *op.DeepCopy())
	}

	return relatedInFlightOperations, nil
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
		c.logger.Info(fmt.Sprintf("no object found for [%s] of type [%s]", objKey, key.ResourceType))
		return nil
	}

	c.logger.Info(fmt.Sprintf("registration count [%d]", len(registrations)))

	for _, regObj := range registrations {
		if regObj.resourceType == key.ResourceType {
			objMeta := obj.(metav1.Object)

			ownerRef := metav1.NewControllerRef(objMeta, regObj.resourceGroupVersionKind)
			ownerRef.BlockOwnerDeletion = ptr.To(false)

			oldOperations, err := c.getOperationsForResource(ownerRef, objMeta.GetNamespace(), regObj.operationType)
			if err != nil {
				return err
			}

			curOp := currentOperation(oldOperations)
			c.logger.Info(fmt.Sprintf("processing registration: %s for resource: %s", regObj.operationType, key.ResourceType))

			var curConditions []metav1.Condition

			if curOp == nil {
				curConditions = []metav1.Condition{}
			} else {
				curConditions = curOp.Status.Conditions
			}

			curConditions = regObj.registration.ProcessOperation(ctx, obj, curConditions)

			if len(curConditions) == 0 {
				// If no conditions are returned, that's the signal that the
				// operation is complete
				curOp = nil
			} else if curOp == nil {
				curOp = &v1alpha1.InFlightOperation{}
				curOp.GenerateName = strings.ToLower(regObj.operationType)
				curOp.Namespace = objMeta.GetNamespace()
				curOp.OwnerReferences = []metav1.OwnerReference{*ownerRef}
				curOp.Status.OperationType = regObj.operationType
			}

			if curOp != nil {
				curOp.Status.Conditions = curConditions
			}

			err = c.processOldAndNewOperations(ctx, curOp, oldOperations)
			if err != nil {
				return fmt.Errorf("failed to process operation: %v", err)
			}
		}
	}

	return nil
}
