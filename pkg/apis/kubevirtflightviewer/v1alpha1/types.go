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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InFlightOperation is a specification for a InFlightOperation resource
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=inflightoperations,shortName=ifo;icoss,scope=Namespaced
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Operation_Type",type="string",JSONPath=".status.operationType",description="Operation Type"
// +kubebuilder:printcolumn:name="Resource_Kind",type="string",JSONPath=".metadata.ownerReferences[0].kind",description="Resource Kind"
// +kubebuilder:printcolumn:name="Resource_Name",type="string",JSONPath=".metadata.ownerReferences[0].name",description="Resource Name"
type InFlightOperation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InFlightOperationSpec   `json:"spec"`
	Status InFlightOperationStatus `json:"status"`
}

// InFlightOperationSpec is the spec for a InFlightOperation resource
type InFlightOperationSpec struct {
}

type InFlightResourceReference struct {
	// API version of the referent.
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,5,opt,name=apiVersion"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
}

// InFlightOperationStatus is the status for a InFlightOperation resource
type InFlightOperationStatus struct {
	// OperationType of operation
	OperationType string `json:"operationType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InFlightOperationList is a list of InFlightOperation resources
type InFlightOperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []InFlightOperation `json:"items"`
}
