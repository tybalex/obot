package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/api/types"
)

type ListRunsOptions struct {
	AgentID  string
	ThreadID string
}

func (c *Client) RunEvents(ctx context.Context, runID string) (result <-chan types.Progress, err error) {
	_, resp, err := c.doStream(ctx, http.MethodGet, fmt.Sprintf("/runs/%s/events", runID), nil)
	if err != nil {
		return
	}

	return toStream[types.Progress](resp), nil
}

func (c *Client) DebugRun(ctx context.Context, runID string) (result types.RunDebug, err error) {
	_, resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/%s/debug", runID), nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result.Frames = map[string]gptscript.CallFrame{}
	err = json.NewDecoder(resp.Body).Decode(&result.Frames)
	return
}

func (c *Client) StreamRuns(ctx context.Context, opts ...ListRunsOptions) (result <-chan types.Run, err error) {
	url := c.runURLFromOpts(opts...)
	_, resp, err := c.doStream(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	return toStream[types.Run](resp), nil
}

func (c *Client) GetRun(ctx context.Context, id string) (result *types.Run, err error) {
	_, resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/"+id), nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return toObject(resp, &types.Run{})
}

func (c *Client) ListRuns(ctx context.Context, opts ...ListRunsOptions) (result types.RunList, err error) {
	defer func() {
		sort.Slice(result.Items, func(i, j int) bool {
			return result.Items[i].Created.Before(result.Items[j].Created)
		})
	}()

	url := c.runURLFromOpts(opts...)
	_, resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = toObject(resp, &result)
	return
}
