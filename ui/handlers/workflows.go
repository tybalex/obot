package handlers

import (
	"net/http"

	"github.com/otto8-ai/otto8/apiclient"
	"github.com/otto8-ai/otto8/apiclient/types"
	"github.com/otto8-ai/otto8/ui/components"
	"github.com/otto8-ai/otto8/ui/pages"
	"github.com/otto8-ai/otto8/ui/webcontext"
)

func RunWorkflow(rw http.ResponseWriter, req *http.Request) error {
	var (
		id  = req.PathValue("id")
		c   = webcontext.Client(req.Context())
		ctx = req.Context()
	)

	wf, err := c.GetWorkflow(ctx, id)
	if err != nil {
		return err
	}

	resp, err := c.Invoke(ctx, wf.ID, "", apiclient.InvokeOptions{
		Async: true,
	})
	if err != nil {
		return err
	}

	return Render(rw, req, components.NewThread(resp.ThreadID, resp.ThreadID))
}

func EditWorkflow(rw http.ResponseWriter, req *http.Request) error {
	var (
		id  = req.PathValue("id")
		c   = webcontext.Client(req.Context())
		ctx = req.Context()
	)

	wf, err := c.GetWorkflow(ctx, id)
	if err != nil {
		return err
	}

	steps, err := c.ListToolReferences(ctx, apiclient.ListToolReferencesOptions{
		ToolType: types.ToolReferenceTypeStepTemplate,
	})
	if err != nil {
		return err
	}

	return Render(rw, req, pages.NewEditWorkflow(*wf, steps))
}
