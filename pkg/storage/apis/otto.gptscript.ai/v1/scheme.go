package v1

import (
	otto_gptscript_ai "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const Version = "v1"

var SchemeGroupVersion = schema.GroupVersion{
	Group:   otto_gptscript_ai.Group,
	Version: Version,
}

func AddToScheme(scheme *runtime.Scheme) error {
	return AddToSchemeWithGV(scheme, SchemeGroupVersion)
}

func AddToSchemeWithGV(scheme *runtime.Scheme, schemeGroupVersion schema.GroupVersion) error {
	scheme.AddKnownTypes(schemeGroupVersion,
		&Agent{},
		&AgentList{},
		&Run{},
		&RunList{},
		&RunState{},
		&RunStateList{},
		&Slug{},
		&SlugList{},
		&Thread{},
		&ThreadList{},
		&Workflow{},
		&WorkflowList{},
		&WorkflowExecution{},
		&WorkflowExecutionList{},
		&WorkflowStep{},
		&WorkflowStepList{},
		&OneDriveLinks{},
		&OneDriveLinksList{},
	)

	// Add common types
	scheme.AddKnownTypes(schemeGroupVersion, &metav1.Status{})

	if schemeGroupVersion == SchemeGroupVersion {
		// Add the watch version that applies
		metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	}
	return nil
}
