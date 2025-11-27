// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=appwrapper

// AppWrapper is the Schema for the appwrappers API
type AppWrapper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppWrapperSpec   `json:"spec,omitempty"`
	Status AppWrapperStatus `json:"status,omitempty"`
}

// AppWrapperSpec defines the desired state of AppWrapper
type AppWrapperSpec struct {
	// Components is the list of components that make up the AppWrapper
	Components []AppWrapperComponent `json:"components"`

	// Suspend indicates whether the AppWrapper is suspended
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// ManagedBy indicates the controller managing this AppWrapper
	// +optional
	ManagedBy *string `json:"managedBy,omitempty"`
}

// AppWrapperComponent defines a component within the AppWrapper
type AppWrapperComponent struct {
	// Annotations for this component
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// DeclaredPodSets defines the pod sets for this component
	// +optional
	DeclaredPodSets []AppWrapperPodSet `json:"podSets,omitempty"`

	// PodSetInfos contains injected pod set information from Kueue
	// +optional
	PodSetInfos []AppWrapperPodSetInfo `json:"podSetInfos,omitempty"`

	// Template is the raw Kubernetes resource template
	Template runtime.RawExtension `json:"template"`
}

// AppWrapperPodSet defines a pod set within a component
type AppWrapperPodSet struct {
	// Replicas is the number of pod replicas
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Path is the JSONPath to the PodTemplateSpec in the component template
	Path string `json:"path"`

	// Annotations for this pod set
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// AppWrapperPodSetInfo contains pod set information injected by Kueue
type AppWrapperPodSetInfo struct {
	// NodeSelector to be merged into the pod template
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations to be merged into the pod template
	// +optional
	Tolerations []string `json:"tolerations,omitempty"`

	// Annotations to be merged into the pod template
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels to be merged into the pod template
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// AppWrapperStatus defines the observed state of AppWrapper
type AppWrapperStatus struct {
	// Phase represents the current phase of the AppWrapper
	Phase AppWrapperPhase `json:"phase,omitempty"`

	// Retries is the number of times the AppWrapper has been reset
	Retries int32 `json:"resettingCount,omitempty"`

	// Conditions represent the latest available observations of the AppWrapper's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ComponentStatus contains status for each component
	// +optional
	ComponentStatus []AppWrapperComponentStatus `json:"componentStatus,omitempty"`
}

// AppWrapperComponentStatus defines the status of a component
type AppWrapperComponentStatus struct {
	// Name of the component
	Name string `json:"name,omitempty"`

	// APIVersion of the component
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the component
	Kind string `json:"kind,omitempty"`

	// Status conditions of the component
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// AppWrapperPhase represents the phase of an AppWrapper
type AppWrapperPhase string

const (
	// AppWrapperEmpty is the initial phase
	AppWrapperEmpty AppWrapperPhase = ""
	// AppWrapperSuspended means the AppWrapper is suspended and components are not deployed
	AppWrapperSuspended AppWrapperPhase = "Suspended"
	// AppWrapperResuming means the AppWrapper is deploying components
	AppWrapperResuming AppWrapperPhase = "Resuming"
	// AppWrapperRunning means all components are deployed and running
	AppWrapperRunning AppWrapperPhase = "Running"
	// AppWrapperResetting means the AppWrapper is deleting components for a retry
	AppWrapperResetting AppWrapperPhase = "Resetting"
	// AppWrapperSuspending means the AppWrapper is undeploying components
	AppWrapperSuspending AppWrapperPhase = "Suspending"
	// AppWrapperSucceeded means the workload completed successfully
	AppWrapperSucceeded AppWrapperPhase = "Succeeded"
	// AppWrapperFailed means the workload failed
	AppWrapperFailed AppWrapperPhase = "Failed"
	// AppWrapperTerminating means the AppWrapper is being deleted
	AppWrapperTerminating AppWrapperPhase = "Terminating"
)

// AppWrapper condition types
const (
	// AppWrapperConditionQuotaReserved indicates Kueue has reserved quota
	AppWrapperConditionQuotaReserved = "QuotaReserved"
	// AppWrapperConditionResourcesDeployed indicates components are deployed
	AppWrapperConditionResourcesDeployed = "ResourcesDeployed"
	// AppWrapperConditionPodsReady indicates all pods are ready
	AppWrapperConditionPodsReady = "PodsReady"
	// AppWrapperConditionUnhealthy indicates an unhealthy state was detected
	AppWrapperConditionUnhealthy = "Unhealthy"
	// AppWrapperConditionDeletingResources indicates resources are being deleted
	AppWrapperConditionDeletingResources = "DeletingResources"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=appwrappers

// AppWrapperList contains a list of AppWrapper
type AppWrapperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppWrapper `json:"items"`
}
