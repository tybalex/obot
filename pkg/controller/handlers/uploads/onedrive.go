package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/acorn-io/baaah/pkg/uncached"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/system"
	"github.com/gptscript-ai/otto/pkg/workspace"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UploadHandler struct {
	invoker           *invoke.Invoker
	workspaceClient   *wclient.Client
	workspaceProvider string
	onedriveTool      string
}

func New(invoker *invoke.Invoker, wc *wclient.Client, workspaceProvider, onedriveTool string) *UploadHandler {
	return &UploadHandler{
		invoker:           invoker,
		workspaceClient:   wc,
		workspaceProvider: workspaceProvider,
		onedriveTool:      onedriveTool,
	}
}

func (u *UploadHandler) CreateThread(req router.Request, _ router.Response) error {
	oneDriveLinks := req.Object.(*v1.OneDriveLinks)
	if oneDriveLinks.Status.ThreadName != "" || oneDriveLinks.Generation == oneDriveLinks.Status.ObservedGeneration {
		return nil
	}

	id, err := u.workspaceClient.Create(req.Ctx, u.workspaceProvider)
	if err != nil {
		return fmt.Errorf("failed to create upload workspace: %w", err)
	}

	thread := &v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: system.ThreadPrefix,
			Namespace:    req.Namespace,
			Labels: map[string]string{
				v1.OneDriveLinksLabel: oneDriveLinks.Name,
			},
		},
		Spec: v1.ThreadSpec{
			AgentName:    oneDriveLinks.Spec.AgentName,
			WorkflowName: oneDriveLinks.Spec.WorkflowName,
			WorkspaceID:  id,
		},
	}
	if err = req.Client.Create(req.Ctx, thread); err != nil {
		_ = u.workspaceClient.Rm(req.Ctx, id)
		return err
	}

	oneDriveLinks.Status.ThreadName = thread.Name

	return nil
}

func (u *UploadHandler) RunUpload(req router.Request, _ router.Response) error {
	oneDriveLinks := req.Object.(*v1.OneDriveLinks)

	if oneDriveLinks.Status.ThreadName == "" || oneDriveLinks.Status.RunName != "" || oneDriveLinks.Generation == oneDriveLinks.Status.ObservedGeneration {
		return nil
	}

	var thread v1.Thread
	if err := req.Get(&thread, oneDriveLinks.Namespace, oneDriveLinks.Status.ThreadName); apierrors.IsNotFound(err) {
		// Might not be in the cache yet.
		return nil
	} else if err != nil {
		return err
	}

	files, err := knowledgeFilesForUploadName(req, oneDriveLinks.Namespace, oneDriveLinks.Name)
	if err != nil {
		return err
	}

	output := map[string]any{
		"files":   compileKnowledgeFilesForOneDriveConnector(files),
		"folders": oneDriveLinks.Status.Folders,
		"status":  oneDriveLinks.Status.Status,
		"error":   oneDriveLinks.Status.Error,
	}

	writer, err := u.workspaceClient.WriteFile(req.Ctx, thread.Spec.WorkspaceID, ".metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	_, status, err := parentObjAndStatus(req, oneDriveLinks)
	if err != nil {
		return fmt.Errorf("failed to get parent status: %w", err)
	}

	b, err := json.Marshal(map[string]any{
		"input": map[string]any{
			"sharedLinks": oneDriveLinks.Spec.SharedLinks,
			"outputDir":   filepath.Join(workspace.GetDir(status.KnowledgeWorkspaceID), oneDriveLinks.Name),
		},
		"output": output,
	})

	if _, err = writer.Write(b); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close metadata file: %w", err)
	}

	r, err := u.invoker.SystemActionWithThread(req.Ctx, &thread, u.onedriveTool, "{}")
	if err != nil {
		return err
	}

	go func() {
		// Don't care about the events here, but we need to pull them out
		r.Wait()
	}()

	oneDriveLinks.Status.RunName = r.Run.Name
	return nil
}

