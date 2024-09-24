package invoke

import (
	"context"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type StepOptions struct {
	PreviousRunName string
	Continue        *string
}

func (i *Invoker) Step(ctx context.Context, c kclient.Client, step *v1.WorkflowStep, opt StepOptions) (*Response, error) {
	agent, err := i.toAgentFromStep(ctx, c, step)
	if err != nil {
		return nil, err
	}

	input, err := i.getInput(step)
	if err != nil {
		return nil, err
	}

	if opt.Continue != nil {
		input = *opt.Continue
	}

	var wfe v1.WorkflowExecution
	if err := c.Get(ctx, router.Key(step.Namespace, step.Spec.WorkflowExecutionName), &wfe); err != nil {
		return nil, err
	}

	return i.Agent(ctx, c, &agent, input, Options{
		Background:            true,
		ThreadName:            wfe.Status.ThreadName,
		PreviousRunName:       opt.PreviousRunName,
		WorkflowName:          wfe.Spec.WorkflowName,
		WorkflowExecutionName: step.Spec.WorkflowExecutionName,
		WorkflowStepName:      step.Name,
		WorkflowStepID:        step.Spec.Step.ID,
	})
}

func (i *Invoker) toAgentFromStep(ctx context.Context, c kclient.Client, step *v1.WorkflowStep) (v1.Agent, error) {
	var (
		wf  v1.Workflow
		wfe v1.WorkflowExecution
	)
	if err := c.Get(ctx, router.Key(step.Namespace, step.Spec.WorkflowExecutionName), &wfe); err != nil {
		return v1.Agent{}, err
	}
	if err := c.Get(ctx, router.Key(step.Namespace, wfe.Spec.WorkflowName), &wf); err != nil {
		return v1.Agent{}, err
	}
	agent, err := i.toAgent(&wf, *wfe.Status.WorkflowManifest)
	if err != nil {
		return v1.Agent{}, err
	}
	if step.Spec.Step.Cache != nil {
		agent.Spec.Manifest.Cache = step.Spec.Step.Cache
	}
	if step.Spec.Step.Temperature != nil {
		agent.Spec.Manifest.Temperature = step.Spec.Step.Temperature
	}
	return agent, nil
}

func (i *Invoker) toAgent(wf *v1.Workflow, manifest v1.WorkflowManifest) (v1.Agent, error) {
	agent := v1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: wf.Namespace,
		},
		Spec: v1.AgentSpec{
			Manifest: manifest.AgentManifest,
		},
		Status: v1.AgentStatus{
			Workspace:          wf.Status.Workspace,
			KnowledgeWorkspace: wf.Status.KnowledgeWorkspace,
		},
	}
	if agent.Spec.Manifest.Prompt == "" {
		agent.Spec.Manifest.Prompt = v1.DefaultWorkflowAgentPrompt
	}
	return agent, nil
}

func (i *Invoker) getInput(step *v1.WorkflowStep) (string, error) {
	var content []string
	if step.Spec.Input != "" {
		content = append(content, step.Spec.Input)
	}
	if step.Spec.Step.Step != "" {
		content = append(content, step.Spec.Step.Step)
	}
	return strings.Join(content, "\n"), nil
}
