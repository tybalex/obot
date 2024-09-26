package apiclient

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gptscript-ai/otto/apiclient/types"
)

func (c *Client) DeleteThread(ctx context.Context, id string) error {
	_, resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/threads/"+id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type ThreadEventsOptions struct {
	Follow bool
}

func (c *Client) ThreadEvents(ctx context.Context, threadID string, opts ThreadEventsOptions) (result <-chan types.Progress, err error) {
	path := fmt.Sprintf("/threads/%s/events", threadID)
	if opts.Follow {
		path += "?follow=true"
	}

	_, resp, err := c.doStream(ctx, http.MethodGet, path, nil)
	if err != nil {
		return
	}

	return toStream[types.Progress](resp), nil
}

type ListThreadsOptions struct {
	AgentID string
}

func (c *Client) UpdateThread(ctx context.Context, id string, thread types.ThreadManifest) (*types.Thread, error) {
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
			return result.Items[i].Created.Time.Before(result.Items[j].Created.Time)
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
