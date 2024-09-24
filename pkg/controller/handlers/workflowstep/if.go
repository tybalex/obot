package workflowstep

import (
	"context"
	"fmt"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) RunIf(req router.Request, resp router.Response) error {
	step := req.Object.(*v1.WorkflowStep)

	if step.Spec.Step.If == nil {
		return nil
	}

	conditionStep := h.defineCondition(step, nil, 0)
	resp.Objects(conditionStep)

	conditionRunName, conditionResult, wait, err := h.conditionResult(req.Ctx, req.Client, conditionStep)
	if err != nil {
		return err
	} else if wait {
		return nil
	}

	steps, err := h.defineIf(step, conditionStep, conditionResult)
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		step.Status.State = v1.WorkflowStepStateComplete
		step.Status.LastRunName = conditionRunName
		return nil
	}

	lastRunName, newState, err := getStateFromSteps(req.Ctx, req.Client, steps)
	if err != nil {
		return err
	}

	step.Status.State = newState
	step.Status.LastRunName = lastRunName
	resp.Objects(steps...)
	return nil
}

func (h *Handler) conditionResult(ctx context.Context, c kclient.Client, step *v1.WorkflowStep) (runName string, result, wait bool, err error) {
	var checkStep v1.WorkflowStep
	if err := c.Get(ctx, router.Key(step.Namespace, step.Name), &checkStep); apierrors.IsNotFound(err) {
		return "", false, true, nil
	} else if err != nil {
		return "", false, false, err
	}

	if checkStep.Status.State != v1.WorkflowStepStateComplete || checkStep.Status.LastRunName == "" {
		return "", false, true, nil
	}

	var run v1.Run
	if err := c.Get(ctx, router.Key(step.Namespace, checkStep.Status.LastRunName), &run); err != nil {
		return "", false, false, err
	}

	if isTrue(run.Status.Output) {
		return run.Name, true, false, nil
	} else if isFalse(run.Status.Output) {
		return run.Name, false, false, nil
	}

	return "", false, true, nil
}

func isTrue(s string) bool {
	check := truthyNormalize(s)
	return check == "true" ||
		check == "yes" ||
		check == "t" ||
		check == "y"
}

func truthyNormalize(s string) string {
	return strings.TrimSpace(strings.ToLower(strings.ReplaceAll(s, `"`, "")))
}

func isFalse(s string) bool {
	check := truthyNormalize(s)
	return check == "false" ||
		check == "no" ||
		check == "f" ||
		check == "n"
}

func toStepCondition(s string) string {
	//input := "STEP_CONDITION: " + step.Spec.Step.If.Condition
	input := `Response with only the word TRUE if the following condition is true, or FALSE if false:\n` + s
	return input
}

func (h *Handler) defineCondition(step, afterStep *v1.WorkflowStep, iteration int) *v1.WorkflowStep {
	afterStepName := step.Spec.AfterWorkflowStepName
	if afterStep != nil {
		afterStepName = afterStep.Name
	}

	var (
		condition = "false"
		suffix    string
	)
	if step.Spec.Step.If != nil {
		condition = step.Spec.Step.If.Condition
		suffix = "{condition}"
	} else if step.Spec.Step.While != nil {
		condition = step.Spec.Step.While.Condition
		suffix = fmt.Sprintf("{condition,index=%d}", iteration)
	}

	newStep := NewStep(step.Namespace, step.Spec.WorkflowExecutionName, afterStepName, v1.Step{
		ID:   step.Spec.Step.ID + suffix,
		Step: toStepCondition(condition),
	})
	return newStep
}

func (h *Handler) defineIf(step *v1.WorkflowStep, conditionStep *v1.WorkflowStep, conditionResult bool) (result []kclient.Object, _ error) {
	var steps []v1.Step
	if conditionResult {
		steps = step.Spec.Step.If.Steps
	} else {
		steps = step.Spec.Step.If.Else
	}

	var lastStepName string
	for i, ifStep := range steps {
		afterStepName := conditionStep.Name
		if i > 0 {
			afterStepName = lastStepName
		}
		newStep := NewStep(step.Namespace, step.Spec.WorkflowExecutionName, afterStepName, ifStep)
		result = append(result, newStep)
		lastStepName = newStep.Name
	}

	return result, nil
}
