package invoke

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/apiclient/types"
	"github.com/gptscript-ai/otto/pkg/render"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
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
	return i.toAgent(&wf, step, *wfe.Status.WorkflowManifest)
}

func (i *Invoker) toAgent(wf *v1.Workflow, step *v1.WorkflowStep, manifest types.WorkflowManifest) (v1.Agent, error) {
	agent := render.Workflow(wf, render.WorkflowOptions{
		ManifestOverride: &manifest,
		Step:             &step.Spec.Step,
	})
	return *agent, nil
}

func concatOrNotJSONMaps(one string, args map[string]string) string {
	result := map[string]string{}
	if one != "" {
		if err := json.Unmarshal([]byte(one), &result); err != nil {
			// Not JSON, just use as is
			return one
		}
	}

	for k, v := range args {
		result[k] = v
	}

	data, _ := json.Marshal(result)
	return string(data)
}

func (i *Invoker) getInput(step *v1.WorkflowStep) (string, error) {
	if step.Spec.Step.Template != nil && step.Spec.Step.Template.Name != "" {
		return concatOrNotJSONMaps(step.Spec.Input, step.Spec.Step.Template.Args), nil
	}

	var content []string
	if step.Spec.Input != "" {
		content = append(content, step.Spec.Input)
	}
	if step.Spec.Step.Step != "" {
		content = append(content, step.Spec.Step.Step)
	}
	return strings.Join(content, "\n"), nil
}
