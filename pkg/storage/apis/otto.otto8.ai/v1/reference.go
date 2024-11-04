package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Reference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReferenceSpec `json:"spec,omitempty"`
	Status EmptyStatus   `json:"status,omitempty"`
}

type ReferenceSpec struct {
	AgentName    string `json:"agentName,omitempty"`
	WorkflowName string `json:"workflowName,omitempty"`
}

func (in *Reference) DeleteRefs() []Ref {
	return []Ref{}
}

type EmptyStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Reference `json:"items"`
}
