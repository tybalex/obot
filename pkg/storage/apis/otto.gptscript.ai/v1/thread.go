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
	Input     string `json:"input,omitempty"`
	AgentName string `json:"agentName,omitempty"`
	Script    []byte `json:"script,omitempty"`
}

type ThreadStatus struct {
	Description   string                   `json:"description,omitempty"`
	LastRunName   string                   `json:"lastRunName,omitempty"`
	LastRunState  gptscriptclient.RunState `json:"lastRunState,omitempty"`
	LastRunOutput string                   `json:"lastRunOutput,omitempty"`
	LastRunError  string                   `json:"lastRunError,omitempty"`
	Conditions    []metav1.Condition       `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ThreadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Thread `json:"items"`
}
