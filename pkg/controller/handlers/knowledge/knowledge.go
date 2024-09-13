package knowledge

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	workspaceClient   *wclient.Client
	ingester          *knowledge.Ingester
	workspaceProvider string
}

func New(wc *wclient.Client, ingester *knowledge.Ingester, wp string) *Handler {
	return &Handler{
		workspaceClient:   wc,
		ingester:          ingester,
		workspaceProvider: wp,
	}
}

type knowledgeable interface {
	kclient.Object
	GetKnowledgeWorkspaceStatus() *v1.KnowledgeWorkspaceStatus
}

func (a *Handler) CreateWorkspace(req router.Request, resp router.Response) error {
	knowledged := req.Object.(knowledgeable)
	status := knowledged.GetKnowledgeWorkspaceStatus()
	if status.KnowledgeWorkspaceID != "" {
		return nil
	}

	knowledgeWorkspaceID, err := a.workspaceClient.Create(req.Ctx, a.workspaceProvider)
	if err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, knowledgeWorkspaceID)
		return err
	}

	status.KnowledgeWorkspaceID = knowledgeWorkspaceID

	if err := req.Client.Status().Update(req.Ctx, knowledged); err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, knowledgeWorkspaceID)
		return err
	}

	return nil
}

func (a *Handler) RemoveWorkspace(req router.Request, resp router.Response) error {
	knowledged := req.Object.(knowledgeable)
	status := knowledged.GetKnowledgeWorkspaceStatus()

	if status.HasKnowledge {
		if err := a.ingester.DeleteKnowledge(req.Ctx, knowledged.GetNamespace(), status.KnowledgeWorkspaceID); err != nil {
			return err
		}
	}

	if status.KnowledgeWorkspaceID != "" {
		return a.workspaceClient.Rm(req.Ctx, status.KnowledgeWorkspaceID)
	}

	return nil
}

func (a *Handler) IngestKnowledge(req router.Request, resp router.Response) error {
	knowleged := req.Object.(knowledgeable)
	status := knowleged.GetKnowledgeWorkspaceStatus()
	if status.KnowledgeGeneration == status.ObservedKnowledgeGeneration || !status.HasKnowledge {
		return nil
	}

	if err := a.ingester.IngestKnowledge(req.Ctx, knowleged.GetNamespace(), status.KnowledgeWorkspaceID); err != nil {
		return err
	}

	status.ObservedKnowledgeGeneration = status.KnowledgeGeneration
	return nil
}
