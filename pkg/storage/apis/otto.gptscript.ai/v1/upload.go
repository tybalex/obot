package v1

import (
	"github.com/acorn-io/baaah/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OneDriveLinksFinalizer = "otto.gptscript.ai/onedrive-links"
	OneDriveLinksLabel     = "otto.gptscript.ai/onedrive-links"
)

var (
	_ conditions.Conditions = (*OneDriveLinks)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type OneDriveLinks struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OnedriveLinksSpec   `json:"spec,omitempty"`
	Status OnedriveLinksStatus `json:"status,omitempty"`
}

func (in *OneDriveLinks) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type OnedriveLinksSpec struct {
	AgentName    string   `json:"agentName,omitempty"`
	WorkflowName string   `json:"workflowName,omitempty"`
	SharedLinks  []string `json:"sharedLinks,omitempty"`
}

type OnedriveLinksStatus struct {
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ThreadName         string             `json:"threadName,omitempty"`
	RunName            string             `json:"runName,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Status             string             `json:"output,omitempty"`
	Error              string             `json:"error,omitempty"`
	Folders            FolderSet          `json:"folders,omitempty"`
}

type OneDriveLinksConnectorStatus struct {
	Status  string                 `json:"output,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Files   map[string]FileDetails `json:"files,omitempty"`
	Folders FolderSet              `json:"folders,omitempty"`
}

type FileDetails struct {
	FilePath  string `json:"filePath"`
	URL       string `json:"url"`
	UpdatedAt string `json:"updatedAt"`
}

type FolderSet map[string]Item

type Item struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type OneDriveLinksList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []OneDriveLinks `json:"items"`
}
