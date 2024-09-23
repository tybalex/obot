package workspace

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	workspaceClient   *wclient.Client
	workspaceProvider string
}

func New(wc *wclient.Client, wp string) *Handler {
	return &Handler{
		workspaceClient:   wc,
		workspaceProvider: wp,
	}
}

type workspaceable interface {
	kclient.Object
	WorkspaceStatus() *v1.WorkspaceStatus
}

func (a *Handler) CreateWorkspace(req router.Request, _ router.Response) error {
	workspaced := req.Object.(workspaceable)
	status := workspaced.WorkspaceStatus()
	if status.WorkspaceID != "" {
		return nil
	}

	workspaceID, err := a.workspaceClient.Create(req.Ctx, a.workspaceProvider)
	if err != nil {
		return err
	}

	status.WorkspaceID = workspaceID

	if err := req.Client.Status().Update(req.Ctx, workspaced); err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, workspaceID)
		return err
	}

	return nil
}

func (a *Handler) RemoveWorkspace(req router.Request, _ router.Response) error {
	workspaced := req.Object.(workspaceable)
	status := workspaced.WorkspaceStatus()
	if status.WorkspaceID != "" {
		if err := a.workspaceClient.Rm(req.Ctx, status.WorkspaceID); err != nil {
			return err
		}
	}

	return nil
}
