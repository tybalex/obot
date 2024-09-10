package threads

import (
	"fmt"
	"os/exec"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/workspace"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func Cleanup(req router.Request, resp router.Response) error {
	thread := req.Object.(*v1.Thread)
	var agent v1.Agent

	if err := req.Get(&agent, thread.Namespace, thread.Spec.AgentName); apierrors.IsNotFound(err) {
		return req.Client.Delete(req.Ctx, thread)
	} else if err != nil {
		return err
	}

	return nil
}

func RemoveWorkspace(wc *wclient.Client, knowledgeBin string) router.HandlerFunc {
	return func(req router.Request, resp router.Response) error {
		thread := req.Object.(*v1.Thread)
		if err := wc.Rm(req.Ctx, thread.Spec.WorkspaceID); err != nil {
			return err
		}

		if thread.Status.HasKnowledge {
			if err := exec.Command(knowledgeBin, "delete-dataset", thread.Spec.KnowledgeWorkspaceID).Run(); err != nil {
				return fmt.Errorf("failed to delete knowledge dataset: %w", err)
			}
		}

		if thread.Spec.KnowledgeWorkspaceID != "" {
			return wc.Rm(req.Ctx, thread.Spec.KnowledgeWorkspaceID)
		}

		return nil
	}
}

func IngestKnowledge(knowledgeBin string) router.HandlerFunc {
	return func(req router.Request, resp router.Response) error {
		thread := req.Object.(*v1.Thread)
		if !thread.Status.IngestKnowledge || !thread.Status.HasKnowledge {
			return nil
		}

		if err := workspace.IngestKnowledge(knowledgeBin, thread.Spec.KnowledgeWorkspaceID); err != nil {
			return err
		}

		thread.Status.IngestKnowledge = false
		return nil
	}
}
