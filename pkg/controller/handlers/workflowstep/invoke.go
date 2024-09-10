package workflowstep

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func (h *Handler) RunInvoke(req router.Request, resp router.Response) error {
	var (
		ctx    = req.Ctx
		client = req.Client
		step   = req.Object.(*v1.WorkflowStep)
	)

	if step.Spec.Step.If != nil || step.Spec.Step.ForEach != nil || step.Spec.Step.While != nil {
		return nil
	}

	input, err := h.getInput(ctx, client, step)
	if err != nil {
		return err
	}

	invokeResp, err := h.Invoker.Step(ctx, step, input)
	if err != nil {
		return err
	}

	step.Status.ThreadName = invokeResp.Thread.Name
	step.Status.LastRunName = invokeResp.Run.Name

	// Ignored error updating
	_ = client.Status().Update(ctx, step)
	invokeResp.Wait()

	// reload run
	if err := client.Get(ctx, router.Key(invokeResp.Run.Namespace, invokeResp.Run.Name), invokeResp.Run); err != nil {
		return err
	}

	if invokeResp.Run.Status.Error == "" {
		step.Status.State = v1.WorkflowStepStateComplete
	} else {
		step.Status.State = v1.WorkflowStepStateError
		step.Status.Error = invokeResp.Run.Status.Error
	}

	return client.Status().Update(ctx, step)
}
