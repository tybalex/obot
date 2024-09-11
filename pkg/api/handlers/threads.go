package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

type ThreadHandler struct {
	WorkspaceClient *wclient.Client
}

func convertThread(thread v1.Thread) types.Thread {
	return types.Thread{
		ID:            thread.Name,
		Created:       thread.CreationTimestamp.Time,
		Description:   thread.Status.Description,
		AgentID:       thread.Spec.AgentName,
		Input:         thread.Spec.Input,
		LastRunName:   thread.Status.LastRunName,
		LastRunState:  thread.Status.LastRunState,
		LastRunOutput: thread.Status.LastRunOutput,
		LastRunError:  thread.Status.LastRunError,
	}
}

func (a *ThreadHandler) List(_ context.Context, req api.Request) error {
	var (
		agentName  = req.Request.PathValue("agent")
		threadList v1.ThreadList
	)
	if err := req.List(&threadList); err != nil {
		return err
	}

	var resp types.ThreadList
	for _, thread := range threadList.Items {
		if agentName == "" || thread.Spec.AgentName == agentName {
			resp.Items = append(resp.Items, convertThread(thread))
		}
	}

	return req.JSON(resp)
}
func (a *ThreadHandler) Files(ctx context.Context, req api.Request) error {
	var (
		id     = req.Request.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	return listFiles(ctx, req, a.WorkspaceClient, thread.Spec.WorkspaceID)
}

func (a *ThreadHandler) UploadFile(ctx context.Context, req api.Request) error {
	var (
		id     = req.Request.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	return uploadFile(ctx, req, a.WorkspaceClient, thread.Spec.WorkspaceID)
}

func (a *ThreadHandler) DeleteFile(ctx context.Context, req api.Request) error {
	var (
		id       = req.Request.PathValue("id")
		filename = req.Request.PathValue("file")
		thread   v1.Thread
	)

	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	return deleteFile(ctx, req, a.WorkspaceClient, thread.Spec.WorkspaceID, filename)
}

func (a *ThreadHandler) Knowledge(ctx context.Context, req api.Request) error {
	var (
		id     = req.Request.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	return listFiles(ctx, req, a.WorkspaceClient, thread.Spec.KnowledgeWorkspaceID)
}

func (a *ThreadHandler) UploadKnowledge(ctx context.Context, req api.Request) error {
	var (
		id     = req.Request.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	if err := uploadFile(ctx, req, a.WorkspaceClient, thread.Spec.KnowledgeWorkspaceID); err != nil {
		return err
	}

	thread.Status.KnowledgeGeneration++
	thread.Status.HasKnowledge = true
	return req.Storage.Status().Update(ctx, &thread)
}

func (a *ThreadHandler) DeleteKnowledge(ctx context.Context, req api.Request) error {
	var (
		id       = req.Request.PathValue("id")
		filename = req.Request.PathValue("file")
		thread   v1.Thread
	)

	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	if err := deleteFile(ctx, req, a.WorkspaceClient, thread.Spec.KnowledgeWorkspaceID, filename); err != nil {
		return err
	}

	files, err := a.WorkspaceClient.Ls(ctx, thread.Spec.KnowledgeWorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %s: %w", thread.Spec.KnowledgeWorkspaceID, err)
	}

	thread.Status.KnowledgeGeneration++
	thread.Status.HasKnowledge = len(files) > 0
	return req.Storage.Status().Update(ctx, &thread)
}

func (a *ThreadHandler) IngestKnowledge(ctx context.Context, req api.Request) error {
	var (
		id     = req.Request.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	files, err := a.WorkspaceClient.Ls(ctx, thread.Spec.KnowledgeWorkspaceID)
	if err != nil {
		return err
	}

	req.WriteHeader(http.StatusNoContent)

	if len(files) == 0 && !thread.Status.HasKnowledge {
		return nil
	}

	thread.Status.KnowledgeGeneration++
	return req.Storage.Status().Update(ctx, &thread)
}
