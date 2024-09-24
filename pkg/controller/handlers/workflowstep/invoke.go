package workflowstep

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/invoke"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func (h *Handler) RunInvoke(req router.Request, resp router.Response) error {
	var (
		ctx         = req.Ctx
		client      = req.Client
		step        = req.Object.(*v1.WorkflowStep)
		lastRunName string
	)

	if step.Spec.Step.If != nil || step.Spec.Step.While != nil {
		return nil
	}

	if step.Spec.AfterWorkflowStepName != "" {
		var previousStep v1.WorkflowStep
		if err := client.Get(ctx, router.Key(step.Namespace, step.Spec.AfterWorkflowStepName), &previousStep); err != nil {
			return err
		}
		if previousStep.Status.LastRunName == "" {
			return fmt.Errorf("previous step %s has no last run name", previousStep.Name)
		}
		lastRunName = previousStep.Status.LastRunName
	}

	var run v1.Run
	if len(step.Status.RunNames) == 0 {
		invokeResp, err := h.invoker.Step(ctx, req.Client, step, invoke.StepOptions{
			PreviousRunName: lastRunName,
		})
		if err != nil {
			return err
		}

		step.Status.ThreadName = invokeResp.Thread.Name
		step.Status.RunNames = []string{invokeResp.Run.Name}

		// Ignored error updating
		_ = client.Status().Update(ctx, step)

		run = *invokeResp.Run
	} else {
		if err := req.Get(&run, step.Namespace, step.Status.RunNames[0]); err != nil {
			return err
		}
	}

	switch run.Status.State {
	case gptscript.Continue, gptscript.Finished:
		if subCall, ok := h.toSubCall(run.Status.Output); ok {
			step.Status.State = v1.WorkflowStepStateSubCall
			step.Status.SubCalls = []v1.SubCall{subCall}
		} else {
			step.Status.State = v1.WorkflowStepStateComplete
			step.Status.LastRunName = step.Status.RunNames[0]
			step.Status.SubCalls = nil
		}
		step.Status.Error = ""
	case gptscript.Error:
		step.Status.State = v1.WorkflowStepStateError
		step.Status.LastRunName = step.Status.RunNames[0]
		step.Status.Error = run.Status.Error
	}

	return nil
}

type call struct {
	Type     string `json:"type,omitempty"`
	Workflow string `json:"workflow,omitempty"`
	Input    any    `json:"input,omitempty"`
}

func (h *Handler) toSubCall(output string) (v1.SubCall, bool) {
	var call call
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &call); err != nil || call.Type != "OttoSubFlow" || call.Workflow == "" {
		return v1.SubCall{}, false
	}

	var inputString string
	switch v := call.Input.(type) {
	case string:
		inputString = v
	default:
		inputBytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		inputString = string(inputBytes)
	}

	if inputString == "{}" {
		inputString = ""
	}

	return v1.SubCall{
		Type:     call.Type,
		Workflow: call.Workflow,
		Input:    inputString,
	}, true
}
