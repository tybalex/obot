package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	"github.com/gptscript-ai/otto/pkg/mvl"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	workspaceClient   *wclient.Client
	ingester          *knowledge.Ingester
	workspaceProvider string
}

func New(wc *wclient.Client, ingester *knowledge.Ingester, wp string) *Handler {
	return &Handler{
		workspaceClient:   wc,
		ingester:          ingester,
		workspaceProvider: wp,
	}
}

func (a *Handler) CreateWorkspace(req router.Request, resp router.Response) error {
	knowledged := req.Object.(knowledge.Knowledgeable)
	status := knowledged.KnowledgeWorkspaceStatus()
	if status.KnowledgeWorkspaceID != "" {
		return nil
	}

	knowledgeWorkspaceID, err := a.workspaceClient.Create(req.Ctx, a.workspaceProvider)
	if err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, knowledgeWorkspaceID)
		return err
	}

	status.KnowledgeWorkspaceID = knowledgeWorkspaceID

	if err := req.Client.Status().Update(req.Ctx, knowledged); err != nil {
		_ = a.workspaceClient.Rm(req.Ctx, knowledgeWorkspaceID)
		return err
	}

	return nil
}

func (a *Handler) RemoveWorkspace(req router.Request, _ router.Response) error {
	knowledged := req.Object.(knowledge.Knowledgeable)
	status := knowledged.KnowledgeWorkspaceStatus()

	if status.HasKnowledge {
		run, err := a.ingester.DeleteKnowledge(req.Ctx, knowledged.AgentName(), knowledged.GetNamespace(), status.KnowledgeWorkspaceID)
		if err != nil {
			return err
		}

		run.Wait()
		if run.Run.Status.Error != "" {
			return fmt.Errorf("failed to delete knowledge: %s", run.Run.Status.Error)
		}
	}

	if status.KnowledgeWorkspaceID != "" {
		return a.workspaceClient.Rm(req.Ctx, status.KnowledgeWorkspaceID)
	}

	return nil
}

func (a *Handler) IngestKnowledge(req router.Request, _ router.Response) error {
	knowledged := req.Object.(knowledge.Knowledgeable)
	status := knowledged.KnowledgeWorkspaceStatus()
	if status.KnowledgeGeneration == status.ObservedKnowledgeGeneration || status.IngestionRunName != "" {
		// If the RunName is set, then there is an ingestion in progress.
		// Wait for it to complete before starting another.
		return nil
	}

	var (
		run *invoke.Response
		err error
	)

	run, err = a.ingester.IngestKnowledge(req.Ctx, knowledged.AgentName(), knowledged.GetNamespace(), status.KnowledgeWorkspaceID)
	if err != nil {
		return err
	}

	go compileFileStatuses(req.Ctx, req.Client, knowledged, run, mvl.Package())

	status.ObservedKnowledgeGeneration = status.KnowledgeGeneration
	status.IngestionRunName = run.Run.Name
	return nil
}

func compileFileStatuses(ctx context.Context, client kclient.Client, knowledged knowledge.Knowledgeable, run *invoke.Response, logger mvl.Logger) {
	for e := range run.Events {
		for _, line := range strings.Split(e.Content, "\n") {
			if line == "" || line[0] != '{' {
				continue
			}
			var ingestionStatus v1.IngestionStatus
			if err := json.Unmarshal([]byte(line), &ingestionStatus); err != nil {
				logger.Errorf("failed to unmarshal event: %s", err)
			}

			if ingestionStatus.Filepath == "" {
				// Not a file status log.
				continue
			}

			var file v1.KnowledgeFile
			if err := client.Get(ctx, router.Key(knowledged.GetNamespace(), v1.ObjectNameFromAbsolutePath(ingestionStatus.Filepath)), &file); err != nil {
				logger.Errorf("failed to get file: %s", err)
				continue
			}

			if err := json.Unmarshal([]byte(line), &file.Status.IngestionStatus); err != nil {
				logger.Errorf("failed to into file status: %s", err)
			}

			if err := client.Status().Update(ctx, &file); err != nil {
				logger.Errorf("failed to update file: %s", err)
			}
		}
	}

	if err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return !apierrors.IsNotFound(err)
	}, func() error {
		if err := client.Get(ctx, router.Key(knowledged.GetNamespace(), knowledged.GetName()), knowledged); err != nil {
			return err
		}

		newStatus := knowledged.KnowledgeWorkspaceStatus()
		newStatus.IngestionRunName = ""
		return client.Status().Update(ctx, knowledged)
	}); err != nil {
		logger.Errorf("failed to update status: %s", err)
	}
}

func (a *Handler) GCFile(req router.Request, _ router.Response) error {
	kFile := req.Object.(*v1.KnowledgeFile)

	if kFile.Spec.UploadName != "" {
		var upload v1.OneDriveLinks
		if err := req.Get(&upload, kFile.Namespace, kFile.Spec.UploadName); apierrors.IsNotFound(err) || !upload.GetDeletionTimestamp().IsZero() {
			return kclient.IgnoreNotFound(req.Delete(kFile))
		} else if err != nil {
			return err
		}
	}

	if parent, err := knowledgeFileParent(req.Ctx, req.Client, kFile); apierrors.IsNotFound(err) || !parent.GetDeletionTimestamp().IsZero() {
		return kclient.IgnoreNotFound(req.Delete(kFile))
	} else if err != nil {
		return err
	}

	return nil
}

func (a *Handler) CleanupFile(req router.Request, _ router.Response) error {
	kFile := req.Object.(*v1.KnowledgeFile)

	parent, err := knowledgeFileParent(req.Ctx, req.Client, kFile)
	if apierrors.IsNotFound(err) {
		// If the parent object is not found, then the workspaces will be deleted and nothing needs to happen here.
		return nil
	}
	if err != nil {
		return err
	}

	status := parent.KnowledgeWorkspaceStatus()

	if err = a.workspaceClient.DeleteFile(req.Ctx, status.KnowledgeWorkspaceID, kFile.Spec.FileName); err != nil {
		if errors.As(err, new(wclient.FileNotFoundError)) {
			// It is important to return nil here and not move forward because when bulk deleting files from a remote provider
			// (like OneDrive), the connector will remove the files from the local disk and the controller will remove the
			// KnowledgeFile objects. We don't want to kick off (possibly) numerous ingestion runs.
			return nil
		}
		return err
	}

	files, err := a.workspaceClient.Ls(req.Ctx, status.KnowledgeWorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %s: %w", status.KnowledgeWorkspaceID, err)
	}

	status.KnowledgeGeneration++
	status.HasKnowledge = len(files) > 0
	return req.Client.Status().Update(req.Ctx, parent)
}

func knowledgeFileParent(ctx context.Context, client kclient.Client, kFile *v1.KnowledgeFile) (knowledge.Knowledgeable, error) {
	switch {
	case kFile.Spec.ThreadName != "":
		var thread v1.Thread
		return &thread, client.Get(ctx, router.Key(kFile.Namespace, kFile.Spec.ThreadName), &thread)

	case kFile.Spec.AgentName != "":
		var agent v1.Agent
		return &agent, client.Get(ctx, router.Key(kFile.Namespace, kFile.Spec.AgentName), &agent)

	case kFile.Spec.WorkflowName != "":
		var workflow v1.Workflow
		return &workflow, client.Get(ctx, router.Key(kFile.Namespace, kFile.Spec.WorkflowName), &workflow)
	}

	return nil, fmt.Errorf("unable to find parent for knowledge file %s", kFile.Name)
}
