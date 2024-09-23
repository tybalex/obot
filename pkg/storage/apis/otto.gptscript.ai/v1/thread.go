package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	gptscriptclient "github.com/gptscript-ai/go-gptscript"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*Thread)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Thread struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ThreadSpec   `json:"spec,omitempty"`
	Status ThreadStatus `json:"status,omitempty"`
}

func (in *Thread) AgentName() string {
	return in.Spec.AgentName
}

func (in *Thread) WorkflowName() string {
	return in.Spec.WorkflowName
}

func (in *Thread) ThreadName() string {
	return in.Name
}

func (in *Thread) KnowledgeWorkspaceStatus() *KnowledgeWorkspaceStatus {
	return &in.Status.KnowledgeWorkspace
}

func (in *Thread) WorkspaceStatus() *WorkspaceStatus {
	return &in.Status.Workspace
}

func (in *Thread) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type ThreadSpec struct {
	Manifest              ThreadManifest `json:"manifest,omitempty"`
	AgentName             string         `json:"agentName,omitempty"`
	WorkflowName          string         `json:"workflowName,omitempty"`
	WorkflowExecutionName string         `json:"workflowExecutionName,omitempty"`
	WorkflowStepName      string         `json:"workflowStepName,omitempty"`
	WorkspaceID           string         `json:"workspaceID,omitempty"`
	KnowledgeWorkspaceID  string         `json:"knowledgeWorkspaceID,omitempty"`
}

type ThreadManifest struct {
	Tools       []string `json:"tools,omitempty"`
	Description string   `json:"description,omitempty"`
}

type ThreadStatus struct {
	LastRunName        string                   `json:"lastRunName,omitempty"`
	LastRunState       gptscriptclient.RunState `json:"lastRunState,omitempty"`
	Conditions         []metav1.Condition       `json:"conditions,omitempty"`
	Workspace          WorkspaceStatus          `json:"workspace,omitempty"`
	KnowledgeWorkspace KnowledgeWorkspaceStatus `json:"knowledgeWorkspace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ThreadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Thread `json:"items"`
}
