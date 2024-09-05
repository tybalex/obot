package handlers

import (
	"context"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/gz"
	v2 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

type RunHandler struct {
}

func convertRun(run v2.Run) types.Run {
	return types.Run{
		ID:            run.Name,
		Created:       run.CreationTimestamp.Time,
		ThreadID:      run.Spec.ThreadName,
		AgentID:       run.Spec.AgentName,
		PreviousRunID: run.Spec.PreviousRunName,
		Input:         run.Spec.Input,
		State:         run.Status.State,
		Output:        run.Status.Output,
		Error:         run.Status.Error,
	}
}

func (a *RunHandler) Debug(_ context.Context, req api.Request) error {
	var (
		runID = req.Request.PathValue("run")
	)

	var run v2.RunState
	if err := req.Get(&run, runID); err != nil {
		return err
	}

	calls := map[string]any{}
	if err := gz.Decompress(&calls, run.Spec.CallFrame); err != nil {
		return err
	}

	return req.JSON(calls)
}

func (a *RunHandler) List(_ context.Context, req api.Request) error {
	var (
		agentName  = req.Request.PathValue("agent")
		threadName = req.Request.PathValue("thread")
		runList    v2.RunList
	)
	if err := req.List(&runList); err != nil {
		return err
	}

	var resp types.RunList
	for _, run := range runList.Items {
		if agentName != "" && run.Spec.AgentName != agentName {
			continue
		}
		if threadName != "" && run.Spec.ThreadName != threadName {
			continue
		}
		resp.Items = append(resp.Items, convertRun(run))
	}

	return req.JSON(resp)
}
