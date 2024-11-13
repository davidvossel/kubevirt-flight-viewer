package controllers

import (
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	restclient "k8s.io/client-go/rest"
	clientset "k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned"
	informers "k8s.io/kubevirt-flight-viewer/pkg/generated/informers/externalversions"
)

/*
	type InFlightOperationRegistration interface {
		ProcessOperation(context.Context, []*flightviewerv1alpha1.InFlightOperation)
	}

	func Register(registration InFlightOperationRegistration, operationName string, resource string) error {
		return nil
	}
*/
func Bootstrap(ctx context.Context, cfg *restclient.Config) {
	logger := klog.FromContext(ctx)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	kvViewerClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	//kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	kvViewerInformerFactory := informers.NewSharedInformerFactory(kvViewerClient, time.Second*30)

	controller := NewController(ctx, kubeClient, kvViewerClient,
		kvViewerInformerFactory.Kubevirtflightviewer().V1alpha1().InFlightOperations())

	//kubeInformerFactory.Start(ctx.Done())
	kvViewerInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, 2); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
