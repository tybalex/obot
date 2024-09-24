package reference

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AssociateWithReference(req router.Request, _ router.Response) error {
	ref := v1.Reference{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
		},
	}

	switch v := req.Object.(type) {
	case *v1.Workflow:
		ref.Name = v.Spec.Manifest.RefName
		ref.Spec.WorkflowName = v.Name
	case *v1.Agent:
		ref.Name = v.Spec.Manifest.RefName
		ref.Spec.AgentName = v.Name
	}

	if ref.Name == "" {
		return nil
	}

	var existingRef v1.Reference
	if err := req.Get(&existingRef, ref.Namespace, ref.Name); apierrors.IsNotFound(err) {
		if err := req.Client.Create(req.Ctx, &ref); err != nil {
			return err
		}
	} else if err != nil {
		return nil
	}

	assigned := existingRef.Spec == ref.Spec
	switch v := req.Object.(type) {
	case *v1.Workflow:
		v.Status.External.RefNameAssigned = assigned
	case *v1.Agent:
		v.Status.External.RefNameAssigned = assigned
	}
	return nil
}