func (u *UploadHandler) HandleUploadRun(req router.Request, _ router.Response) error {
	oneDriveLinks := req.Object.(*v1.OneDriveLinks)

	if oneDriveLinks.Status.RunName == "" || oneDriveLinks.Generation == oneDriveLinks.Status.ObservedGeneration {
		return nil
	}

	var run v1.Run
	if err := req.Get(&run, oneDriveLinks.Namespace, oneDriveLinks.Status.RunName); apierrors.IsNotFound(err) {
		// Might not be in the cache yet.
		return nil
	} else if err != nil || !run.Status.State.IsTerminal() {
		return err
	}

	var thread v1.Thread
	if err := req.Get(&thread, oneDriveLinks.Namespace, oneDriveLinks.Status.ThreadName); err != nil {
		return err
	}

	parentObj, status, err := parentObjAndStatus(req, oneDriveLinks)
	if err != nil {
		return err
	}

	// Read the output metadata from the connector tool.
	file, err := u.workspaceClient.OpenFile(req.Ctx, thread.Spec.WorkspaceID, ".metadata.json")
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}

	var output map[string]v1.OneDriveLinksConnectorStatus
	if err = json.NewDecoder(file).Decode(&output); err != nil {
		return fmt.Errorf("failed to decode metadata file: %w", err)
	}

	if err = file.Close(); err != nil {
		return fmt.Errorf("failed to close metadata file: %w", err)
	}

	fileMetadata, knowledgeFileNamesFromOutput, err := compileKnowledgeFilesFromOneDriveConnector(req.Ctx, req.Client, oneDriveLinks, output["output"].Files, status)
	if err != nil {
		return err
	}

	if err = deleteKnowledgeFilesNotIncluded(req.Ctx, req.Client, oneDriveLinks.Namespace, oneDriveLinks.Name, knowledgeFileNamesFromOutput); err != nil {
		return err
	}

	// Put the metadata file in the agent knowledge workspace
	writer, err := u.workspaceClient.WriteFile(req.Ctx, status.KnowledgeWorkspaceID, filepath.Join(oneDriveLinks.Name, ".knowledge.json"))
	if err != nil {
		return fmt.Errorf("failed to create knowledge metadata file: %w", err)
	}
	if err = json.NewEncoder(writer).Encode(map[string]any{"metadata": fileMetadata}); err != nil {
		return fmt.Errorf("failed to encode metadata file: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close metadata file: %w", err)
	}

	// Reset the agent knowledge generation so that ingestion kicks off again.
	status.KnowledgeGeneration++
	if err = req.Client.Status().Update(req.Ctx, parentObj); err != nil {
		return fmt.Errorf("failed to update agent observed knowledge generation: %w", err)
	}

	oneDriveLinks.Status.Error = output["output"].Error
	oneDriveLinks.Status.Status = output["output"].Status
	oneDriveLinks.Status.Folders = output["output"].Folders

	// Reset thread name and observed generation so future ingests will create a new thread
	oneDriveLinks.Status.ThreadName = ""
	oneDriveLinks.Status.RunName = ""
	oneDriveLinks.Status.ObservedGeneration = oneDriveLinks.Generation

	return nil
}

func (u *UploadHandler) Cleanup(req router.Request, _ router.Response) error {
	onedriveLinks := req.Object.(*v1.OneDriveLinks)
	parentObj, status, err := parentObjAndStatus(req, onedriveLinks)
	if apierrors.IsNotFound(err) {
		// If the parent object is gone, then other handlers will ensure things are cleaned up.
		return nil
	} else if err != nil {
		return err
	}

	// Delete the threads associated with the onedrive links. The runs will be cleaned up by their cleanup handler.
	var threads v1.ThreadList
	if err := req.List(&threads, &client.ListOptions{
		Namespace: onedriveLinks.Namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			v1.OneDriveLinksLabel: onedriveLinks.Name,
		}),
	}); err != nil {
		return err
	}

	for _, thread := range threads.Items {
		if err := client.IgnoreNotFound(req.Delete(&thread)); err != nil {
			return fmt.Errorf("failed to delete thread %q: %w", thread.Name, err)
		}
	}

	// Delete the directory in the workspace for this onedrive link.
	if err := u.workspaceClient.RmDir(req.Ctx, status.KnowledgeWorkspaceID, onedriveLinks.Name); err != nil {
		return fmt.Errorf("failed to delete directory %q in workspace %s: %w", onedriveLinks.Name, status.KnowledgeWorkspaceID, err)
	}

	files, err := u.workspaceClient.Ls(req.Ctx, status.KnowledgeWorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to list files in workspace %s: %w", status.KnowledgeWorkspaceID, err)
	}

	// Reset the agent knowledge generation so that ingestion kicks off again.
	status.KnowledgeGeneration++
	status.HasKnowledge = len(files) > 0
	if err = req.Client.Status().Update(req.Ctx, parentObj); err != nil {
		return fmt.Errorf("failed to update agent observed knowledge generation: %w", err)
	}

	return nil
}

func (u *UploadHandler) GC(req router.Request, _ router.Response) error {
	onedriveLinks := req.Object.(*v1.OneDriveLinks)

	if onedriveLinks.Spec.AgentName != "" {
		var agent v1.Agent
		err := req.Get(&agent, onedriveLinks.Namespace, onedriveLinks.Spec.AgentName)
		if apierrors.IsNotFound(err) {
			return client.IgnoreNotFound(req.Delete(onedriveLinks))
		}

		return err
	} else if onedriveLinks.Spec.WorkflowName != "" {
		var workflow v1.Workflow
		err := req.Get(&workflow, onedriveLinks.Namespace, onedriveLinks.Spec.WorkflowName)
		if apierrors.IsNotFound(err) {
			return client.IgnoreNotFound(req.Delete(onedriveLinks))
		}

		return err
	}

	return nil
}

