package agents

import (
	"encoding/json"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/otto8-ai/otto8/pkg/aihelper"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/otto8-ai/otto8/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AgentHandler struct {
	aihelper *aihelper.AIHelper
}

func New(aihelper *aihelper.AIHelper) *AgentHandler {
	return &AgentHandler{
		aihelper: aihelper,
	}
}

func (a *AgentHandler) WorkspaceObjects(req router.Request, _ router.Response) error {
	agent := req.Object.(*v1.Agent)
	if agent.Status.WorkspaceName == "" {
		ws := &v1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    req.Namespace,
				GenerateName: system.WorkspacePrefix,
			},
			Spec: v1.WorkspaceSpec{
				AgentName: agent.Name,
			},
		}
		if err := req.Client.Create(req.Ctx, ws); err != nil {
			return err
		}

		agent.Status.WorkspaceName = ws.Name
	}

	if agent.Status.KnowledgeWorkspaceName == "" {
		ws := &v1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    req.Namespace,
				GenerateName: system.WorkspacePrefix,
			},
			Spec: v1.WorkspaceSpec{
				AgentName:   agent.Name,
				IsKnowledge: true,
			},
		}
		if err := req.Client.Create(req.Ctx, ws); err != nil {
			return err
		}

		agent.Status.KnowledgeWorkspaceName = ws.Name
	}

	return nil
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

	// Don't generate anything until we have a prompt
	if agent.Spec.Manifest.Prompt == "" {
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
