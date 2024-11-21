/*
Copyright 2018 The CDI Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	"fmt"
	"net/http"

	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
	cdiv1alpha1 "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned/typed/core/v1alpha1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned/typed/core/v1beta1"
	forkliftv1beta1 "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned/typed/forklift/v1beta1"
	uploadv1beta1 "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned/typed/upload/v1beta1"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	CdiV1alpha1() cdiv1alpha1.CdiV1alpha1Interface
	CdiV1beta1() cdiv1beta1.CdiV1beta1Interface
	ForkliftV1beta1() forkliftv1beta1.ForkliftV1beta1Interface
	UploadV1beta1() uploadv1beta1.UploadV1beta1Interface
}

// Clientset contains the clients for groups.
type Clientset struct {
	*discovery.DiscoveryClient
	cdiV1alpha1     *cdiv1alpha1.CdiV1alpha1Client
	cdiV1beta1      *cdiv1beta1.CdiV1beta1Client
	forkliftV1beta1 *forkliftv1beta1.ForkliftV1beta1Client
	uploadV1beta1   *uploadv1beta1.UploadV1beta1Client
}

// CdiV1alpha1 retrieves the CdiV1alpha1Client
func (c *Clientset) CdiV1alpha1() cdiv1alpha1.CdiV1alpha1Interface {
	return c.cdiV1alpha1
}

// CdiV1beta1 retrieves the CdiV1beta1Client
func (c *Clientset) CdiV1beta1() cdiv1beta1.CdiV1beta1Interface {
	return c.cdiV1beta1
}

// ForkliftV1beta1 retrieves the ForkliftV1beta1Client
func (c *Clientset) ForkliftV1beta1() forkliftv1beta1.ForkliftV1beta1Interface {
	return c.forkliftV1beta1
}

// UploadV1beta1 retrieves the UploadV1beta1Client
func (c *Clientset) UploadV1beta1() uploadv1beta1.UploadV1beta1Interface {
	return c.uploadV1beta1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	// share the transport between all clients
	httpClient, err := rest.HTTPClientFor(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return NewForConfigAndClient(&configShallowCopy, httpClient)
}

// NewForConfigAndClient creates a new Clientset for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfigAndClient will generate a rate-limiter in configShallowCopy.
func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error
	cs.cdiV1alpha1, err = cdiv1alpha1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.cdiV1beta1, err = cdiv1beta1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.forkliftV1beta1, err = forkliftv1beta1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	cs.uploadV1beta1, err = uploadv1beta1.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	cs, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.cdiV1alpha1 = cdiv1alpha1.New(c)
	cs.cdiV1beta1 = cdiv1beta1.New(c)
	cs.forkliftV1beta1 = forkliftv1beta1.New(c)
	cs.uploadV1beta1 = uploadv1beta1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}