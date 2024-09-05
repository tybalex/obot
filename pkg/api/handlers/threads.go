package handlers

import (
	"context"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

type ThreadHandler struct {
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
