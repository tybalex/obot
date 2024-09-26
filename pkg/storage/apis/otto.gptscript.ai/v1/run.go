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
	RunFinalizer               = "otto.gptscript.ai/run"
	WorkflowExecutionFinalizer = "otto.gptscript.ai/workflow-execution"
	KnowledgeFileFinalizer     = "otto.gptscript.ai/knowledge-file"
	WorkspaceFinalizer         = "otto.gptscript.ai/workspace"
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

func (in *Run) GetColumns() [][]string {
	return [][]string{
		{"Name", "Name"},
		{"PreviousRun", "Spec.PreviousRunName"},
		{"State", "Status.State"},
		{"Thread", "Spec.ThreadName"},
		{"Agent", "Spec.AgentName"},
		{"Workflow", "Spec.WorkflowName"},
		{"Step", "Spec.WorkflowStepName"},
		{"Created", "{{ago .CreationTimestamp}}"},
	}
}

func (in *Run) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type Progress struct {
	// RunID should be populated for all progress events to associate this event with a run
	// If RunID is not populated, the event will not specify tied to any particular run
	RunID string `json:"runID,omitempty"`

	// Content is the output data. The content for all events should be concatenated to form the entire output
	// If you wish to print event and are not concerned with tracking the internal progress when one can just
	// only the content field in a very simple loop
	Content string `json:"content"`

	// NOTE: Only one of the follow fields will be populated, never more than one. If none of the below fields are
	// populated, you should only care about the content field which should have some content to print. You should
	// process the below fields first before considering the content field.

	// Some input that was provided to the run
	Input string `json:"input,omitempty"`
	// If prompt is set content will also me set, but you can ignore the content field and instead handle the explicit
	// information in the prompt field which will provider more information for things such as OAuth
	Prompt *Prompt `json:"prompt,omitempty"`
	// The step that is currently being executed. When this is set the following events are assumed to be part of
	// this step until the next step is set. This field is not always set, only set when the set changes
	Step *Step `json:"step,omitempty"`
	// ToolInput indicates the LLM is currently generating tool arguments which can sometime take a while
	ToolInput *ToolInput `json:"toolInput,omitempty"`
	// ToolCall indicates the LLM is currently calling a tool.
	ToolCall *ToolCall `json:"toolCall,omitempty"`
	// WaitingOnModel indicates we are waiting for the model to start responding with content
	WaitingOnModel bool `json:"waitingOnModel,omitempty"`
	// Error indicates that an error occurred
	Error string `json:"error,omitempty"`
}

type Prompt struct {
	ID        string            `json:"id,omitempty"`
	Time      metav1.Time       `json:"time,omitempty"`
	Message   string            `json:"message,omitempty"`
	Fields    []string          `json:"fields,omitempty"`
	Sensitive bool              `json:"sensitive,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ToolInput struct {
	Content          string `json:"content,omitempty"`
	InternalToolName string `json:"internalToolName,omitempty"`
}

type ToolCall struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Input       string `json:"input,omitempty"`
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

func (in *Run) DeleteRefs() []Ref {
	return []Ref{
		{ObjType: &Thread{}, Name: in.Spec.ThreadName},
	}
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
