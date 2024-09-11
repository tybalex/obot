package threads

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type ThreadHandler struct {
	Workspace *wclient.Client
	Ingester  *knowledge.Ingester
}

func (t *ThreadHandler) Cleanup(req router.Request, resp router.Response) error {
	thread := req.Object.(*v1.Thread)

	if thread.Spec.AgentName != "" {
		var agent v1.Agent
		if err := req.Get(&agent, thread.Namespace, thread.Spec.AgentName); apierrors.IsNotFound(err) {
			return req.Delete(thread)
		} else if err != nil {
			return err
		}
	}

	if thread.Spec.WorkflowStepName != "" {
		var step v1.WorkflowStep
		if err := req.Get(&step, thread.Namespace, thread.Spec.WorkflowStepName); apierrors.IsNotFound(err) {
			return req.Delete(thread)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (t *ThreadHandler) RemoveWorkspaces(req router.Request, resp router.Response) error {
	thread := req.Object.(*v1.Thread)
	if err := t.Workspace.Rm(req.Ctx, thread.Spec.WorkspaceID); err != nil {
		return err
	}

	if thread.Status.HasKnowledge {
		if err := t.Ingester.DeleteKnowledge(req.Ctx, thread.Namespace, thread.Spec.KnowledgeWorkspaceID); err != nil {
			return err
		}
	}

	if thread.Spec.KnowledgeWorkspaceID != "" {
		return t.Workspace.Rm(req.Ctx, thread.Spec.KnowledgeWorkspaceID)
	}

	return nil
}

func (t *ThreadHandler) IngestKnowledge(req router.Request, resp router.Response) error {
	thread := req.Object.(*v1.Thread)
	if !thread.Status.IngestKnowledge || !thread.Status.HasKnowledge {
		return nil
	}

	if err := t.Ingester.IngestKnowledge(req.Ctx, thread.Namespace, thread.Spec.KnowledgeWorkspaceID); err != nil {
		return err
	}

	thread.Status.IngestKnowledge = false
	return nil
}
