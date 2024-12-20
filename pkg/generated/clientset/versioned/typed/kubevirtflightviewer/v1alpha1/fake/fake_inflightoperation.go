/*
Copyright The Kubernetes Authors.

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

package fake

import (
	gentype "k8s.io/client-go/gentype"
	v1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/apis/kubevirtflightviewer/v1alpha1"
	kubevirtflightviewerv1alpha1 "k8s.io/kubevirt-flight-viewer/pkg/generated/clientset/versioned/typed/kubevirtflightviewer/v1alpha1"
)

// fakeInFlightOperations implements InFlightOperationInterface
type fakeInFlightOperations struct {
	*gentype.FakeClientWithList[*v1alpha1.InFlightOperation, *v1alpha1.InFlightOperationList]
	Fake *FakeKubevirtflightviewerV1alpha1
}

func newFakeInFlightOperations(fake *FakeKubevirtflightviewerV1alpha1, namespace string) kubevirtflightviewerv1alpha1.InFlightOperationInterface {
	return &fakeInFlightOperations{
		gentype.NewFakeClientWithList[*v1alpha1.InFlightOperation, *v1alpha1.InFlightOperationList](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("inflightoperations"),
			v1alpha1.SchemeGroupVersion.WithKind("InFlightOperation"),
			func() *v1alpha1.InFlightOperation { return &v1alpha1.InFlightOperation{} },
			func() *v1alpha1.InFlightOperationList { return &v1alpha1.InFlightOperationList{} },
			func(dst, src *v1alpha1.InFlightOperationList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.InFlightOperationList) []*v1alpha1.InFlightOperation {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.InFlightOperationList, items []*v1alpha1.InFlightOperation) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
