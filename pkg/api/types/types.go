package types

import (
	"time"

	gptscriptclient "github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type List[T any] struct {
	Items []T `json:"items"`
}

type Metadata struct {
	ID      string            `json:"id,omitempty"`
	Created time.Time         `json:"created,omitempty"`
	Deleted *time.Time        `json:"deleted,omitempty"`
	Links   map[string]string `json:"links,omitempty"`
}

func MetadataFrom(obj kclient.Object, linkKV ...string) Metadata {
	m := Metadata{
		ID:      obj.GetName(),
		Created: obj.GetCreationTimestamp().Time,
		Links:   map[string]string{},
	}
	if delTime := obj.GetDeletionTimestamp(); delTime != nil {
		m.Deleted = &delTime.Time
	}
	for i := 0; i < len(linkKV); i += 2 {
		m.Links[linkKV[i]] = linkKV[i+1]
	}
	return m
}

type Agent struct {
	Metadata
	v1.AgentManifest
	v1.AgentExternalStatus
}

type AgentList List[Agent]

type Workflow struct {
	Metadata
	v1.WorkflowManifest
	v1.WorkflowExternalStatus
}

type WorkflowList List[Workflow]

type Thread struct {
	Metadata
	v1.ThreadManifest
	AgentID          string                   `json:"agentID,omitempty"`
	WorkflowID       string                   `json:"workflowID,omitempty"`
	LastRunID        string                   `json:"lastRunID,omitempty"`
	LastRunState     gptscriptclient.RunState `json:"lastRunState,omitempty"`
	PreviousThreadID string                   `json:"previousThreadID,omitempty"`
}

type ThreadList List[Thread]

type File struct {
	Name string `json:"name,omitempty"`
}

type FileList List[File]

type Run struct {
	ID             string    `json:"id,omitempty"`
	Created        time.Time `json:"created,omitempty"`
	ThreadID       string    `json:"threadID,omitempty"`
	AgentID        string    `json:"agentID,omitempty"`
	WorkflowID     string    `json:"workflowID,omitempty"`
	WorkflowStepID string    `json:"workflowStepID,omitempty"`
	PreviousRunID  string    `json:"previousRunID,omitempty"`
	Input          string    `json:"input"`
	State          string    `json:"state,omitempty"`
	Output         string    `json:"output,omitempty"`
	Error          string    `json:"error,omitempty"`
}

type RunList List[Run]

type RunDebug struct {
	Spec   v1.RunSpec                           `json:"spec"`
	Status v1.RunStatus                         `json:"status"`
	Frames map[string]gptscriptclient.CallFrame `json:"frames"`
}

type Credential struct {
	ContextID string     `json:"contextID,omitempty"`
	Name      string     `json:"name,omitempty"`
	EnvVars   []string   `json:"envVars,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type CredentialList List[Credential]

type InvokeResponse struct {
	Events   <-chan Progress
	ThreadID string
}

type Progress v1.Progress

type KnowledgeFile struct {
	Metadata
	FileName        string             `json:"fileName"`
	AgentID         string             `json:"agentID,omitempty"`
	WorkflowID      string             `json:"workflowID,omitempty"`
	ThreadID        string             `json:"threadID,omitempty"`
	UploadID        string             `json:"uploadID,omitempty"`
	IngestionStatus v1.IngestionStatus `json:"ingestionStatus,omitempty"`
	FileDetails     v1.FileDetails     `json:"fileDetails,omitempty"`
}

type KnowledgeFileList List[KnowledgeFile]

type OneDriveLinks struct {
	Metadata
	AgentID     string       `json:"agentID,omitempty"`
	WorkflowID  string       `json:"workflowID,omitempty"`
	SharedLinks []string     `json:"sharedLinks,omitempty"`
	ThreadID    string       `json:"threadID,omitempty"`
	RunID       string       `json:"runID,omitempty"`
	Status      string       `json:"output,omitempty"`
	Error       string       `json:"error,omitempty"`
	Folders     v1.FolderSet `json:"folders,omitempty"`
}

type OneDriveLinksList List[OneDriveLinks]

type ToolReferenceManifest struct {
	Name      string               `json:"name"`
	ToolType  v1.ToolReferenceType `json:"toolType"`
	Reference string               `json:"reference,omitempty"`
}

type ToolReference struct {
	Metadata
	ToolReferenceManifest
	Error       string            `json:"error,omitempty"`
	Description string            `json:"description,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
}

type ToolReferenceList List[ToolReference]
