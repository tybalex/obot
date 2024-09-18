package uploads

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/invoke"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/system"
	"github.com/gptscript-ai/otto/pkg/workspace"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
			Finalizers:   []string{v1.ThreadFinalizer},
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

	_, status, err := parentObjAndStatus(req, oneDriveLinks)
	if err != nil {
		return err
	}

	writer, err := u.workspaceClient.WriteFile(req.Ctx, thread.Spec.WorkspaceID, ".metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	b, err := json.Marshal(map[string]any{
		"input": map[string]any{
			"sharedLinks": oneDriveLinks.Spec.SharedLinks,
			"outputDir":   filepath.Join(workspace.GetDir(status.KnowledgeWorkspaceID), oneDriveLinks.Name),
		},
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

	file, err := u.workspaceClient.OpenFile(req.Ctx, thread.Spec.WorkspaceID, ".metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	var output map[string]v1.OnedriveLinksStatus
	if err = json.NewDecoder(file).Decode(&output); err != nil {
		return fmt.Errorf("failed to decode metadata file: %w", err)
	}

	if err = file.Close(); err != nil {
		return fmt.Errorf("failed to close metadata file: %w", err)
	}

	// Put the metadata file in the agent knowledge workspace
	fileMetadata := make(map[string]any, len(output["output"].Files))
	outputDir := filepath.Join(workspace.GetDir(status.KnowledgeWorkspaceID), oneDriveLinks.Name)
	for _, v := range output["output"].Files {
		fileRelPath, err := filepath.Rel(outputDir, v.FilePath)
		if err != nil {
			fileRelPath = v.FilePath
		}

		fileMetadata[fileRelPath] = map[string]any{
			"source": v.URL,
		}
	}

	writer, err := u.workspaceClient.WriteFile(req.Ctx, status.KnowledgeWorkspaceID, filepath.Join(oneDriveLinks.Name, ".knowledge.json"))
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	if err = json.NewEncoder(writer).Encode(map[string]any{"metadata": fileMetadata}); err != nil {
		return fmt.Errorf("failed to encode metadata file: %w", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close metadata file: %w", err)
	}

	// Reset the agent knowledge generation so that ingestion kicks off again.
	status.KnowledgeGeneration++
	status.HasKnowledge = true
	if err = req.Client.Status().Update(req.Ctx, parentObj); err != nil {
		return fmt.Errorf("failed to update agent observed knowledge generation: %w", err)
	}

	outputStatus := output["output"]
	oneDriveLinks.Status.Error = outputStatus.Error
	oneDriveLinks.Status.Status = outputStatus.Status
	oneDriveLinks.Status.Files = outputStatus.Files
	oneDriveLinks.Status.Folders = outputStatus.Folders

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

func parentObjAndStatus(req router.Request, onedriveLinks *v1.OneDriveLinks) (client.Object, *v1.KnowledgeWorkspaceStatus, error) {
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
