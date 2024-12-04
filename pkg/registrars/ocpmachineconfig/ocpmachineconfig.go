package ocpmachineconfig

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	machineconfigv1 "github.com/openshift/api/machineconfiguration/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RegisterOperation() {
	controllers.RegisterOperation(&operation{}, "Updating", "machineconfigpools", schema.GroupVersionKind{
		Version: machineconfigv1.GroupVersion.Version,
		Group:   machineconfigv1.GroupVersion.Group,
		Kind:    "MachineConfigPool",
	})
}

type operation struct {
}

func findStatusCondition(conditions []machineconfigv1.MachineConfigPoolCondition, conditionType string) *machineconfigv1.MachineConfigPoolCondition {
	for i := range conditions {
		if string(conditions[i].Type) == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

func (o *operation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	logger := klog.FromContext(ctx)

	pool := obj.(*machineconfigv1.MachineConfigPool)

	updatingCondition := findStatusCondition(pool.Status.Conditions, "Updating")
	if updatingCondition != nil && string(updatingCondition.Status) == "True" {
		logger.V(4).Info(fmt.Sprintf("processing updating operation for machine config pool [%s]", pool.Name))
		return &v1alpha1.InFlightOperationState{
			TransitionState: v1alpha1.TransitionStateProgressing,
			Reason:          "Updating",
			Message:         fmt.Sprintf("Updating machine config pool [%s] ", pool.Name),
		}
	}

	return nil
}
