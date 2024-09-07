package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	gptscriptclient "github.com/gptscript-ai/go-gptscript"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*Run)(nil)
)

const (
	RunFinalizer = "otto.gptscript.ai/run"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec   `json:"spec,omitempty"`
	Status RunStatus `json:"status,omitempty"`
}

func (in *Run) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type Progress struct {
	Content string       `json:"content"`
	Tool    ToolProgress `json:"tool"`
}

type ToolProgress struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Input       string `json:"input,omitempty"`
}

type RunSpec struct {
	ThreadName      string `json:"threadName,omitempty"`
	AgentName       string `json:"agentName,omitempty"`
	PreviousRunName string `json:"previousRunName,omitempty"`
	Input           string `json:"input"`
}

type RunStatus struct {
	Conditions []metav1.Condition       `json:"conditions,omitempty"`
	State      gptscriptclient.RunState `json:"state,omitempty"`
	Output     string                   `json:"output"`
	Error      string                   `json:"error,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Run `json:"items"`
}
