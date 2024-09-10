package workflowstep

import (
	"fmt"
	"slices"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) RunWhile(req router.Request, resp router.Response) error {
	step := req.Object.(*v1.WorkflowStep)

	if step.Spec.Step.While == nil {
		return nil
	}

	count := step.Spec.Step.While.MaxLoops
	if count <= 0 {
		count = 3
	}

	var finalState = v1.WorkflowStepStateComplete

	// Do one extra iteration to check the final state
	count++
	for i := 0; i < count; i++ {
		steps, err := h.defineWhile(i, step)
		if err != nil {
			return err
		}
		newState, err := getStateFromSteps(req.Ctx, req.Client, steps)
		if err != nil {
			return err
		}

		if newState == v1.WorkflowStepStateRunning || newState == v1.WorkflowStepStateError {
			finalState = newState
			resp.Objects(steps...)
			break
		}

		if newState == v1.WorkflowStepStatePending {
			if i == count-1 {
				finalState = v1.WorkflowStepStateError
				break
			}
			ok, err := h.getCondition(req.Ctx, req.Client, step)
			if err != nil {
				return err
			}
			if !ok {
				finalState = v1.WorkflowStepStateComplete
			} else if i > 0 {
				finalState = v1.WorkflowStepStateRunning
			} else {
				finalState = v1.WorkflowStepStatePending
			}
			resp.Objects(steps...)
			break
		}

		resp.Objects(steps...)
	}

	step.Status.State = finalState
	return nil
}

func (h *Handler) defineWhile(groupIndex int, step *v1.WorkflowStep) (result []kclient.Object, _ error) {
	steps := step.Spec.Step.While.Steps

	var (
		lastStepName string
	)

	for i, forStep := range steps {
		stepPath := append(step.Spec.Path, fmt.Sprint(groupIndex), fmt.Sprint(i))
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
				Step:                   forStep,
				Path:                   stepPath,
				GroupIndex:             &groupIndex,
				StepIndex:              &i,
				WorkflowName:           step.Spec.WorkflowName,
				WorkflowExecutionName:  step.Spec.WorkflowExecutionName,
				WorkspaceID:            step.Spec.WorkspaceID,
			},
		})
	}

	return result, nil
}
