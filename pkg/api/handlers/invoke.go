package handlers

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

type InvokeHandler struct {
	Invoker *invoke.Invoker
}

func (i *InvokeHandler) Invoke(ctx context.Context, req api.Request) error {
	var (
		agentID  = req.PathValue("agent")
		agent    v1.Agent
		threadID = req.PathValue("thread")
	)

	if err := req.Get(&agent, agentID); err != nil {
		return err
	}

	input, err := req.Body()
	if err != nil {
		return err
	}

	resp, err := i.Invoker.Invoke(ctx, &agent, string(input), invoke.Options{
		ThreadName: threadID,
	})
	if err != nil {
		return err
	}

	req.ResponseWriter.Header().Set("X-Otto-Thread-Id", resp.Thread.Name)
	req.ResponseWriter.Header().Set("X-Otto-Run-Id", resp.Run.Name)

	// Check if SSE is requested
	sendEvents := slices.Contains(req.Request.Header.Values("Accept"), "text/event-stream")
	if sendEvents {
		req.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
	}

	var lastFlush time.Time
	for event := range resp.Events {
		if sendEvents {
			if _, err := req.Write([]byte("data: ")); err != nil {
				return err
			}
			if err := json.NewEncoder(req.ResponseWriter).Encode(event); err != nil {
				return err
			}
			if _, err := req.Write([]byte("\n\n")); err != nil {
				return err
			}
			req.Flush()
		} else {
			if _, err := req.Write([]byte(event.Content)); err != nil {
				return err
			}
			if lastFlush.IsZero() || time.Since(lastFlush) > 500*time.Millisecond {
				req.Flush()
				lastFlush = time.Now()
			}
		}
	}

	return nil
}
