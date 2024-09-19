package client

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gptscript-ai/otto/pkg/api/types"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func (c *Client) DeleteThread(ctx context.Context, id string) error {
	_, resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/threads/"+id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type ListThreadsOptions struct {
	AgentID string
}

func (c *Client) UpdateThread(ctx context.Context, id string, thread v1.ThreadManifest) (*types.Thread, error) {
	_, resp, err := c.putJSON(ctx, fmt.Sprintf("/threads/%s", id), thread)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return toObject(resp, &types.Thread{})
}

func (c *Client) ListThreads(ctx context.Context, opts ...ListThreadsOptions) (result types.ThreadList, err error) {
	defer func() {
		sort.Slice(result.Items, func(i, j int) bool {
			return result.Items[i].Created.Before(result.Items[j].Created)
		})
	}()

	var opt ListThreadsOptions
	for _, o := range opts {
		if o.AgentID != "" {
			opt.AgentID = o.AgentID
		}
	}
	url := "/threads"
	if opt.AgentID != "" {
		url = fmt.Sprintf("/agents/%s", opt.AgentID) + url
	}
	_, resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = toObject(resp, &result)
	return
}
