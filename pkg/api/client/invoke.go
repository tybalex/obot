package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

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

	return &types.InvokeResponse{
		Events:   toStream[types.Progress](resp),
		RunID:    resp.Header.Get("X-Otto-Run-Id"),
		ThreadID: resp.Header.Get("X-Otto-Thread-Id"),
	}, nil
}
