package agents

import (
	"encoding/json"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/aihelper"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

type AgentHandler struct {
	workspaceClient   *wclient.Client
	ingester          *knowledge.Ingester
	knowledgeBin      string
	workspaceProvider string
	aihelper          *aihelper.AIHelper
}

func New(wc *wclient.Client, ingester *knowledge.Ingester, knowledgeBin, wp string, aihelper *aihelper.AIHelper) *AgentHandler {
	return &AgentHandler{
		workspaceClient:   wc,
		ingester:          ingester,
		knowledgeBin:      knowledgeBin,
		workspaceProvider: wp,
		aihelper:          aihelper,
	}
}

const nameDescriptionPrompt = `
Given the following agent definition, suggest an appropriate name and description.
Be slightly funny, with a robot theme, and keep it short and sweet.
Response in json format.
{
	"name": "Agent Name",
	"description": "Agent Description"
}
`

func (a *AgentHandler) Suggestion(req router.Request, resp router.Response) error {
	agent := req.Object.(*v1.Agent)
	if agent.Spec.Manifest.Name != "" && agent.Spec.Manifest.Description != "" {
		return nil
	}

	input, err := json.Marshal(agent.Spec.Manifest)
	if err != nil {
		return err
	}

	var out struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := a.aihelper.GenerateObject(req.Ctx, &out, nameDescriptionPrompt, string(input)); err != nil {
		return err
	}

	var updated bool
	if agent.Spec.Manifest.Name == "" && out.Name != "" {
		agent.Spec.Manifest.Name = out.Name
		updated = true
	}
	if agent.Spec.Manifest.Description == "" && out.Description != "" {
		agent.Spec.Manifest.Description = out.Description
		updated = true
	}

	if updated {
		if err := req.Client.Update(req.Ctx, agent); err != nil {
			return err
		}
	}

	return nil
}

func (a *AgentHandler) CreateWorkspaces(req router.Request, resp router.Response) error {
	agent := req.Object.(*v1.Agent)
	if agent.Status.WorkspaceID != "" {
		return nil
	}

	workspaceID, err := a.workspaceClient.Create(req.Ctx, a.workspaceProvider)
	if err != nil {
		return err
	}

	knowledgeWorkspaceID, err := a.workspaceClient.Create(req.Ctx, a.workspaceProvider)
	if err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, workspaceID)
		return err
	}

	agent.Status.KnowledgeWorkspaceID = knowledgeWorkspaceID
	agent.Status.WorkspaceID = workspaceID

	if err := req.Client.Status().Update(req.Ctx, agent); err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, workspaceID)
		_ = a.workspaceClient.Rm(req.Ctx, knowledgeWorkspaceID)
		return err
	}

	return nil
}

func (a *AgentHandler) RemoveWorkspaces(req router.Request, resp router.Response) error {
	agent := req.Object.(*v1.Agent)
	if err := a.workspaceClient.Rm(req.Ctx, agent.Status.WorkspaceID); err != nil {
		return err
	}

	if agent.Status.HasKnowledge {
		if err := a.ingester.DeleteKnowledge(req.Ctx, agent.Namespace, agent.Status.KnowledgeWorkspaceID); err != nil {
			return err
		}
	}

	if agent.Status.KnowledgeWorkspaceID != "" {
		return a.workspaceClient.Rm(req.Ctx, agent.Status.KnowledgeWorkspaceID)
	}
	return nil
}

func (a *AgentHandler) IngestKnowledge(req router.Request, resp router.Response) error {
	agent := req.Object.(*v1.Agent)
	if agent.Status.KnowledgeGeneration == agent.Status.ObservedKnowledgeGeneration || !agent.Status.HasKnowledge {
		return nil
	}

	if err := a.ingester.IngestKnowledge(req.Ctx, agent.Namespace, agent.Status.KnowledgeWorkspaceID); err != nil {
		return err
	}

	agent.Status.ObservedKnowledgeGeneration = agent.Status.KnowledgeGeneration
	return nil
}
