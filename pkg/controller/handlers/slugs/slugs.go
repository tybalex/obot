package slugs

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SlugGC(req router.Request, _ router.Response) error {
	slug := req.Object.(*v1.Slug)

	if slug.Spec.WorkflowName != "" {
		var workflow v1.Workflow
		if err := req.Get(&workflow, slug.Namespace, slug.Spec.WorkflowName); apierrors.IsNotFound(err) {
			return req.Delete(slug)
		} else if err != nil {
			return err
		}
		if workflow.Spec.Manifest.Slug != slug.Name {
			return req.Delete(slug)
		}
	}

	if slug.Spec.AgentName != "" {
		var agent v1.Agent
		if err := req.Get(&agent, slug.Namespace, slug.Spec.AgentName); apierrors.IsNotFound(err) {
			return req.Delete(slug)
		} else if err != nil {
			return err
		}
		if agent.Spec.Manifest.Slug != slug.Name {
			return req.Delete(slug)
		}
	}

	return nil
}

func AssociateWithSlug(req router.Request, _ router.Response) error {
	slug := v1.Slug{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
		},
	}

	switch v := req.Object.(type) {
	case *v1.Workflow:
		slug.Name = v.Spec.Manifest.Slug
		slug.Spec.WorkflowName = v.Name
	case *v1.Agent:
		slug.Name = v.Spec.Manifest.Slug
		slug.Spec.AgentName = v.Name
	}

	if slug.Name == "" {
		return nil
	}

	var existingSlug v1.Slug
	if err := req.Get(&existingSlug, slug.Namespace, slug.Name); apierrors.IsNotFound(err) {
		if err := req.Client.Create(req.Ctx, &slug); err != nil {
			return err
		}
	} else if err != nil {
		return nil
	}

	assigned := existingSlug.Spec == slug.Spec
	switch v := req.Object.(type) {
	case *v1.Workflow:
		v.Status.External.SlugAssigned = assigned
	case *v1.Agent:
		v.Status.External.SlugAssigned = assigned
	}
	return nil
}
