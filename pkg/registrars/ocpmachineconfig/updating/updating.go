package updating

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	machineconfigv1 "github.com/openshift/api/machineconfiguration/v1"
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

func (o *operation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	pool := obj.(*machineconfigv1.MachineConfigPool)
	logger.Info(fmt.Sprintf("processing starting operation for machine config pool [%s]", pool.Name))

	return []metav1.Condition{}
	/*
		TODO check for updating and use similar logic to this vmi condition
				condition := meta.FindStatusCondition(conditions, "Progressing")
				if condition == nil {
					condition = &metav1.Condition{
						Type:               "Progressing",
						ObservedGeneration: vmi.Generation,
						Status:             metav1.ConditionTrue,
						Reason:             "Starting",
						Message:            fmt.Sprintf("Starting vm ", vmi.Name),
						LastTransitionTime: metav1.NewTime(time.Now()),
					}
				}

				meta.SetStatusCondition(&conditions, *condition)
			return conditions
	*/

}
