package invoke

import (
	"context"

	"github.com/gptscript-ai/otto/pkg/render"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invoker) Step(ctx context.Context, step *v1.WorkflowStep, input string) (*Response, error) {
	thread := v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "t1",
			Namespace:    step.Namespace,
		},
		Spec: v1.ThreadSpec{
			WorkflowStepName: step.Name,
			Input:            input,
			WorkspaceID:      step.Spec.WorkspaceID,
		},
	}

	if err := i.storage.Create(ctx, &thread); err != nil {
		return nil, err
	}

	tools := render.Step(step)

	return i.createRunFromTools(ctx, &thread, tools, input, runOptions{
		WorkflowName:     step.Spec.WorkflowName,
		WorkflowStepName: step.Spec.AfterWorkflowStepName,
	})
}
