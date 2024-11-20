package vmi

import (
	//	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtv1 "kubevirt.io/api/core/v1"
)

func RegisterOperation() {
	controllers.RegisterOperation(&migrationOperation{}, "LiveMigration", "virtualmachineinstances")
}

type migrationOperation struct {
}

func (m *migrationOperation) ProcessOperation(ctx context.Context, obj interface{}, conditions []metav1.Condition) []metav1.Condition {

	logger := klog.FromContext(ctx)

	vmi := obj.(*virtv1.VirtualMachineInstance)
	logger.Info(fmt.Sprintf("processing live migration operation for vmi [%s]", vmi.Name))

	if vmi.Status.MigrationState == nil || vmi.Status.MigrationState.EndTimestamp != nil {
		// return empty conditions when no migration is in progress
		// This signals no in-flight migration is taking place
		return []metav1.Condition{}
	}

	condition := meta.FindStatusCondition(conditions, "Progressing")
	if condition == nil {
		condition = &metav1.Condition{
			Type:               "Progressing",
			ObservedGeneration: vmi.Generation,
			Status:             metav1.ConditionTrue,
			Reason:             "LiveMigrationProgressing",
			Message:            fmt.Sprintf("Live migration is progressing to target node [%s] with target pod [%s]", vmi.Status.MigrationState.TargetNode, vmi.Status.MigrationState.TargetPod),
			LastTransitionTime: metav1.NewTime(time.Now()),
		}

	}

	meta.SetStatusCondition(&conditions, *condition)

	return conditions
}
