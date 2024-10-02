package types

type WorkflowState string

const (
	WorkflowStatePending  WorkflowState = "Pending"
	WorkflowStateRunning  WorkflowState = "Running"
	WorkflowStateError    WorkflowState = "Error"
	WorkflowStateComplete WorkflowState = "Complete"
)

type Thread struct {
	Metadata
	ThreadManifest
	AgentID        string `json:"agentID,omitempty"`
	WorkflowID     string `json:"workflowID,omitempty"`
	State          string `json:"state,omitempty"`
	LastRunID      string `json:"lastRunID,omitempty"`
	ParentThreadID string `json:"parentThreadID,omitempty"`
}

type ThreadList List[Thread]

type ThreadManifest struct {
	Tools       []string `json:"tools,omitempty"`
	Description string   `json:"description,omitempty"`
}
