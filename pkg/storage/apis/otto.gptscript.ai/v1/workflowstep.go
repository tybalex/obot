package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	"github.com/gptscript-ai/otto/apiclient/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ conditions.Conditions = (*WorkflowStep)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowStep struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowStepSpec   `json:"spec,omitempty"`
	Status WorkflowStepStatus `json:"status,omitempty"`
}

func (in *WorkflowStep) GetColumns() [][]string {
	return [][]string{
		{"Name", "Name"},
		{"State", "Status.State"},
		{"After", "Spec.AfterWorkflowStepName"},
		{"Runs", "{{ .Status.RunNames | arrayNoSpace }}"},
		{"LastRun", "Status.LastRunName"},
		{"StepID", "Spec.Step.ID"},
		{"WFE", "Spec.WorkflowExecutionName"},
		{"Created", "{{ago .CreationTimestamp}}"},
	}
}

func (in *WorkflowStep) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type WorkflowStepSpec struct {
	AfterWorkflowStepName string     `json:"afterWorkflowStepName,omitempty"`
	Step                  types.Step `json:"step,omitempty"`
	WorkflowExecutionName string     `json:"workflowExecutionName,omitempty"`
}

func (in *WorkflowStep) DeleteRefs() []Ref {
	refs := []Ref{
		{ObjType: &WorkflowExecution{}, Name: in.Spec.WorkflowExecutionName},
		{ObjType: &Run{}, Name: in.Status.LastRunName},
	}
	for _, run := range in.Status.RunNames {
		refs = append(refs, Ref{ObjType: &Run{}, Name: run})
	}
	return refs
}

type WorkflowStepState string

const (
	WorkflowStepStatePending  WorkflowStepState = "Pending"
	WorkflowStepStateRunning  WorkflowStepState = "Running"
	WorkflowStepStateError    WorkflowStepState = "Error"
	WorkflowStepStateComplete WorkflowStepState = "Complete"
	WorkflowStepStateSubCall  WorkflowStepState = "SubCall"
)

type WorkflowStepStatus struct {
	State       WorkflowStepState  `json:"state,omitempty"`
	SubCalls    []SubCall          `json:"subCalls,omitempty"`
	Error       string             `json:"message,omitempty"`
	ThreadName  string             `json:"threadName,omitempty"`
	RunNames    []string           `json:"runNames,omitempty"`
	LastRunName string             `json:"lastRunName,omitempty"`
	Conditions  []metav1.Condition `json:"conditions,omitempty"`
}

type SubCall struct {
	Type     string `json:"type,omitempty"`
	Workflow string `json:"workflow,omitempty"`
	Input    string `json:"input,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowStepList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WorkflowStep `json:"items"`
}
