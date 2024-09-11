package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*Slug)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Slug struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlugSpec   `json:"spec,omitempty"`
	Status SlugStatus `json:"status,omitempty"`
}

func (in *Slug) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type SlugSpec struct {
	AgentName    string `json:"agentName,omitempty"`
	WorkflowName string `json:"workflowName,omitempty"`
}

type SlugStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SlugList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Slug `json:"items"`
}
