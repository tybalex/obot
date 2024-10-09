package knowledge

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.gptscript.ai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *Handler) CleanupIngestRequests(req router.Request, _ router.Response) error {
	reIngestRequest := req.Object.(*v1.IngestKnowledgeRequest)

	var ws v1.Workspace
	// Use !Before here for checking time because the lastIngestionRunStarted time and the ingestion request maybe the at the same second.
	if err := req.Get(&ws, reIngestRequest.Namespace, reIngestRequest.Spec.WorkspaceName); apierrors.IsNotFound(err) || err == nil && !ws.Status.LastIngestionRunStarted.Before(&reIngestRequest.CreationTimestamp) {
		return kclient.IgnoreNotFound(req.Client.Delete(req.Ctx, reIngestRequest))
	} else if err != nil {
		return err
	}

	return nil
}
