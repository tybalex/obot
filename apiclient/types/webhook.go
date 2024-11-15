package types

type Webhook struct {
	Metadata
	WebhookManifest
	RefNameAssigned            bool  `json:"refNameAssigned,omitempty"`
	LastSuccessfulRunCompleted *Time `json:"lastSuccessfulRunCompleted,omitempty"`
}

type WebhookManifest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	RefName          string   `json:"refName"`
	Workflow         string   `json:"workflow"`
	Headers          []string `json:"headers"`
	Secret           string   `json:"secret"`
	ValidationHeader string   `json:"validationHeader"`
}

type WebhookExternalStatus struct {
	RefName         string `json:"refName,omitempty"`
	RefNameAssigned bool   `json:"refNameAssigned,omitempty"`
}

type WebhookList List[Webhook]
