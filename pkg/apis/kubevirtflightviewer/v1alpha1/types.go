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
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InFlightClusterOperation is a specification for a InFlightClusterOperation resource
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=inflightclusteroperations,shortName=ifco;ifcos,scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Operation_Type",type="string",JSONPath=".status.operationType",description="Operation Type"
// +kubebuilder:printcolumn:name="Resource_Kind",type="string",JSONPath=".status.resourceReference.kind",description="Resource Kind"
// +kubebuilder:printcolumn:name="Resource_Name",type="string",JSONPath=".status.resourceReference.name",description="Resource Name"
// +kubebuilder:printcolumn:name="Resource_Namespace",type="string",JSONPath=".status.resourceReference.namespace",description="Resource Namespace"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.operationState.message",description="Message"
type InFlightClusterOperation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//	Spec   InFlightOperationSpec   `json:"spec"`
	Status InFlightOperationStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InFlightClusterOperationList is a list of InFlightClusterOperation resources
type InFlightClusterOperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []InFlightClusterOperation `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InFlightOperation is a specification for a InFlightOperation resource
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=inflightoperations,shortName=ifo;ifos,scope=Namespaced
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Operation_Type",type="string",JSONPath=".status.operationType",description="Operation Type"
// +kubebuilder:printcolumn:name="Resource_Kind",type="string",JSONPath=".status.resourceReference.kind",description="Resource Kind"
// +kubebuilder:printcolumn:name="Resource_Name",type="string",JSONPath=".status.resourceReference.name",description="Resource Name"
// +kubebuilder:printcolumn:name="Resource_Namespace",type="string",JSONPath=".status.resourceReference.namespace",description="Resource Namespace"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.operationState.message",description="Message"
type InFlightOperation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//	Spec   InFlightOperationSpec   `json:"spec"`
	Status InFlightOperationStatus `json:"status"`
}

type InFlightOperationResourceReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion"`
	// Kind of the referent.
	Kind string `json:"kind"`
	// Name of the referent.
	Name string `json:"name"`
	// Namespace of the referent. Optional for cluster scoped resources
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// UID of the referent.
	UID types.UID `json:"uid"`
}

type InFlightOperationTransitionState string

const (
	TransitionStateProgressing InFlightOperationTransitionState = "Progressing"
	TransitionStateBlocked     InFlightOperationTransitionState = "Blocked"
	TransitionStateSucceeded   InFlightOperationTransitionState = "Succeeded"
	TransitionStateFailed      InFlightOperationTransitionState = "Failed"
)

type InFlightOperationState struct {
	TransitionState InFlightOperationTransitionState `json:"transitionState"`

	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// InFlightOperationStatus is the status for a InFlightOperation resource
type InFlightOperationStatus struct {
	// OperationType of operation
	OperationType string `json:"operationType"`

	// OperationState reflects what is currently happening for an inflight operation
	OperationState *InFlightOperationState `json:"operationState"`

	// ResourceReference is the resource this operation is related to
	ResourceReference *InFlightOperationResourceReference `json:"resourceReference"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InFlightOperationList is a list of InFlightOperation resources
type InFlightOperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []InFlightOperation `json:"items"`
}
