package threads

import (
	"fmt"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/aihelper"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

type ThreadHandler struct {
	workspace *wclient.Client
	ingester  *knowledge.Ingester
	aihelper  *aihelper.AIHelper
}

func New(workspace *wclient.Client, ingester *knowledge.Ingester, aihelper *aihelper.AIHelper) *ThreadHandler {
	return &ThreadHandler{
		workspace: workspace,
		ingester:  ingester,
		aihelper:  aihelper,
	}
}

func (t *ThreadHandler) MoveWorkspacesToStatus(req router.Request, _ router.Response) error {
	thread := req.Object.(*v1.Thread)
	if thread.Status.Workspace.WorkspaceID == "" || thread.Status.KnowledgeWorkspace.KnowledgeWorkspaceID == "" {
		thread.Status.Workspace.WorkspaceID = thread.Spec.WorkspaceID
		thread.Status.KnowledgeWorkspace.KnowledgeWorkspaceID = thread.Spec.KnowledgeWorkspaceID
	}

	return nil
}

// HasKnowledge is a dumb optimazation to avoid updating the status of the thread because when you read
// the knowledge status of the thread, we copy it from the spec to the status. This causes a write when there
// before this knowledge, and we just don't need that.
func (t *ThreadHandler) HasKnowledge(handler router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		if req.Object == nil {
			return nil
		}
		thread := req.Object.(*v1.Thread)
		if thread.Status.KnowledgeWorkspace.HasKnowledge {
			return handler.Handle(req, resp)
		}
		return nil
	})
}

func (t *ThreadHandler) Description(req router.Request, _ router.Response) error {
	thread := req.Object.(*v1.Thread)

	if thread.Spec.Manifest.Description != "" || thread.Status.LastRunName == "" {
		return nil
	}

	var run v1.Run
	if err := req.Get(&run, thread.Namespace, thread.Status.LastRunName); err != nil {
		return err
	}

	for run.Spec.PreviousRunName != "" {
		var prevRun v1.Run
		if err := req.Get(&prevRun, thread.Namespace, run.Spec.PreviousRunName); err != nil {
			return err
		}
		if prevRun.Spec.ThreadName == thread.Name {
			run = prevRun
		} else {
			break
		}
	}

	var desc string
	err := t.aihelper.GenerateObject(req.Ctx, &desc,
		"Given the following start of a conversation, generate a short title of the conversation",
		fmt.Sprintf(`User: %s\n
Assistant: %s\n`, run.Spec.Input, run.Status.Output))
	if err != nil {
		return err
	}

	thread.Spec.Manifest.Description = desc
	return req.Client.Update(req.Ctx, thread)
}
