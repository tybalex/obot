package types

type WorkflowState string

const (
	WorkflowStatePending  WorkflowState = "Pending"
	WorkflowStateRunning  WorkflowState = "Running"
	WorkflowStateError    WorkflowState = "Error"
	WorkflowStateComplete WorkflowState = "Complete"
	WorkflowStateSubCall  WorkflowState = "SubCall"
	WorkflowStateBlocked  WorkflowState = "Blocked"
)

func (in WorkflowState) IsBlocked() bool {
	return in == WorkflowStateBlocked || in == WorkflowStateError
}

func (in WorkflowState) IsTerminal() bool {
	return in == WorkflowStateComplete || in == WorkflowStateError
}

type Thread struct {
	Metadata
	ThreadManifest
	AgentID        string `json:"agentID,omitempty"`
	WorkflowID     string `json:"workflowID,omitempty"`
	State          string `json:"state,omitempty"`
	LastRunID      string `json:"lastRunID,omitempty"`
	CurrentRunID   string `json:"currentRunID,omitempty"`
	ParentThreadID string `json:"parentThreadID,omitempty"`
}

type ThreadList List[Thread]

type ThreadManifest struct {
	Tools       []string `json:"tools,omitempty"`
	Description string   `json:"description,omitempty"`
}
