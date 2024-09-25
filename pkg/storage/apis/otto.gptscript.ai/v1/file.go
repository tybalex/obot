package v1

import (
	"crypto/sha256"
	"fmt"

	"github.com/acorn-io/baaah/pkg/fields"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KnowledgeFile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KnowledgeFileSpec   `json:"spec,omitempty"`
	Status KnowledgeFileStatus `json:"status,omitempty"`
}

func (k *KnowledgeFile) DeleteRefs() []Ref {
	return []Ref{
		{ObjType: new(Workspace), Name: k.Spec.WorkspaceName},
		{ObjType: new(OneDriveLinks), Name: k.Spec.UploadName},
	}
}

func (k *KnowledgeFile) Has(field string) bool {
	return k.Get(field) != ""
}

func (k *KnowledgeFile) Get(field string) string {
	if k == nil {
		return ""
	}

	switch field {
	case "spec.workspaceName":
		return k.Spec.WorkspaceName
	case "spec.uploadName":
		return k.Spec.UploadName
	}

	return ""
}

func (*KnowledgeFile) FieldNames() []string {
	return []string{"spec.workspaceName", "spec.uploadName"}
}

var _ fields.Fields = (*KnowledgeFile)(nil)

type KnowledgeFileSpec struct {
	FileName      string `json:"fileName"`
	WorkspaceName string `json:"workspaceName,omitempty"`
	UploadName    string `json:"uploadName,omitempty"`
}

type KnowledgeFileStatus struct {
	IngestionStatus IngestionStatus `json:"ingestionStatus,omitempty"`
	FileDetails     FileDetails     `json:"fileDetails,omitempty"`
	UploadID        string          `json:"uploadID,omitempty"`
}

type IngestionStatus struct {
	Count        int    `json:"count,omitempty"`
	Reason       string `json:"reason,omitempty"`
	AbsolutePath string `json:"absolute_path,omitempty"`
	BasePath     string `json:"basePath,omitempty"`
	Filename     string `json:"filename,omitempty"`
	VectorStore  string `json:"vectorstore,omitempty"`
	Message      string `json:"msg,omitempty"`
	Flow         string `json:"flow,omitempty"`
	RootPath     string `json:"rootPath,omitempty"`
	Filepath     string `json:"filepath,omitempty"`
	Phase        string `json:"phase,omitempty"`
	NumDocuments int    `json:"num_documents,omitempty"`
	Stage        string `json:"stage,omitempty"`
	Status       string `json:"status,omitempty"`
	Component    string `json:"component,omitempty"`
	FileType     string `json:"filetype,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KnowledgeFileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KnowledgeFile `json:"items"`
}

func ObjectNameFromAbsolutePath(absolutePath string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(absolutePath)))
}
