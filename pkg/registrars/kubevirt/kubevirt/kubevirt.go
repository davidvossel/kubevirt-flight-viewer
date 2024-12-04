package kubevirt

import (
	"context"

	"k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	"k8s.io/kubevirt-flight-viewer/pkg/controllers"

	virtv1 "kubevirt.io/api/core/v1"
)

func RegisterOperation() {
	controllers.RegisterOperation(&operation{}, "Updating", "kubevirts", virtv1.KubeVirtGroupVersionKind)
}

type operation struct {
}

func (o *operation) ProcessOperation(ctx context.Context, obj interface{}) *v1alpha1.InFlightOperationState {

	//logger := klog.FromContext(ctx)

	//kv := obj.(*virtv1.KubeVirt)

	// TODO implement this
	//	logger.Info(fmt.Sprintf("processing starting operation for kv [%s]", kv.Name))

	return nil
}