func parentObjAndStatus(req router.Request, onedriveLinks *v1.OneDriveLinks) (knowledge.Knowledgeable, *v1.KnowledgeWorkspaceStatus, error) {
	if onedriveLinks.Spec.AgentName != "" {
		var agent v1.Agent
		if err := req.Get(&agent, onedriveLinks.Namespace, onedriveLinks.Spec.AgentName); err != nil {
			return nil, nil, err
		}

		return &agent, &agent.Status.KnowledgeWorkspace, nil
	} else if onedriveLinks.Spec.WorkflowName != "" {
		var workflow v1.Workflow
		if err := req.Get(&workflow, onedriveLinks.Namespace, onedriveLinks.Spec.WorkflowName); err != nil {
			return nil, nil, err
		}

		return &workflow, &workflow.Status.KnowledgeWorkspace, nil
	}

	return nil, nil, fmt.Errorf("no parent object found for onedrive link %q", onedriveLinks.Name)
}

func knowledgeFilesForUploadName(req router.Request, namespace, name string) (v1.KnowledgeFileList, error) {
	var files v1.KnowledgeFileList
	return files, req.List(&files, &client.ListOptions{
		FieldSelector: fields.SelectorFromSet(map[string]string{"spec.uploadName": name}),
		Namespace:     namespace,
	})
}

func compileKnowledgeFilesFromOneDriveConnector(ctx context.Context, c client.Client, oneDriveLinks *v1.OneDriveLinks, files map[string]v1.FileDetails, status *v1.KnowledgeWorkspaceStatus) (map[string]any, map[string]struct{}, error) {
	var (
		errs []error
		// fileMetadata is the metadata for the knowledge tool, translated from the connector output.
		fileMetadata                 = make(map[string]any, len(files))
		outputDir                    = workspace.GetDir(status.KnowledgeWorkspaceID)
		knowledgeFileNamesFromOutput = make(map[string]struct{}, len(files))
	)
	for id, v := range files {
		fileRelPath, err := filepath.Rel(outputDir, v.FilePath)
		if err != nil {
			fileRelPath = v.FilePath
		}

		fileMetadata[fileRelPath] = map[string]any{
			"source": v.URL,
		}

		name := v1.ObjectNameFromAbsolutePath(v.FilePath)
		knowledgeFileNamesFromOutput[name] = struct{}{}
		newKnowledgeFile := &v1.KnowledgeFile{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: oneDriveLinks.Namespace,
			},
			Spec: v1.KnowledgeFileSpec{
				AgentName:    oneDriveLinks.Spec.AgentName,
				WorkflowName: oneDriveLinks.Spec.WorkflowName,
				FileName:     fileRelPath,
				UploadName:   oneDriveLinks.Name,
			},
		}
		if err := c.Create(ctx, newKnowledgeFile); err == nil || apierrors.IsAlreadyExists(err) {
			// If the file was created or already existed, ensure it has the latest details from the metadata.
			if err := retry.OnError(retry.DefaultRetry, func(err error) bool {
				return !apierrors.IsNotFound(err)
			}, func() error {
				if err := c.Get(ctx, router.Key(newKnowledgeFile.Namespace, newKnowledgeFile.Name), uncached.Get(newKnowledgeFile)); err != nil {
					return err
				}

				newKnowledgeFile.Status.FileDetails = v
				newKnowledgeFile.Status.UploadID = id
				return c.Status().Update(ctx, newKnowledgeFile)
			}); err != nil {
				errs = append(errs, fmt.Errorf("failed to update knowledge file %q status: %w", newKnowledgeFile.Name, err))
			}
		} else if err != nil {
			errs = append(errs, err)
		}

		status.HasKnowledge = true
	}

	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("failed to create knowledge files: %w", errors.Join(errs...))
	}

	return fileMetadata, knowledgeFileNamesFromOutput, nil
}

func deleteKnowledgeFilesNotIncluded(ctx context.Context, c client.Client, namespace, name string, filenames map[string]struct{}) error {
	var knowledgeFiles v1.KnowledgeFileList
	if err := c.List(ctx, uncached.List(&knowledgeFiles), &client.ListOptions{
		Namespace:     namespace,
		FieldSelector: fields.SelectorFromSet(map[string]string{"spec.uploadName": name}),
	}); err != nil {
		return fmt.Errorf("failed to list knowledge files: %w", err)
	}

	var errs []error
	for _, knowledgeFile := range knowledgeFiles.Items {
		if _, exists := filenames[knowledgeFile.Name]; !exists {
			if err := c.Delete(ctx, &knowledgeFile); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete knowledge file %q: %w", knowledgeFile.Name, err))
			}
		}
	}

	return errors.Join(errs...)
}

func compileKnowledgeFilesForOneDriveConnector(files v1.KnowledgeFileList) map[string]v1.FileDetails {
	knowledgeFileStatuses := make(map[string]v1.FileDetails, len(files.Items))
	for _, file := range files.Items {
		knowledgeFileStatuses[file.Status.UploadID] = file.Status.FileDetails
	}

	return knowledgeFileStatuses
}
