package vm

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	virtv1 "kubevirt.io/api/core/v1"
)

func RegisterOperation() {
	controllers.RegisterOperation(&startingOperation{}, "Starting", "virtualmachineinstances", virtv1.VirtualMachineInstanceGroupVersionKind)
	controllers.RegisterOperation(&stoppingOperation{}, "Stopping", "virtualmachineinstances", virtv1.VirtualMachineInstanceGroupVersionKind)
	controllers.RegisterOperation(&migratingOperation{}, "LiveMigrating", "virtualmachineinstances", virtv1.VirtualMachineInstanceGroupVersionKind)
}

type startingOperation struct {
}

func (m *startingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	logger := klog.FromContext(ctx)

	vmi := obj.(*virtv1.VirtualMachineInstance)

	if vmi.DeletionTimestamp != nil {
		// return empty state when no stopping is in progress
		// This signals no in-flight stopping is taking place
		return nil
	} else if vmi.Status.Phase != virtv1.VmPhaseUnset &&
		vmi.Status.Phase != virtv1.Pending &&
		vmi.Status.Phase != virtv1.Scheduling &&
		vmi.Status.Phase != virtv1.Scheduled {

		// return empty state when no stopping is in progress
		// This signals no in-flight stopping is taking place
		return nil
	}

	logger.Info(fmt.Sprintf("processing starting operation for vmi [%s]", vmi.Name))
	return &v1alpha1.InFlightOperationState{
		TransitionState: v1alpha1.TransitionStateProgressing,
		Reason:          "Starting",
		Message:         fmt.Sprintf("starting vm [%s]", vmi.Name),
	}
}

type stoppingOperation struct {
}

func (s *stoppingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {
	vmi := obj.(*virtv1.VirtualMachineInstance)

	// return empty state when no stopping is in progress
	// This signals no in-flight stopping is taking place
	if vmi.DeletionTimestamp == nil {
		return nil
	} else if vmi.Status.Phase == virtv1.Succeeded || vmi.Status.Phase == virtv1.Failed {
		return nil
	}

	logger := klog.FromContext(ctx)
	logger.Info(fmt.Sprintf("processing stopping operation for vmi [%s]", vmi.Name))

	return &v1alpha1.InFlightOperationState{
		TransitionState: v1alpha1.TransitionStateProgressing,
		Reason:          "Stopping",
		Message:         fmt.Sprintf("terminating vm [%s]", vmi.Name),
	}

}

type migratingOperation struct {
}

func (m *migratingOperation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {
	vmi := obj.(*virtv1.VirtualMachineInstance)
	if vmi.Status.MigrationState == nil {
		return nil
	} else if vmi.Status.MigrationState.EndTimestamp != nil {
		return nil
	}

	logger := klog.FromContext(ctx)
	logger.Info(fmt.Sprintf("processing live migrating operation for vmi [%s]", vmi.Name))
	return &v1alpha1.InFlightOperationState{
		TransitionState: v1alpha1.TransitionStateProgressing,
		Reason:          "LiveMigrating",
		Message:         fmt.Sprintf("live migrating vm [%s] ", vmi.Name),
	}
}
