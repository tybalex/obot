package apiclient

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/otto8-ai/otto8/apiclient/types"
)

func (c *Client) UpdateWorkflow(ctx context.Context, id string, manifest types.WorkflowManifest) (*types.Workflow, error) {
	_, resp, err := c.putJSON(ctx, fmt.Sprintf("/workflows/%s", id), manifest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return toObject(resp, &types.Workflow{})
}

func (c *Client) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	_, resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/workflows/"+id), nil)
	if err != nil {
		return nil, err
	}

	return toObject(resp, &types.Workflow{})
}

func (c *Client) CreateWorkflow(ctx context.Context, workflow types.WorkflowManifest) (*types.Workflow, error) {
	_, resp, err := c.postJSON(ctx, fmt.Sprintf("/workflows"), workflow)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return toObject(resp, &types.Workflow{})
}

type ListWorkflowsOptions struct {
	RefName string
}

func (c *Client) ListWorkflows(ctx context.Context, opts ListWorkflowsOptions) (result types.WorkflowList, err error) {
	defer func() {
		sort.Slice(result.Items, func(i, j int) bool {
			return result.Items[i].Metadata.Created.Time.After(result.Items[j].Metadata.Created.Time)
		})
	}()

	_, resp, err := c.doRequest(ctx, http.MethodGet, "/workflows", nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = toObject(resp, &result)
	if err != nil {
		return result, err
	}

	if opts.RefName != "" {
		var filtered types.WorkflowList
		for _, workflow := range result.Items {
			if workflow.RefName == opts.RefName && workflow.RefNameAssigned {
				filtered.Items = append(filtered.Items, workflow)
			}
		}
		result = filtered
	}

	return
}

func (c *Client) DeleteWorkflow(ctx context.Context, id string) error {
	_, resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/workflows/"+id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
