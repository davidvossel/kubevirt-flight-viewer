package ocpmachineconfig

import (
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	machineconfigv1 "github.com/openshift/api/machineconfiguration/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (o *operation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	pool := obj.(*machineconfigv1.MachineConfigPool)
	logger.Info(fmt.Sprintf("processing updating operation for machine config pool [%s]", pool.Name))

	updatingCondition := findStatusCondition(pool.Status.Conditions, "Updating")
	if updatingCondition != nil && string(updatingCondition.Status) == "True" {
		condition := meta.FindStatusCondition(conditions, "Progressing")
		if condition == nil {
			condition = &metav1.Condition{
				Type:               "Progressing",
				ObservedGeneration: pool.Generation,
				Status:             metav1.ConditionTrue,
				Reason:             "Updating",
				Message:            fmt.Sprintf("Updating machine config pool [%s] ", pool.Name),
				LastTransitionTime: metav1.NewTime(time.Now()),
			}
		}

		meta.SetStatusCondition(&conditions, *condition)
		return conditions
	}

	return []metav1.Condition{}
}
