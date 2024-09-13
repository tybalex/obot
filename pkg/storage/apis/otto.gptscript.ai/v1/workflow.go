package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*Workflow)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

func (in *Workflow) GetKnowledgeWorkspaceStatus() *KnowledgeWorkspaceStatus {
	return &in.Status.KnowledgeWorkspace
}

func (in *Workflow) GetWorkspaceStatus() *WorkspaceStatus {
	return &in.Status.Workspace
}

func (in *Workflow) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type WorkflowSpec struct {
	Manifest WorkflowManifest `json:"manifest,omitempty"`
}

type WorkflowExternalStatus struct {
	SlugAssigned bool `json:"slugAssigned,omitempty"`
}

type WorkflowStatus struct {
	External           WorkflowExternalStatus   `json:"external,omitempty"`
	Workspace          WorkspaceStatus          `json:"workspace,omitempty"`
	KnowledgeWorkspace KnowledgeWorkspaceStatus `json:"knowledgeWorkspace,omitempty"`
	Conditions         []metav1.Condition       `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Workflow `json:"items"`
}
