package apiclient

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/otto/apiclient/types"
)

type ListToolReferencesOptions struct {
	ToolType types.ToolReferenceType
}

func (c *Client) GetToolReference(ctx context.Context, id string) (result *types.ToolReference, _ error) {
	_, resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/toolreferences/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return toObject(resp, &types.ToolReference{})
}

func (c *Client) ListToolReferences(ctx context.Context, opts ListToolReferencesOptions) (result types.ToolReferenceList, _ error) {
	path := "/toolreferences"
	if opts.ToolType != "" {
		path = fmt.Sprintf("/toolreferences?type=%s", opts.ToolType)
	}
	_, resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	_, err = toObject(resp, &result)
	return result, err
}

func (c *Client) UpdateToolReference(ctx context.Context, id, reference string) (*types.ToolReference, error) {
	_, resp, err := c.putJSON(ctx, fmt.Sprintf("/toolreferences/%s", id), map[string]string{
		"reference": reference,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return toObject(resp, &types.ToolReference{})
}

func (c *Client) DeleteToolReference(ctx context.Context, id string, toolType types.ToolReferenceType) error {
	path := fmt.Sprintf("/toolreferences/%s", id)
	if toolType != "" {
		path = fmt.Sprintf("/toolreferences/%s?type=%s", id, toolType)
	}
	_, resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) CreateToolReference(ctx context.Context, manifest types.ToolReferenceManifest) (*types.ToolReference, error) {
	_, resp, err := c.postJSON(ctx, fmt.Sprintf("/toolreferences"), &manifest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return toObject(resp, &types.ToolReference{})
}
