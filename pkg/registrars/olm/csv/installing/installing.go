package starting

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RegisterOperation() {
	controllers.RegisterOperation(&operation{}, "Installing", "clusterserviceversions", schema.GroupVersionKind{
		Version: olmv1alpha1.SchemeGroupVersion.Version,
		Group:   olmv1alpha1.SchemeGroupVersion.Group,
		Kind:    "ClusterServiceVersion",
	})
}

type operation struct {
}

func (o *operation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	csv := obj.(*olmv1alpha1.ClusterServiceVersion)
	logger.Info(fmt.Sprintf("processing installing operation for csv [%s]", csv.Name))

	return []metav1.Condition{}

	// TODO implement this, use template below
	/*
		if vmi.DeletionTimestamp != nil {
			// return empty conditions when no stopping is in progress
			// This signals no in-flight stopping is taking place
			return []metav1.Condition{}
		} else if vmi.Status.Phase != virtv1.VmPhaseUnset &&
			vmi.Status.Phase != virtv1.Pending &&
			vmi.Status.Phase != virtv1.Scheduling &&
			vmi.Status.Phase != virtv1.Scheduled {

			// return empty conditions when no stopping is in progress
			// This signals no in-flight stopping is taking place
		}

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
