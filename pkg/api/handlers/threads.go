package handlers

import (
	"fmt"
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/events"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	wclient "github.com/thedadams/workspace-provider/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ThreadHandler struct {
	workspaceClient *wclient.Client
	events          *events.Emitter
}

func NewThreadHandler(wc *wclient.Client, events *events.Emitter) *ThreadHandler {
	return &ThreadHandler{
		workspaceClient: wc,
		events:          events,
	}
}

func convertThread(thread v1.Thread) types.Thread {
	return types.Thread{
		Metadata: types.MetadataFrom(&thread),
		ThreadManifest: v1.ThreadManifest{
			Description: thread.Spec.Manifest.Description,
			Tools:       thread.Spec.Manifest.Tools,
		},
		AgentID:      thread.Spec.AgentName,
		WorkflowID:   thread.Spec.WorkflowName,
		LastRunID:    thread.Status.LastRunName,
		LastRunState: thread.Status.LastRunState,
	}
}

func (a *ThreadHandler) Events(req api.Context) error {
	var (
		id     = req.PathValue("id")
		follow = req.URL.Query().Get("follow") == "true"
		thread v1.Thread
	)

	if err := req.Get(&thread, id); err != nil {
		return err
	}

	events, err := a.events.Watch(req.Context(), req.Namespace(), events.WatchOptions{
		Follow:     follow,
		History:    true,
		ThreadName: thread.Name,
	})
	if err != nil {
		return err
	}

	return req.WriteEvents(events)
}

func (a *ThreadHandler) Delete(req api.Context) error {
	var (
		id = req.PathValue("id")
	)
	return req.Delete(&v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: req.Namespace(),
		},
	})
}

func (a *ThreadHandler) Update(req api.Context) error {
	var (
		id        = req.PathValue("id")
		newThread v1.ThreadManifest
		existing  v1.Thread
	)

	if err := req.Get(&existing, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	if err := req.Read(&newThread); err != nil {
		return err
	}

	existing.Spec.Manifest = newThread
	if err := req.Update(&existing); err != nil {
		return err
	}

	return req.Write(convertThread(existing))
}

func (a *ThreadHandler) List(req api.Context) error {
	var (
		agentName  = req.PathValue("agent")
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

	return req.Write(resp)
}

func (a *ThreadHandler) Files(req api.Context) error {
	var (
		id     = req.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	return listFiles(req.Context(), req, a.workspaceClient, thread.Spec.WorkspaceID)
}

func (a *ThreadHandler) UploadFile(req api.Context) error {
	var (
		id     = req.PathValue("id")
		thread v1.Thread
	)
	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	if err := uploadFile(req.Context(), req, a.workspaceClient, thread.Spec.WorkspaceID); err != nil {
		return err
	}

	req.WriteHeader(http.StatusCreated)
	return nil
}

func (a *ThreadHandler) DeleteFile(req api.Context) error {
	var (
		id     = req.PathValue("id")
		thread v1.Thread
	)

	if err := req.Get(&thread, id); err != nil {
		return fmt.Errorf("failed to get thread with id %s: %w", id, err)
	}

	return deleteFile(req.Context(), req, a.workspaceClient, thread.Spec.WorkspaceID)
}

func (a *ThreadHandler) Knowledge(req api.Context) error {
	return listKnowledgeFiles(req, new(v1.Thread))
}

func (a *ThreadHandler) UploadKnowledge(req api.Context) error {
	return uploadKnowledge(req, a.workspaceClient, req.PathValue("id"), new(v1.Thread))
}

func (a *ThreadHandler) DeleteKnowledge(req api.Context) error {
	return deleteKnowledge(req, req.PathValue("file"), req.PathValue("id"), new(v1.Thread))
}

func (a *ThreadHandler) IngestKnowledge(req api.Context) error {
	return ingestKnowledge(req, a.workspaceClient, req.PathValue("id"), new(v1.Thread))
}
