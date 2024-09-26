package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*Agent)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec,omitempty"`
	Status AgentStatus `json:"status,omitempty"`
}

func (in *Agent) AgentName() string {
	return in.Name
}

func (in *Agent) WorkflowName() string {
	return ""
}

func (in *Agent) ThreadName() string {
	return ""
}

func (in *Agent) KnowledgeWorkspaceStatus() *KnowledgeWorkspaceStatus {
	return &in.Status.KnowledgeWorkspace
}

func (in *Agent) WorkspaceStatus() *WorkspaceStatus {
	return &in.Status.Workspace
}

func (in *Agent) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type AgentSpec struct {
	Manifest            AgentManifest `json:"manifest,omitempty"`
	InputFilters        []string      `json:"inputFilters,omitempty"`
	CredentialContextID string        `json:"credentialContextID,omitempty"`
}

type AgentStatus struct {
	Conditions         []metav1.Condition       `json:"conditions,omitempty"`
	External           AgentExternalStatus      `json:"external,omitempty"`
	Workspace          WorkspaceStatus          `json:"workspace,omitempty"`
	KnowledgeWorkspace KnowledgeWorkspaceStatus `json:"knowledgeWorkspace,omitempty"`
}

type WorkspaceStatus struct {
	WorkspaceID string `json:"workspaceID,omitempty"`
}

// +k8s:deepcopy-gen=false

type Knowledgeable interface {
	GetKnowledgeWorkspaceStatus() *KnowledgeWorkspaceStatus
}

type KnowledgeWorkspaceStatus struct {
	HasKnowledge                bool   `json:"hasKnowledge,omitempty"`
	KnowledgeGeneration         int64  `json:"knowledgeGeneration,omitempty"`
	ObservedKnowledgeGeneration int64  `json:"observedKnowledgeGeneration,omitempty"`
	KnowledgeWorkspaceID        string `json:"knowledgeWorkspaceID,omitempty"`
	IngestionRunName            string `json:"ingestionRunName,omitempty"`
}

type AgentExternalStatus struct {
	RefNameAssigned bool `json:"refNameAssigned,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Agent `json:"items"`
}
