package csv

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RegisterOperation() {

	controllers.RegisterOperation(&deletingOperation{}, "Deleting", "clusterserviceversions", schema.GroupVersionKind{
		Version: olmv1alpha1.SchemeGroupVersion.Version,
		Group:   olmv1alpha1.SchemeGroupVersion.Group,
		Kind:    "ClusterServiceVersion",
	})
	controllers.RegisterOperation(&installingOperation{}, "Installing", "clusterserviceversions", schema.GroupVersionKind{
		Version: olmv1alpha1.SchemeGroupVersion.Version,
		Group:   olmv1alpha1.SchemeGroupVersion.Group,
		Kind:    "ClusterServiceVersion",
	})
}

type installingOperation struct {
}

func (o *installingOperation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	csv := obj.(*olmv1alpha1.ClusterServiceVersion)
	logger.Info(fmt.Sprintf("processing installing operation for csv [%s]", csv.Name))

	switch csv.Status.Phase {
	case olmv1alpha1.CSVPhaseSucceeded, olmv1alpha1.CSVPhaseFailed, olmv1alpha1.CSVPhaseDeleting, olmv1alpha1.CSVPhaseUnknown, olmv1alpha1.CSVPhaseReplacing:
		// not installing if in any of these phases
		return []metav1.Condition{}
	}

	condition := meta.FindStatusCondition(conditions, "Progressing")
	if condition == nil {
		condition = &metav1.Condition{
			Type:               "Progressing",
			ObservedGeneration: csv.Generation,
			Status:             metav1.ConditionTrue,
			Reason:             "Installing",
			Message:            fmt.Sprintf("phase [%s] ", csv.Status.Phase),
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}

	meta.SetStatusCondition(&conditions, *condition)

	return conditions
}

type deletingOperation struct {
}

func (o *deletingOperation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	csv := obj.(*olmv1alpha1.ClusterServiceVersion)
	logger.Info(fmt.Sprintf("processing deleting operation for csv [%s]", csv.Name))

	if csv.Status.Phase != olmv1alpha1.CSVPhaseDeleting {

		// not deleting
		return []metav1.Condition{}
	}

	condition := meta.FindStatusCondition(conditions, "Progressing")
	if condition == nil {
		condition = &metav1.Condition{
			Type:               "Progressing",
			ObservedGeneration: csv.Generation,
			Status:             metav1.ConditionTrue,
			Reason:             "Deleting",
			Message:            fmt.Sprintf("phase [%s] ", csv.Status.Phase),
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}

	meta.SetStatusCondition(&conditions, *condition)

	return conditions
}
