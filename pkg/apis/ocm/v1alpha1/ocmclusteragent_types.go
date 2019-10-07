package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OCMClusterAgentSpec defines the desired state of OCMClusterAgent
// +k8s:openapi-gen=true
type OCMClusterAgentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// OCMClusterAgentStatus defines the observed state of OCMClusterAgent
// +k8s:openapi-gen=true
type OCMClusterAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OCMClusterAgent is the Schema for the ocmclusteragents API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type OCMClusterAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OCMClusterAgentSpec   `json:"spec,omitempty"`
	Status OCMClusterAgentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OCMClusterAgentList contains a list of OCMClusterAgent
type OCMClusterAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OCMClusterAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OCMClusterAgent{}, &OCMClusterAgentList{})
}
