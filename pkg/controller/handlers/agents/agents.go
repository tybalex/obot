package agents

import (
	"fmt"
	"os/exec"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/workspace"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

type AgentHandler struct {
	WorkspaceClient *wclient.Client
	KnowledgeBin    string
}

func (a *AgentHandler) RemoveWorkspaces(req router.Request, resp router.Response) error {
	agent := req.Object.(*v1.Agent)
	if err := a.WorkspaceClient.Rm(req.Ctx, agent.Spec.WorkspaceID); err != nil {
		return err
	}

	if agent.Status.HasKnowledge {
		if err := exec.Command(a.KnowledgeBin, "delete-dataset", agent.Spec.KnowledgeWorkspaceID).Run(); err != nil {
			return fmt.Errorf("failed to delete knowledge dataset: %w", err)
		}
	}

	if agent.Spec.KnowledgeWorkspaceID != "" {
		return a.WorkspaceClient.Rm(req.Ctx, agent.Spec.KnowledgeWorkspaceID)
	}
	return nil
}

func (a *AgentHandler) IngestKnowledge(req router.Request, resp router.Response) error {
	agent := req.Object.(*v1.Agent)
	if !agent.Status.IngestKnowledge || !agent.Status.HasKnowledge {
		return nil
	}

	if err := workspace.IngestKnowledge(a.KnowledgeBin, agent.Spec.KnowledgeWorkspaceID); err != nil {
		return err
	}

	agent.Status.IngestKnowledge = false
	return nil
}
