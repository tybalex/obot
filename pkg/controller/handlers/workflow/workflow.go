package workflow

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/mvl"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

var log = mvl.Package()

type Handler struct {
	WorkspaceClient   *wclient.Client
	WorkspaceProvider string
}

func (h *Handler) CreateWorkspace(req router.Request, resp router.Response) error {
	ws := req.Object.(*v1.Workflow)
	if ws.Status.WorkspaceID != "" {
		return nil
	}

	w, err := h.WorkspaceClient.Create(req.Ctx, h.WorkspaceProvider)
	if err != nil {
		return err
	}

	ws.Status.WorkspaceID = w
	if err := req.Client.Status().Update(req.Ctx, ws); err != nil {
		// Delete workspace since we failed to update the workflow
		if err := h.WorkspaceClient.Rm(req.Ctx, w); err != nil {
			log.Errorf("failed to delete workspace %s: %v", w, err)
		}
		return err
	}

	return nil
}

func (h *Handler) Finalize(req router.Request, resp router.Response) error {
	wf := req.Object.(*v1.Workflow)
	if wf.Status.WorkspaceID == "" {
		return nil
	}
	return h.WorkspaceClient.Rm(req.Ctx, wf.Status.WorkspaceID)
}
