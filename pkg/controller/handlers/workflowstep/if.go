package workflowstep

import (
	"context"
	"fmt"
	"slices"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) RunIf(req router.Request, resp router.Response) error {
	step := req.Object.(*v1.WorkflowStep)

	if step.Spec.Step.If == nil {
		return nil
	}

	steps, err := h.defineIf(req.Ctx, req.Client, step)
	if err != nil {
		return err
	}

	newState, err := getStateFromSteps(req.Ctx, req.Client, steps)
	if err != nil {
		return err
	}

	step.Status.State = newState
	resp.Objects(steps...)
	return nil
}

func (h *Handler) defineIf(ctx context.Context, client kclient.Client, step *v1.WorkflowStep) (result []kclient.Object, _ error) {
	var steps []v1.Step
	if ok, err := h.getCondition(ctx, client, step); err != nil {
		return nil, err
	} else if ok {
		steps = step.Spec.Step.If.Steps
	} else {
		steps = step.Spec.Step.If.Else
	}

	var lastStepName string
	for i, ifStep := range steps {
		stepPath := append(step.Spec.Path, fmt.Sprint(i))
		stepName := name.SafeHashConcatName(slices.Concat([]string{step.Spec.WorkflowExecutionName}, stepPath)...)
		afterStepName := step.Spec.AfterWorkflowStepName
		if i > 0 {
			afterStepName = lastStepName
		}
		result = append(result, &v1.WorkflowStep{
			ObjectMeta: metav1.ObjectMeta{
				Name:      stepName,
				Namespace: step.Namespace,
			},
			Spec: v1.WorkflowStepSpec{
				ParentWorkflowStepName: step.Name,
				AfterWorkflowStepName:  afterStepName,
				Step:                   ifStep,
				Path:                   stepPath,
				StepIndex:              &i,
				WorkflowName:           step.Spec.WorkflowName,
				WorkflowExecutionName:  step.Spec.WorkflowExecutionName,
				WorkspaceID:            step.Spec.WorkspaceID,
			},
		})
	}

	return result, nil
}
