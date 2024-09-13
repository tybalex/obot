package workflowstep

import (
	"context"
	"encoding/json"

	"github.com/gptscript-ai/otto/pkg/expression"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) getInput(ctx context.Context, client kclient.Client, step *v1.WorkflowStep) (string, error) {
	if len(step.Spec.Step.Input.Args) > 0 {
		resultMap := map[string]any{}
		for key, value := range step.Spec.Step.Input.Args {
			result, err := expression.Eval(ctx, client, step, value)
			if err != nil {
				return "", err
			}
			resultMap[key] = result
		}
		x, err := json.Marshal(resultMap)
		return string(x), err
	}

	return expression.EvalString(ctx, client, step, step.Spec.Step.Input.Content)
}

func (h *Handler) getCondition(ctx context.Context, client kclient.Client, step *v1.WorkflowStep) (bool, error) {
	if step.Spec.Step.If != nil {
		return expression.EvalBool(ctx, client, step, step.Spec.Step.If.Condition)
	}
	if step.Spec.Step.While != nil {
		return expression.EvalBool(ctx, client, step, step.Spec.Step.While.Condition)
	}
	return false, nil
}

func (h *Handler) getItems(ctx context.Context, client kclient.Client, step *v1.WorkflowStep) ([]any, error) {
	return expression.EvalArray(ctx, client, step, step.Spec.Step.ForEach.Items)
}
