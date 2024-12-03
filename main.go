/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kvupdating "k8s.io/kubevirt-flight-viewer/pkg/registrars/kubevirt/kubevirt/updating"
	"k8s.io/kubevirt-flight-viewer/pkg/registrars/kubevirt/vm/starting"
	"k8s.io/kubevirt-flight-viewer/pkg/registrars/kubevirt/vm/stopping"
	"k8s.io/kubevirt-flight-viewer/pkg/registrars/livemigration"
	ocpmachinestarting "k8s.io/kubevirt-flight-viewer/pkg/registrars/ocpmachine/starting"
	ocpmachineconfigupdating "k8s.io/kubevirt-flight-viewer/pkg/registrars/ocpmachineconfig/updating"
	csvinstalling "k8s.io/kubevirt-flight-viewer/pkg/registrars/olm/csv/installing"
	"k8s.io/kubevirt-flight-viewer/pkg/signals"

	"k8s.io/kubevirt-flight-viewer/pkg/controllers"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the shutdown signal gracefully
	ctx := signals.SetupSignalHandler()
	logger := klog.FromContext(ctx)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logger.Error(err, "Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	controllers.Bootstrap(ctx, cfg)
}

func init() {
	// register operations
	livemigration.RegisterOperation()
	stopping.RegisterOperation()
	starting.RegisterOperation()
	ocpmachinestarting.RegisterOperation()
	kvupdating.RegisterOperation()
	ocpmachineconfigupdating.RegisterOperation()
	csvinstalling.RegisterOperation()

	// register flags
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
