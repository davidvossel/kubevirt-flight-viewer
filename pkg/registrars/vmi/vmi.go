package vmi

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"

	"k8s.io/klog/v2"
	flightviewerv1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	virtv1 "kubevirt.io/api/core/v1"
)

func RegisterOperation() {
	controllers.RegisterOperation(&migrationOperation{}, "LiveMigration", "virtualmachineinstances")
}

type migrationOperation struct {
}

func (m *migrationOperation) ProcessOperation(ctx context.Context, obj interface{}, op *flightviewerv1alpha1.InFlightOperationSpec) *flightviewerv1alpha1.InFlightOperationSpec {

	logger := klog.FromContext(ctx)

	vmi := obj.(*virtv1.VirtualMachineInstance)
	logger.Info(fmt.Sprintf("processing live migration operation for vmi [%s]", vmi.Name))

	if op != nil {
		op = &flightviewerv1alpha1.InFlightOperationSpec{}
	}

	return op
}
