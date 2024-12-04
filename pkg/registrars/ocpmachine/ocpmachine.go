package ocpmachine

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"

	"k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RegisterOperation() {
	controllers.RegisterOperation(&startingOperation{}, "MachineStarting", "machines", schema.GroupVersionKind{
		Version: machinev1beta1.GroupVersion.Version,
		Group:   machinev1beta1.GroupVersion.Group,
		Kind:    "Machine",
	})
}

type startingOperation struct {
}

func (s *startingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	machine := obj.(*machinev1beta1.Machine)

	// return empty conditions when no starting is in progress
	// This signals no in-flight starting is taking place
	if machine.DeletionTimestamp == nil {
		return nil
	}

	// TODO implement
	//logger := klog.FromContext(ctx)
	//logger.V(4).Info(fmt.Sprintf("processing starting operation for machine [%s]", machine.Name))

	return nil
}
