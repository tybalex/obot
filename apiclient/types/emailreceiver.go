package types

type EmailReceiver struct {
	Metadata
	EmailReceiverManifest
	AliasAssigned bool `json:"aliasAssigned,omitempty"`
}

type EmailReceiverManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Alias          string   `json:"alias"`
	Workflow       string   `json:"workflow"`
	AllowedSenders []string `json:"allowedSenders,omitempty"`
}

type EmailReceiverList List[EmailReceiver]
