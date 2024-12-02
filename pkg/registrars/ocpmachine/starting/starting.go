package starting

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (s *startingOperation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	machine := obj.(*machinev1beta1.Machine)
	logger.Info(fmt.Sprintf("processing starting operation for machine [%s]", machine.Name))

	// return empty conditions when no starting is in progress
	// This signals no in-flight starting is taking place
	if machine.DeletionTimestamp == nil {
		return []metav1.Condition{}
	}

	/*
		condition := meta.FindStatusCondition(conditions, "Progressing")
		if condition == nil {
			condition = &metav1.Condition{
				Type:               "Progressing",
				ObservedGeneration: machine.Generation,
				Status:             metav1.ConditionTrue,
				Reason:             "Stopping",
				Message:            fmt.Sprintf("VM is terminating"),
				LastTransitionTime: metav1.NewTime(time.Now()),
			}
		}

		meta.SetStatusCondition(&conditions, *condition)
	*/

	return conditions
}
