package vmi

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"

	flightviewerv1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
)

func RegisterOperation() {
	controllers.RegisterOperation(&migrationOperation{}, "LiveMigration", "virtualmachineinstances")
}

type migrationOperation struct {
}

func (m *migrationOperation) ProcessOperation(ctx context.Context, obj interface{}, curOps []*flightviewerv1alpha1.InFlightOperation) {

	//logger := klog.FromContext(ctx)
	//logger.Info("processing vmi for live migration")
}
