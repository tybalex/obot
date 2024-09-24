package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*Reference)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Reference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReferenceSpec   `json:"spec,omitempty"`
	Status ReferenceStatus `json:"status,omitempty"`
}

func (in *Reference) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type ReferenceSpec struct {
	AgentName    string `json:"agentName,omitempty"`
	WorkflowName string `json:"workflowName,omitempty"`
}

func (in *Reference) DeleteRefs() []Ref {
	return []Ref{}
}

type ReferenceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Reference `json:"items"`
}
