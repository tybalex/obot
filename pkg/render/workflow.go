package render

import (
	"github.com/gptscript-ai/otto/apiclient/types"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowOptions struct {
	Step             *types.Step
	ManifestOverride *types.WorkflowManifest
}

func Workflow(wf *v1.Workflow, opts WorkflowOptions) *v1.Agent {
	agentManifest := wf.Spec.Manifest.AgentManifest
	if opts.ManifestOverride != nil {
		agentManifest = opts.ManifestOverride.AgentManifest
	}

	agent := v1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: wf.Namespace,
		},
		Spec: v1.AgentSpec{
			Manifest:            agentManifest,
			CredentialContextID: wf.Name,
		},
		Status: v1.AgentStatus{
			WorkspaceName:          wf.Status.WorkspaceName,
			KnowledgeWorkspaceName: wf.Status.KnowledgeWorkspaceName,
		},
	}

	if step := opts.Step; step != nil {
		if step.Cache != nil {
			agent.Spec.Manifest.Cache = step.Cache
		}
		if step.Temperature != nil {
			agent.Spec.Manifest.Temperature = step.Temperature
		}

		agent.Spec.Manifest.Tools = append(agent.Spec.Manifest.Tools, step.Tools...)
		agent.Spec.Manifest.Agents = append(agent.Spec.Manifest.Agents, step.Agents...)
		agent.Spec.Manifest.Workflows = append(agent.Spec.Manifest.Workflows, step.Workflows...)
		if step.Template != nil && step.Template.Name != "" {
			agent.Spec.InputFilters = append(agent.Spec.InputFilters, step.Template.Name)
		}
	}

	if agent.Spec.Manifest.Prompt == "" {
		agent.Spec.Manifest.Prompt = v1.DefaultWorkflowAgentPrompt
	}

	return &agent
}
