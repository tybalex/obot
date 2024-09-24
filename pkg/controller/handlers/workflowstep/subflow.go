package workflowstep

import (
	"fmt"
	"strings"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/render"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/system"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) RunSubflow(req router.Request, resp router.Response) error {
	step := req.Object.(*v1.WorkflowStep)

	if len(step.Status.SubCalls) > 0 {
		resp.DisablePrune()
	}

	if step.Status.State != v1.WorkflowStepStateSubCall {
		return nil
	}

	wfs, err := render.WorkflowByName(req.Ctx, req.Client, req.Namespace)
	if err != nil {
		return err
	}

	for i, subCall := range step.Status.SubCalls {
		if len(step.Status.RunNames) > i+1 {
			continue
		}

		wf, ok := wfs[subCall.Workflow]
		if !ok {
			return nil
		}

		wfe := &v1.WorkflowExecution{
			ObjectMeta: metav1.ObjectMeta{
				Name:       name.SafeConcatName(system.WorkflowExecutionPrefix+strings.TrimPrefix(step.Name, system.WorkflowStepPrefix), fmt.Sprintf("%d-%s", i, subCall.Workflow)),
				Namespace:  step.Namespace,
				Finalizers: []string{v1.WorkflowExecutionFinalizer},
			},
			Spec: v1.WorkflowExecutionSpec{
				Input:                 subCall.Input,
				WorkflowName:          wf.Name,
				AfterWorkflowStepName: step.Spec.AfterWorkflowStepName,
				WorkspaceID:           wf.Status.Workspace.WorkspaceID,
			},
		}
		resp.Objects(wfe)

		out, done, err := h.getSubflowOutput(req, wfe)
		if err != nil {
			return err
		}

		if !done {
			return nil
		}

		resp, err := h.invoker.Step(req.Ctx, req.Client, step, invoke.StepOptions{
			PreviousRunName: step.Status.RunNames[i],
			Continue:        &out,
		})
		if err != nil {
			return err
		}

		step.Status.RunNames = append(step.Status.RunNames, resp.Run.Name)
		return req.Client.Status().Update(req.Ctx, step)
	}

	nextRunName := step.Status.RunNames[len(step.Status.SubCalls)]

	var run v1.Run
	if err := req.Get(&run, step.Namespace, nextRunName); err != nil {
		return err
	}

	switch run.Status.State {
	case gptscript.Continue, gptscript.Finished:
		if subCall, ok := h.toSubCall(run.Status.Output); ok {
			step.Status.SubCalls = append(step.Status.SubCalls, subCall)
		} else {
			step.Status.State = v1.WorkflowStepStateComplete
			step.Status.LastRunName = nextRunName
			step.Status.Error = ""
		}
	case gptscript.Error:
		step.Status.State = v1.WorkflowStepStateError
		step.Status.LastRunName = nextRunName
	}

	return nil
}

func (h *Handler) getLastRun(req router.Request, step *v1.WorkflowStep) (string, error) {
	var check v1.WorkflowStep
	if err := req.Get(&check, step.Namespace, step.Name); apierrors.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	if check.Status.State != v1.WorkflowStepStateComplete {
		return "", nil
	}

	return check.Status.LastRunName, nil
}

func (h *Handler) getSubflowOutput(req router.Request, wfe *v1.WorkflowExecution) (string, bool, error) {
	var (
		check v1.WorkflowExecution
	)

	if err := req.Get(&check, wfe.Namespace, wfe.Name); apierrors.IsNotFound(err) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	if check.Status.State != v1.WorkflowStateComplete {
		return "", false, nil
	}

	return check.Status.Output, true, nil
}
