package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
)

func listFiles(ctx context.Context, req api.Context, wc *wclient.Client, workspaceID string) error {
	files, err := wc.Ls(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %q: %w", workspaceID, err)
	}

	return req.Write(types.FileList{Items: files})
}

func uploadKnowledge(req api.Context, workspaceClient *wclient.Client, parentName string, toUpdate knowledge.Knowledgeable) error {
	if err := req.Get(toUpdate, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", req.PathValue("id"), err)
	}

	status := toUpdate.GetKnowledgeWorkspaceStatus()
	if err := uploadFile(req.Context(), req, workspaceClient, status.KnowledgeWorkspaceID); err != nil {
		return err
	}

	status.KnowledgeGeneration++
	status.HasKnowledge = true
	return req.Storage.Status().Update(req.Context(), toUpdate)
}

func uploadFile(ctx context.Context, req api.Context, wc *wclient.Client, workspaceID string) error {
	file := req.PathValue("file")
	if file == "" {
		return fmt.Errorf("file path parameter is required")
	}

	writer, err := wc.WriteFile(ctx, workspaceID, file)
	if err != nil {
		return fmt.Errorf("failed to upload file %q to workspace %q: %w", file, workspaceID, err)
	}

	_, err = io.Copy(writer, req.Request.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %q to workspace %q: %w", file, workspaceID, err)
	}

	req.WriteHeader(http.StatusCreated)

	return nil
}

func deleteKnowledge(req api.Context, workspaceClient *wclient.Client, filename, parentName string, toUpdate knowledge.Knowledgeable) error {
	if err := req.Get(toUpdate, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	status := toUpdate.GetKnowledgeWorkspaceStatus()
	if err := deleteFile(req.Context(), req, workspaceClient, status.KnowledgeWorkspaceID, filename); err != nil {
		return err
	}

	files, err := workspaceClient.Ls(req.Context(), status.KnowledgeWorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %s: %w", status.KnowledgeWorkspaceID, err)
	}

	status.KnowledgeGeneration++
	status.HasKnowledge = len(files) > 0
	return req.Storage.Status().Update(req.Context(), toUpdate)
}

func deleteFile(ctx context.Context, req api.Context, wc *wclient.Client, workspaceID, filename string) error {
	if err := wc.DeleteFile(ctx, workspaceID, filename); err != nil {
		return fmt.Errorf("failed to delete file %q from workspace %q: %w", filename, workspaceID, err)
	}

	req.WriteHeader(http.StatusNoContent)

	return nil
}

func ingestKnowledge(req api.Context, workspaceClient *wclient.Client, parentName string, toUpdate knowledge.Knowledgeable) error {
	if err := req.Get(toUpdate, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", req.PathValue("id"), err)
	}

	status := toUpdate.GetKnowledgeWorkspaceStatus()
	files, err := workspaceClient.Ls(req.Context(), status.KnowledgeWorkspaceID)
	if err != nil {
		return err
	}

	req.WriteHeader(http.StatusNoContent)

	if len(files) == 0 && !status.HasKnowledge {
		return nil
	}

	status.KnowledgeGeneration++
	return req.Storage.Status().Update(req.Context(), toUpdate)
}
