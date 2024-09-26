package types

type ToolReferenceType string

const (
	ToolReferenceTypeTool         ToolReferenceType = "tool"
	ToolReferenceTypeStepTemplate ToolReferenceType = "stepTemplate"
)

type ToolReferenceManifest struct {
	Name      string            `json:"name"`
	ToolType  ToolReferenceType `json:"toolType"`
	Reference string            `json:"reference,omitempty"`
}

type ToolReference struct {
	Metadata
	ToolReferenceManifest
	Error       string            `json:"error,omitempty"`
	Description string            `json:"description,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
}

type ToolReferenceList List[ToolReference]
