package types

type Webhook struct {
	Metadata
	WebhookManifest
	RefNameAssigned            bool  `json:"refNameAssigned,omitempty"`
	LastSuccessfulRunCompleted *Time `json:"lastSuccessfulRunCompleted,omitempty"`
}

type WebhookManifest struct {
	Description      string   `json:"description,omitempty"`
	RefName          string   `json:"refName,omitempty"`
	WorkflowID       string   `json:"workflowID,omitempty"`
	Headers          []string `json:"headers,omitempty"`
	Secret           string   `json:"secret,omitempty"`
	ValidationHeader string   `json:"validationHeader,omitempty"`
}

type WebhookExternalStatus struct {
	RefName         string `json:"refName,omitempty"`
	RefNameAssigned bool   `json:"refNameAssigned,omitempty"`
}

type WebhookList List[Webhook]
