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
	AgentWorkspaceFinalizer    = "otto.gptscript.ai/agent-workspace"
	AgentKnowledgeFinalizer    = "otto.gptscript.ai/agent-knowledge"
	RunFinalizer               = "otto.gptscript.ai/run"
	ThreadWorkspaceFinalizer   = "otto.gptscript.ai/thread-workspace"
	ThreadKnowledgeFinalizer   = "otto.gptscript.ai/thread-knowledge"
	WorkflowExecutionFinalizer = "otto.gptscript.ai/workflow-execution"
	WorkflowWorkspaceFinalizer = "otto.gptscript.ai/workflow-workspace"
	WorkflowKnowledgeFinalizer = "otto.gptscript.ai/workflow-knowledge"
	KnowledgeFileFinalizer     = "otto.gptscript.ai/knowledge-file"
)

const (
	PreviousRunNameLabel = "otto.gptscript.ai/previous-run-name"
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
	RunID          string       `json:"runID,omitempty"`
	Content        string       `json:"content"`
	Input          string       `json:"input,omitempty"`
	Prompt         *Prompt      `json:"prompt,omitempty"`
	Step           *Step        `json:"step,omitempty"`
	Tool           ToolProgress `json:"tool"`
	WaitingOnModel bool         `json:"waitingOnModel,omitempty"`
	Error          string       `json:"error,omitempty"`
}

type Prompt struct {
	ID        string            `json:"id,omitempty"`
	Time      metav1.Time       `json:"time,omitempty"`
	Message   string            `json:"message,omitempty"`
	Fields    []string          `json:"fields,omitempty"`
	Sensitive bool              `json:"sensitive,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ToolProgress struct {
	Name                   string `json:"name,omitempty"`
	Description            string `json:"description,omitempty"`
	Input                  string `json:"input,omitempty"`
	PartialInput           string `json:"partialInput,omitempty"`
	GeneratingInputForName string `json:"generatingInputForName,omitempty"`
	GeneratingInput        bool   `json:"generatingInput,omitempty"`
}

type RunSpec struct {
	Background            bool     `json:"background,omitempty"`
	ThreadName            string   `json:"threadName,omitempty"`
	AgentName             string   `json:"agentName,omitempty"`
	WorkflowName          string   `json:"workflowName,omitempty"`
	WorkflowExecutionName string   `json:"workflowExecutionName,omitempty"`
	WorkflowStepName      string   `json:"workflowStepName,omitempty"`
	WorkflowStepID        string   `json:"workflowStepID,omitempty"`
	WorkspaceID           string   `json:"workspaceID,omitempty"`
	PreviousRunName       string   `json:"previousRunName,omitempty"`
	Input                 string   `json:"input"`
	Env                   []string `json:"env,omitempty"`
	Tool                  string   `json:"tool,omitempty"`
	CredentialContextIDs  []string `json:"credentialContextIDs,omitempty"`
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
