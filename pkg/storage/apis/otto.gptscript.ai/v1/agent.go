package v1

import (
	"github.com/otto8-ai/otto8/apiclient/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec,omitempty"`
	Status AgentStatus `json:"status,omitempty"`
}

type AgentSpec struct {
	Manifest            types.AgentManifest `json:"manifest,omitempty"`
	InputFilters        []string            `json:"inputFilters,omitempty"`
	CredentialContextID string              `json:"credentialContextID,omitempty"`
}

type AgentStatus struct {
	RefNameAssigned   bool     `json:"refNameAssigned,omitempty"`
	KnowledgeSetNames []string `json:"knowledgeSetNames,omitempty"`
	WorkspaceName     string   `json:"workspaceName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Agent `json:"items"`
}
