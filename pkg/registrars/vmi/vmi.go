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

func (m *migrationOperation) ProcessOperation(ctx context.Context, obj interface{}, curOps []flightviewerv1alpha1.InFlightOperation) []flightviewerv1alpha1.InFlightOperation {

	logger := klog.FromContext(ctx)

	vmi := obj.(*virtv1.VirtualMachineInstance)
	logger.Info(fmt.Sprintf("processing live migration operation for vmi [%s]", vmi.Name))

	var cur flightviewerv1alpha1.InFlightOperation

	if len(curOps) == 0 {
		cur = flightviewerv1alpha1.InFlightOperation{}
	} else {
		cur = curOps[0]
	}

	return []flightviewerv1alpha1.InFlightOperation{cur}
}
