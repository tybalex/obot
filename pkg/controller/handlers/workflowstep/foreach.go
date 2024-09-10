package workflowstep

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/gz"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) RunForEach(req router.Request, resp router.Response) error {
	step := req.Object.(*v1.WorkflowStep)

	if step.Spec.Step.ForEach == nil {
		return nil
	}

	steps, err := h.defineFor(req.Ctx, req.Client, step)
	if err != nil {
		return err
	}

	var (
		newState v1.WorkflowStepState
		allSteps []kclient.Object
	)
	for _, itemSteps := range steps {
		allSteps = append(allSteps, itemSteps...)
		newState, err = getStateFromSteps(req.Ctx, req.Client, itemSteps)
		if err != nil {
			return err
		}
		if newState != v1.WorkflowStepStateComplete {
			break
		}
	}

	step.Status.State = newState
	resp.Objects(allSteps...)
	return nil
}
func (h *Handler) defineFor(ctx context.Context, client kclient.Client, step *v1.WorkflowStep) (result [][]kclient.Object, _ error) {
	items, err := h.getItems(ctx, client, step)
	if err != nil {
		return nil, err
	}
	steps := step.Spec.Step.ForEach.Steps

	for groupIndex := 0; groupIndex < len(items); groupIndex++ {
		var (
			lastStepName string
			itemSteps    []kclient.Object
			forItem      []byte
		)

		x, err := json.Marshal(items[groupIndex])
		if err != nil {
			return nil, err
		}
		forItem, err = gz.Compress(x)
		if err != nil {
			return nil, err
		}

		for i, forStep := range steps {
			stepPath := append(step.Spec.Path, fmt.Sprint(groupIndex), fmt.Sprint(i))
			stepName := name.SafeHashConcatName(slices.Concat([]string{step.Spec.WorkflowExecutionName}, stepPath)...)
			afterStepName := step.Spec.AfterWorkflowStepName
			if i > 0 {
				afterStepName = lastStepName
			}
			itemSteps = append(itemSteps, &v1.WorkflowStep{
				ObjectMeta: metav1.ObjectMeta{
					Name:      stepName,
					Namespace: step.Namespace,
				},
				Spec: v1.WorkflowStepSpec{
					ParentWorkflowStepName: step.Name,
					AfterWorkflowStepName:  afterStepName,
					Step:                   forStep,
					Path:                   stepPath,
					GroupIndex:             &groupIndex,
					StepIndex:              &i,
					ForItem:                forItem,
					WorkflowName:           step.Spec.WorkflowName,
					WorkflowExecutionName:  step.Spec.WorkflowExecutionName,
					WorkspaceID:            step.Spec.WorkspaceID,
				},
			})
		}

		result = append(result, itemSteps)
	}

	return result, nil
}
