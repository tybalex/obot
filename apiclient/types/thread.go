package types

import gptscriptclient "github.com/gptscript-ai/go-gptscript"

type Thread struct {
	Metadata
	ThreadManifest
	AgentID          string                   `json:"agentID,omitempty"`
	WorkflowID       string                   `json:"workflowID,omitempty"`
	LastRunID        string                   `json:"lastRunID,omitempty"`
	LastRunState     gptscriptclient.RunState `json:"lastRunState,omitempty"`
	PreviousThreadID string                   `json:"previousThreadID,omitempty"`
}

type ThreadList List[Thread]

type ThreadManifest struct {
	Tools       []string `json:"tools,omitempty"`
	Description string   `json:"description,omitempty"`
}
