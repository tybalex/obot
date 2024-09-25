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

func (in *Thread) DeleteRefs() []Ref {
	return []Ref{}
}

type ThreadManifest struct {
	Tools       []string `json:"tools,omitempty"`
	Description string   `json:"description,omitempty"`
}

type ThreadStatus struct {
	LastRunName        string                   `json:"lastRunName,omitempty"`
	LastRunState       gptscriptclient.RunState `json:"lastRunState,omitempty"`
	PreviousThreadName string                   `json:"previousThreadName,omitempty"`
	Conditions         []metav1.Condition       `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ThreadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Thread `json:"items"`
}
