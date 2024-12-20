package controllers

import (
	"context"
	"time"

	// cacheinformers"k8s.io/client-go/informers"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	informers "k8s.io/kubevirt-flight-viewer/pkg/generated/informers/externalversions"
	kubev1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	machineconfigv1 "github.com/openshift/api/machineconfiguration/v1"
	ocpmachineclient "github.com/openshift/client-go/machine/clientset/versioned"
	ocpmachineconfigclient "github.com/openshift/client-go/machineconfiguration/clientset/versioned"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	clientset "k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned"
	cdiclient "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned"
)

const defaultResync = time.Second * 30

func cdiRestClient(cfg *restclient.Config) (restclient.Interface, error) {
	shallowCopy := *cfg

	cdiClient, err := cdiclient.NewForConfig(&shallowCopy)
	if err != nil {
		return nil, err
	}

	return cdiClient.CdiV1beta1().RESTClient(), nil
}

func ocpMachineConfigRestClient(cfg *restclient.Config) (restclient.Interface, error) {
	shallowCopy := *cfg

	ocpClient, err := ocpmachineconfigclient.NewForConfig(&shallowCopy)
	if err != nil {
		return nil, err
	}

	return ocpClient.MachineconfigurationV1().RESTClient(), nil
}

func ocpMachineRestClient(cfg *restclient.Config) (restclient.Interface, error) {
	shallowCopy := *cfg

	ocpClient, err := ocpmachineclient.NewForConfig(&shallowCopy)
	if err != nil {
		return nil, err
	}

	return ocpClient.MachineV1beta1().RESTClient(), nil
}

func olmRestClient(cfg *restclient.Config) (*restclient.RESTClient, error) {

	knownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(olmv1alpha1.SchemeGroupVersion,
			&olmv1alpha1.ClusterServiceVersion{},
			&olmv1alpha1.ClusterServiceVersionList{},
		)
		metav1.AddToGroupVersion(scheme, olmv1alpha1.SchemeGroupVersion)
		return nil
	}

	schemeBuilder := runtime.NewSchemeBuilder(knownTypes)

	scheme := runtime.NewScheme()
	addToScheme := schemeBuilder.AddToScheme
	codecs := serializer.NewCodecFactory(scheme)
	//parameterCodec := runtime.NewParameterCodec(Scheme)
	addToScheme(scheme)
	addToScheme(k8sscheme.Scheme)

	shallowCopy := *cfg
	shallowCopy.GroupVersion = &olmv1alpha1.SchemeGroupVersion
	shallowCopy.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: codecs}
	shallowCopy.APIPath = "/apis"
	shallowCopy.ContentType = runtime.ContentTypeJSON
	if cfg.UserAgent == "" {
		cfg.UserAgent = restclient.DefaultKubernetesUserAgent()
	}

	return restclient.RESTClientFor(&shallowCopy)
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

	// KubeVirt informers
	kvRC, err := kvRestClient(cfg)
	if err != nil {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	lw := cache.NewListWatchFromClient(kvRC, "virtualmachineinstances", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["virtualmachineinstances"] = cache.NewSharedIndexInformer(lw, &kubev1.VirtualMachineInstance{}, defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	lw = cache.NewListWatchFromClient(kvRC, "kubevirts", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["kubevirts"] = cache.NewSharedIndexInformer(lw, &kubev1.KubeVirt{}, defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	// CDI Informers
	cdiRC, err := cdiRestClient(cfg)
	if err != nil {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	// DataVolume Informer
	lw = cache.NewListWatchFromClient(cdiRC, "datavolumes", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["datavolumes"] = cache.NewSharedIndexInformer(lw, &cdiv1.DataVolume{}, defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	// OCP Machine Informers
	machineRC, err := ocpMachineRestClient(cfg)
	if err != nil {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	// Machine Informer
	lw = cache.NewListWatchFromClient(machineRC, "machines", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["machines"] = cache.NewSharedIndexInformer(lw, &machinev1beta1.Machine{}, defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	// MachineSet Informer
	lw = cache.NewListWatchFromClient(machineRC, "machinesets", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["machinesets"] = cache.NewSharedIndexInformer(lw, &machinev1beta1.MachineSet{}, defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	// OCP Machine Config Informers
	machineConfigRC, err := ocpMachineConfigRestClient(cfg)
	if err != nil {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// MachineConfigPool Informer
	lw = cache.NewListWatchFromClient(machineConfigRC, "machineconfigpools", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["machineconfigpools"] = cache.NewSharedIndexInformer(lw, &machineconfigv1.MachineConfigPool{}, defaultResync, cache.Indexers{})

	// OLM Informers
	olmRC, err := olmRestClient(cfg)
	if err != nil {
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	// Register OLM types to the client
	//olmoperators.Install(olmclient.Scheme())

	lw = cache.NewListWatchFromClient(olmRC, "clusterserviceversions", k8sv1.NamespaceAll, fields.Everything())
	resourceInformers["clusterserviceversions"] = cache.NewSharedIndexInformer(lw, &olmv1alpha1.ClusterServiceVersion{}, defaultResync, cache.Indexers{})

	kvViewerInformerFactory := informers.NewSharedInformerFactory(kvViewerClient, defaultResync)
	controller, err := NewController(ctx, kubeClient, kvViewerClient,
		kvViewerInformerFactory.Kubevirtflightviewer().V1alpha1().InFlightOperations(),
		kvViewerInformerFactory.Kubevirtflightviewer().V1alpha1().InFlightClusterOperations(),
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
