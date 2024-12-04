package csv

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
)

func RegisterOperation() {

	controllers.RegisterOperation(&replacingOperation{}, "Replacing", "clusterserviceversions", schema.GroupVersionKind{
		Version: olmv1alpha1.SchemeGroupVersion.Version,
		Group:   olmv1alpha1.SchemeGroupVersion.Group,
		Kind:    "ClusterServiceVersion",
	})

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

func (o *installingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	csv := obj.(*olmv1alpha1.ClusterServiceVersion)

	switch csv.Status.Phase {
	case olmv1alpha1.CSVPhaseSucceeded, olmv1alpha1.CSVPhaseFailed, olmv1alpha1.CSVPhaseDeleting, olmv1alpha1.CSVPhaseUnknown, olmv1alpha1.CSVPhaseReplacing:
		// not installing if in any of these phases
		return nil
	}

	logger := klog.FromContext(ctx)
	logger.Info(fmt.Sprintf("processing installing operation for csv [%s]", csv.Name))
	return &v1alpha1.InFlightOperationState{
		TransitionState: v1alpha1.TransitionStateProgressing,
		Reason:          "Installing",
		Message:         fmt.Sprintf("phase [%s]: %s", string(csv.Status.Phase), csv.Status.Message),
	}

}

type deletingOperation struct {
}

func (o *deletingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	csv := obj.(*olmv1alpha1.ClusterServiceVersion)

	if csv.Status.Phase != olmv1alpha1.CSVPhaseDeleting {
		// not deleting
		return nil
	}

	logger := klog.FromContext(ctx)
	logger.Info(fmt.Sprintf("processing deleting operation for csv [%s]", csv.Name))

	return &v1alpha1.InFlightOperationState{
		TransitionState: v1alpha1.TransitionStateProgressing,
		Reason:          "Deleting",
		Message:         fmt.Sprintf("phase [%s]: %s", csv.Status.Phase, csv.Status.Message),
	}
}

type replacingOperation struct {
}

func (o *replacingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	csv := obj.(*olmv1alpha1.ClusterServiceVersion)

	if csv.Status.Phase != olmv1alpha1.CSVPhaseReplacing {
		// not replacing
		return nil
	}

	logger := klog.FromContext(ctx)
	logger.Info(fmt.Sprintf("processing replacing operation for csv [%s]", csv.Name))
	return &v1alpha1.InFlightOperationState{
		TransitionState: v1alpha1.TransitionStateProgressing,
		Reason:          "Replacing",
		Message:         fmt.Sprintf("phase [%s]: %s", csv.Status.Phase, csv.Status.Message),
	}
}
