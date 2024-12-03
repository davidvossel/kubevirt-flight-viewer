package kubevirt

import (
	"context"

	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtv1 "kubevirt.io/api/core/v1"
)

func RegisterOperation() {
	controllers.RegisterOperation(&operation{}, "Updating", "kubevirts", virtv1.KubeVirtGroupVersionKind)
}

type operation struct {
}

func (o *operation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	//logger := klog.FromContext(ctx)

	//kv := obj.(*virtv1.KubeVirt)

	// TODO implement this
	//	logger.Info(fmt.Sprintf("processing starting operation for kv [%s]", kv.Name))

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
