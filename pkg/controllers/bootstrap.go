package controllers

import (
	"context"
	"time"

	// cacheinformers"k8s.io/client-go/informers"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	informers "k8s.io/kubevirt-flight-viewer/pkg/generated/informers/externalversions"
	kubev1 "kubevirt.io/api/core/v1"

	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	clientset "k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned"
)

const defaultResync = time.Second * 30

func init() {
}

func kvRestClient(cfg *restclient.Config) (*restclient.RESTClient, error) {

	schemeBuilder := runtime.NewSchemeBuilder(kubev1.AddKnownTypesGenerator(kubev1.GroupVersions))
	scheme := runtime.NewScheme()
	addToScheme := schemeBuilder.AddToScheme
	codecs := serializer.NewCodecFactory(scheme)
	//parameterCodec := runtime.NewParameterCodec(Scheme)
	addToScheme(scheme)
	addToScheme(k8sscheme.Scheme)

	shallowCopy := *cfg
	shallowCopy.GroupVersion = &kubev1.StorageGroupVersion
	shallowCopy.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: codecs}
	shallowCopy.APIPath = "/apis"
	shallowCopy.ContentType = runtime.ContentTypeJSON
	if cfg.UserAgent == "" {
		cfg.UserAgent = restclient.DefaultKubernetesUserAgent()
	}

	return restclient.RESTClientFor(&shallowCopy)
}

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

	resourceInformers := map[string]cache.SharedIndexInformer{}

	// VMI informer

	restClient, err := kvRestClient(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes rest client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	lw := cache.NewListWatchFromClient(restClient, "virtualmachineinstances", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["virtualmachineinstances"] = cache.NewSharedIndexInformer(lw, &kubev1.VirtualMachineInstance{}, defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	//kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	kvViewerInformerFactory := informers.NewSharedInformerFactory(kvViewerClient, defaultResync)

	controller, err := NewController(ctx, kubeClient, kvViewerClient,
		kvViewerInformerFactory.Kubevirtflightviewer().V1alpha1().InFlightOperations(),
		resourceInformers,
	)
	if err != nil {
		logger.Error(err, "Error building controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	kvViewerInformerFactory.Start(ctx.Done())
	for _, informer := range resourceInformers {
		go informer.Run(ctx.Done())
	}

	if err = controller.Run(ctx, 2); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
