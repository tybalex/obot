package types

type EmailReceiver struct {
	Metadata
	EmailReceiverManifest
	AddressAssigned *bool  `json:"aliasAssigned,omitempty"`
	EmailAddress    string `json:"emailAddress,omitempty"`
}

type EmailReceiverManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Alias          string   `json:"alias,omitempty"`
	Workflow       string   `json:"workflow"`
	AllowedSenders []string `json:"allowedSenders,omitempty"`
}

type EmailReceiverList List[EmailReceiver]
