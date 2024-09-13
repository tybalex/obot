package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gptscript-ai/otto/pkg/api/types"
)

type InvokeOptions struct {
	ThreadID string
}

func (c *Client) Invoke(ctx context.Context, agentID string, input string, opt ...InvokeOptions) (*types.InvokeResponse, error) {
	var (
		opts InvokeOptions
	)
	for _, o := range opt {
		if o.ThreadID != "" {
			opts.ThreadID = o.ThreadID
		}
	}

	url := fmt.Sprintf("/invoke/%s?events=true", agentID)
	if opts.ThreadID != "" {
		url = fmt.Sprintf("/invoke/%s/threads/%s?events=true", agentID, opts.ThreadID)
	}

	_, resp, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(input)), "Accept", "text/event-stream")
	if err != nil {
		return nil, err
	}

	events := make(chan types.Progress)
	go func() {
		defer resp.Body.Close()
		defer close(events)
		lines := bufio.NewScanner(resp.Body)
		for lines.Scan() {
			var event types.Progress
			data := strings.TrimPrefix(lines.Text(), "data: ")
			if len(data) == 0 {
				continue
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				events <- types.Progress{
					Error: err.Error(),
				}
			} else {
				events <- event
			}
		}

		if err := lines.Err(); err != nil {
			events <- types.Progress{
				Error: err.Error(),
			}
		}
	}()

	return &types.InvokeResponse{
		Events:   events,
		RunID:    resp.Header.Get("X-Otto-Run-Id"),
		ThreadID: resp.Header.Get("X-Otto-Thread-Id"),
	}, nil
}
