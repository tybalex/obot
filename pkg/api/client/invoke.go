package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api/types"
)

type InvokeOptions struct {
	ThreadID string
	Async    bool
}

func (c *Client) Invoke(ctx context.Context, agentID string, input string, opt ...InvokeOptions) (*types.InvokeResponse, error) {
	var (
		opts InvokeOptions
	)
	for _, o := range opt {
		if o.ThreadID != "" {
			opts.ThreadID = o.ThreadID
		}
		if o.Async {
			opts.Async = o.Async
		}
	}

	url := fmt.Sprintf("/invoke/%s?async=%v", agentID, opts.Async)
	if opts.ThreadID != "" {
		url = fmt.Sprintf("/invoke/%s/threads/%s?async=%v", agentID, opts.ThreadID, opts.Async)
	}

	_, resp, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(input)), "Accept", "text/event-stream")
	if err != nil {
		return nil, err
	}

	if opts.Async {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return &types.InvokeResponse{
			RunID:    resp.Header.Get("X-Otto-Run-Id"),
			ThreadID: resp.Header.Get("X-Otto-Thread-Id"),
		}, nil
	}

	return &types.InvokeResponse{
		Events:   toStream[types.Progress](resp),
		RunID:    resp.Header.Get("X-Otto-Run-Id"),
		ThreadID: resp.Header.Get("X-Otto-Thread-Id"),
	}, nil
}
