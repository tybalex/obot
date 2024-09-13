package threads

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type ThreadHandler struct {
	Workspace *wclient.Client
	ingester  *knowledge.Ingester
}

func New(wc *wclient.Client, ingester *knowledge.Ingester) *ThreadHandler {
	return &ThreadHandler{
		Workspace: wc,
		ingester:  ingester,
	}
}

func (t *ThreadHandler) Cleanup(req router.Request, resp router.Response) error {
	thread := req.Object.(*v1.Thread)

	if thread.Spec.AgentName != "" {
		var agent v1.Agent
		if err := req.Get(&agent, thread.Namespace, thread.Spec.AgentName); apierrors.IsNotFound(err) {
			return req.Delete(thread)
		} else if err != nil {
			return err
		}
	}

	if thread.Spec.WorkflowStepName != "" {
		var step v1.WorkflowStep
		if err := req.Get(&step, thread.Namespace, thread.Spec.WorkflowStepName); apierrors.IsNotFound(err) {
			return req.Delete(thread)
		} else if err != nil {
			return err
		}
	}

	return nil
}
