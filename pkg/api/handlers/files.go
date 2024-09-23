package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/storage/selectors"
	"github.com/gptscript-ai/otto/pkg/workspace"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func listFiles(ctx context.Context, req api.Context, wc *wclient.Client, workspaceID string) error {
	files, err := wc.Ls(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %q: %w", workspaceID, err)
	}

	return req.Write(types.FileList{Items: files})
}

func listKnowledgeFiles(req api.Context, parentObj knowledge.Knowledgeable) error {
	if err := req.Get(parentObj, req.PathValue("id")); err != nil {
		return fmt.Errorf("failed to get the parent object: %w", err)
	}

	var files v1.KnowledgeFileList
	if err := req.Storage.List(req.Context(), &files, &client.ListOptions{
		FieldSelector: fields.SelectorFromSet(selectors.RemoveEmpty(map[string]string{
			"spec.agentName":    parentObj.AgentName(),
			"spec.workflowName": parentObj.WorkflowName(),
			"spec.threadName":   parentObj.ThreadName(),
		})),
		Namespace: parentObj.GetNamespace(),
	}); err != nil {
		return err
	}

	return req.Write(files)
}

func uploadKnowledge(req api.Context, workspaceClient *wclient.Client, parentName string, toUpdate knowledge.Knowledgeable) error {
	if err := req.Get(toUpdate, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", req.PathValue("id"), err)
	}

	status := toUpdate.KnowledgeWorkspaceStatus()
	if err := uploadFile(req.Context(), req, workspaceClient, status.KnowledgeWorkspaceID); err != nil {
		return err
	}

	filename := req.PathValue("file")
	file := &v1.KnowledgeFile{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1.ObjectNameFromAbsolutePath(
				filepath.Join(workspace.GetDir(status.KnowledgeWorkspaceID), filename),
			),
			Namespace: toUpdate.GetNamespace(),
		},
		Spec: v1.KnowledgeFileSpec{
			FileName:     filename,
			AgentName:    toUpdate.AgentName(),
			WorkflowName: toUpdate.WorkflowName(),
			ThreadName:   toUpdate.ThreadName(),
		},
	}

	if err := req.Storage.Create(req.Context(), file); err != nil && !apierrors.IsAlreadyExists(err) {
		_ = deleteFile(req.Context(), req, workspaceClient, status.KnowledgeWorkspaceID)
		return err
	}

	status.KnowledgeGeneration++
	status.HasKnowledge = true
	if err := req.Storage.Status().Update(req.Context(), toUpdate); err != nil {
		return err
	}

	return req.Write(file)
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

	return nil
}

func deleteKnowledge(req api.Context, filename, parentName string, toUpdate knowledge.Knowledgeable) error {
	if err := req.Get(toUpdate, parentName); err != nil {
		return fmt.Errorf("failed to get parent with id %s: %w", parentName, err)
	}

	fileObjectName := v1.ObjectNameFromAbsolutePath(
		filepath.Join(workspace.GetDir(toUpdate.KnowledgeWorkspaceStatus().KnowledgeWorkspaceID), filename),
	)

	if err := req.Storage.Delete(req.Context(), &v1.KnowledgeFile{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: toUpdate.GetNamespace(),
			Name:      fileObjectName,
		},
	}); err != nil {
		var apiErr *apierrors.StatusError
		if errors.As(err, &apiErr) {
			apiErr.ErrStatus.Details.Name = filename
			apiErr.ErrStatus.Message = strings.ReplaceAll(apiErr.ErrStatus.Message, fileObjectName, filename)
		}
		return err
	}

	req.WriteHeader(http.StatusNoContent)
	return nil
}

func deleteFile(ctx context.Context, req api.Context, wc *wclient.Client, workspaceID string) error {
	filename := req.PathValue("file")
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

	status := toUpdate.KnowledgeWorkspaceStatus()
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
