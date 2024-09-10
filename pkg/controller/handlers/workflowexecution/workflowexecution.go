package workflowexecution

import (
	"context"
	"fmt"

	"github.com/acorn-io/baaah/pkg/name"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/mvl"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var log = mvl.Package()

type Handler struct {
	WorkspaceClient *wclient.Client
}

func (h *Handler) Finalize(req router.Request, resp router.Response) error {
	we := req.Object.(*v1.WorkflowExecution)
	if we.Status.WorkspaceID != "" {
		if err := h.WorkspaceClient.Rm(req.Ctx, we.Status.WorkspaceID); err != nil {
			return err
		}
		we.Status.WorkspaceID = ""
	}

	return nil
}

func (h *Handler) Cleanup(req router.Request, resp router.Response) error {
	var (
		we       = req.Object.(*v1.WorkflowExecution)
		workflow v1.Workflow
	)

	if err := req.Get(&workflow, we.Namespace, we.Spec.WorkflowName); apierror.IsNotFound(err) {
		return req.Delete(we)
	} else if err != nil {
		return err
	}

	return nil
}

func (h *Handler) Run(req router.Request, resp router.Response) error {
	we := req.Object.(*v1.WorkflowExecution)

	switch we.Status.State {
	case v1.WorkflowStateError, v1.WorkflowStateComplete:
		resp.DisablePrune()
		return nil
	}

	if we.Status.State != v1.WorkflowStateRunning {
		we.Status.State = v1.WorkflowStateRunning
		if err := req.Client.Status().Update(req.Ctx, we); err != nil {
			return err
		}
	}

	we.Status.StatusMessage = ""

	if we.Status.WorkflowManifest == nil {
		if err := h.loadManifest(req, we); err != nil {
			return err
		}
	}

	if we.Status.WorkspaceID == "" {
		if ok, err := h.createWorkspace(req.Ctx, req.Client, we); err != nil {
			return err
		} else if !ok {
			return nil
		}
	}

	var (
		steps        []kclient.Object
		lastStepName string
	)

	for i, step := range we.Status.WorkflowManifest.Steps {
		name := name.SafeHashConcatName(we.Name, fmt.Sprint(i))
		steps = append(steps, &v1.WorkflowStep{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: we.Namespace,
			},
			Spec: v1.WorkflowStepSpec{
				AfterWorkflowStepName: lastStepName,
				Step:                  step,
				Path:                  []string{fmt.Sprint(i)},
				WorkflowName:          we.Spec.WorkflowName,
				WorkflowExecutionName: we.Name,
				WorkspaceID:           we.Status.WorkspaceID,
			},
		})

		lastStepName = name
	}

	newState, err := h.getState(req.Ctx, req.Client, steps)
	if err != nil {
		return err
	}

	we.Status.State = newState
	resp.Objects(steps...)
	return nil
}

func (h *Handler) getState(ctx context.Context, client kclient.Client, steps []kclient.Object) (v1.WorkflowState, error) {
	for i, obj := range steps {
		step := obj.(*v1.WorkflowStep).DeepCopy()
		if err := client.Get(ctx, router.Key(step.Namespace, step.Name), step); apierror.IsNotFound(err) {
			return v1.WorkflowStateRunning, nil
		} else if err != nil {
			return "", err
		}
		if step.Status.State == v1.WorkflowStepStateError {
			return v1.WorkflowStateError, nil
		}
		if len(steps)-1 == i && step.Status.State == v1.WorkflowStepStateComplete {
			return v1.WorkflowStateComplete, nil
		}
	}

	return v1.WorkflowStateRunning, nil
}

func (h *Handler) createWorkspace(ctx context.Context, client kclient.Client, we *v1.WorkflowExecution) (bool, error) {
	var workspace v1.Workflow
	if err := client.Get(ctx, router.Key(we.Namespace, we.Spec.WorkflowName), &workspace); err != nil {
		return false, err
	}

	if workspace.Status.WorkspaceID == "" {
		we.Status.StatusMessage = "Waiting for workflow workspace to be created"
		return false, nil
	}

	workspaceID, err := h.WorkspaceClient.Create(ctx, "directory", workspace.Status.WorkspaceID)
	if err != nil {
		return false, err
	}
	we.Status.WorkspaceID = workspaceID
	if err := client.Status().Update(ctx, we); err != nil {
		// Delete workspace since we failed to update the workflow
		if err := h.WorkspaceClient.Rm(ctx, workspaceID); err != nil {
			log.Errorf("failed to delete workspace %s: %v", workspaceID, err)
		}
	}

	return true, nil
}

func (h *Handler) loadManifest(req router.Request, we *v1.WorkflowExecution) error {
	var wf v1.Workflow
	if err := req.Get(&wf, we.Namespace, we.Spec.WorkflowName); err != nil {
		return err
	}
	we.Status.WorkflowManifest = &wf.Spec.Manifest
	return req.Client.Status().Update(req.Ctx, we)
}
