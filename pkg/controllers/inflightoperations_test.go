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
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/ktesting"

	v1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/apis/v1alpha1/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned/fake"
	informers "k8s.io/kubevirt-flight-viewer/pkg/generated/informers/externalversions"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	client     *fake.Clientset
	kubeclient *k8sfake.Clientset
	// Objects to put in the store.
	inflightOperationsLister []*v1alpha1.InFlightOperation

	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newInFlightOperation(name string, replicas *int32) *v1alpha1.InFlightOperation {
	return &v1alpha1.InFlightOperation{
		TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		// TODO set status
	}
}

func (f *fixture) newController(ctx context.Context) (*Controller, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())

	c := NewController(
		ctx,
		f.kubeclient,
		f.client,
		i.Kubevirtflightviewer().V1alpha1().InFlightOperations(),
		i.Kubevirtflightviewer().V1alpha1().InFlightlusterOperations(),
		map[string]cache.SharedIndexInformer{})

	c.recorder = &record.FakeRecorder{}

	for _, f := range f.inflightOperationsLister {
		i.Kubevirtflightviewer().V1alpha1().InFlightOperations().Informer().GetIndexer().Add(f)
	}

	return c, i
}
func (f *fixture) expectCreateIFOAction(i *v1alpha1.InFlightOperation) {
	f.kubeactions = append(f.kubeactions, core.NewCreateAction(schema.GroupVersionResource{Resource: "inflightoperations"}, i.Namespace, i))
}

func TestCreatesIFO(t *testing.T) {
	f := newFixture(t)
	ifo := newInFlightOperation("test")
	_, ctx := ktesting.NewTestContext(t)

	f.inflightOperationsLister = append(f.inflightOperationsLister, ifo)
	f.objects = append(f.objects, ifo)

	expectedIFO := v1alpha1.InFlightOperation{}
	f.expectCreateIFOAction(expectedIFO)

	controller := f.newController(ctx)

	vmi := *virtv1.VirtualMachineInstance{}

	controller.reconcileRegistrations(ctx, vmi, key)
}

/*
func (f *fixture) run(ctx context.Context, ifoRef cache.ObjectName) {
	f.runController(ctx, ifoRef, true, false)
}

func (f *fixture) runExpectError(ctx context.Context, ifoRef cache.ObjectName) {
	f.runController(ctx, ifoRef, true, true)
}

func (f *fixture) runController(ctx context.Context, ifoRef cache.ObjectName, startInformers bool, expectError bool) {
	c, i, k8sI := f.newController(ctx)
	if startInformers {
		i.Start(ctx.Done())
		k8sI.Start(ctx.Done())
	}

	err := c.syncHandler(ctx, ifoRef)
	if !expectError && err != nil {
		f.t.Errorf("error syncing ifo: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing ifo, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}

	k8sActions := filterInformerActions(f.kubeclient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "ifos") ||
				action.Matches("watch", "ifos") ||
				action.Matches("list", "deployments") ||
				action.Matches("watch", "deployments")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectUpdateDeploymentAction(d *apps.Deployment) {
	f.kubeactions = append(f.kubeactions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "deployments"}, d.Namespace, d))
}

func (f *fixture) expectUpdateInFlightOperationStatusAction(ifo *v1alpha1.InFlightOperation) {
	action := core.NewUpdateSubresourceAction(schema.GroupVersionResource{Resource: "ifos"}, "status", ifo.Namespace, ifo)
	f.actions = append(f.actions, action)
}

func getRef(ifo *v1alpha1.InFlightOperation, t *testing.T) cache.ObjectName {
	ref := cache.MetaObjectToName(ifo)
	return ref
}
*/
