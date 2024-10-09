package handlers

import (
	"net/http"
	"strings"

	"github.com/otto8-ai/otto8/apiclient"
	"github.com/otto8-ai/otto8/apiclient/types"
	"github.com/otto8-ai/otto8/ui/components"
	"github.com/otto8-ai/otto8/ui/pages"
	"github.com/otto8-ai/otto8/ui/webcontext"
)

func AddStep(rw http.ResponseWriter, req *http.Request) error {
	var (
		workflowID = req.PathValue("workflow_id")
		parentID   = req.PathValue("parent_id")
		toolRefID  = req.PathValue("id")
		c          = webcontext.Client(req.Context())
		ctx        = req.Context()
	)

	if parentID == "_" {
		parentID = ""
	}

	wf, err := c.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}

	tools, err := c.ListToolReferences(ctx, apiclient.ListToolReferencesOptions{
		ToolType: types.ToolReferenceTypeStepTemplate,
	})
	if err != nil {
		return err
	}

	var newStep types.Step
	switch toolRefID {
	case "sys.if":
		newStep = types.Step{
			If: &types.If{},
		}
	case "sys.while":
		newStep = types.Step{
			While: &types.While{},
		}
	case "sys.prompt":
		// nothing to change
	default:
		toolRef, err := c.GetToolReference(ctx, toolRefID)
		if err != nil {
			return err
		}

		newStep = types.Step{
			Template: &types.Template{
				Name: toolRef.ID,
				Args: map[string]string{},
			},
		}

		for key := range toolRef.Params {
			newStep.Template.Args[key] = ""
		}
	}

	manifest := types.AppendStep(&wf.WorkflowManifest, parentID, newStep)
	wf, err = c.UpdateWorkflow(ctx, wf.ID, *manifest)
	if err != nil {
		return err
	}

	return Render(rw, req, components.NewEditWorkflow("edit-workflow", *wf, tools))
}

func CreateWorkflow(rw http.ResponseWriter, req *http.Request) error {
	var (
		c   = webcontext.Client(req.Context())
		ctx = req.Context()
	)

	wf, err := c.CreateWorkflow(ctx, types.WorkflowManifest{})
	if err != nil {
		return err
	}

	rw.Header().Set("HX-Redirect", "/ui/workflows/"+wf.ID+"/edit")
	return nil
}

func Workflows(rw http.ResponseWriter, req *http.Request) error {
	var (
		c   = webcontext.Client(req.Context())
		ctx = req.Context()
	)

	workflows, err := c.ListWorkflows(ctx, apiclient.ListWorkflowsOptions{})
	if err != nil {
		return err
	}

	return Render(rw, req, pages.Workflows(workflows.Items))
}

func WorkflowThread(rw http.ResponseWriter, req *http.Request) error {
	var (
		threadID = req.PathValue("thread_id")
	)

	return Render(rw, req, components.NewThread(threadID, threadID))
}

func NewStep(rw http.ResponseWriter, req *http.Request) error {
	var (
		workflowID = req.PathValue("workflow_id")
		parentID   = req.URL.Query().Get("parentID")
		c          = webcontext.Client(req.Context())
		ctx        = req.Context()
	)

	if parentID == "" {
		parentID = "_"
	}

	wf, err := c.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}

	steps, err := c.ListToolReferences(ctx, apiclient.ListToolReferencesOptions{
		ToolType: types.ToolReferenceTypeStepTemplate,
	})
	if err != nil {
		return err
	}

	return Render(rw, req, components.NewStepModal(components.NewStepModalData{
		Workflow:      *wf,
		StepTemplates: steps,
		ParentID:      parentID,
	}))
}

func UpdateStep(rw http.ResponseWriter, req *http.Request) error {
	var (
		workflowID = req.PathValue("workflow_id")
		stepID     = req.PathValue("id")
		c          = webcontext.Client(req.Context())
		ctx        = req.Context()
	)

	wf, err := c.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}

	step, parentID := types.FindStep(&wf.WorkflowManifest, stepID)
	if step == nil {
		return types.NewErrNotFound("failed to find step %s in workflow %s", stepID, workflowID)
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	var (
		args    = map[string]string{}
		hasArgs bool
	)
	for key, value := range req.Form {
		if len(value) == 0 {
			continue
		}
		if p, ok := strings.CutPrefix(key, "template."); ok {
			hasArgs = true
			args[p] = value[0]
		}

		switch key {
		case "sys.prompt":
			step.SetPrompt(value[0])
		case "sys.condition":
			step.SetCondition(value[0])
		}
	}

	if hasArgs {
		step.SetArgs(args)
	}

	types.SetStep(&wf.WorkflowManifest, *step)
	if _, err := c.UpdateWorkflow(ctx, workflowID, wf.WorkflowManifest); err != nil {
		return err
	}

	toolRefList, err := c.ListToolReferences(ctx, apiclient.ListToolReferencesOptions{})
	if err != nil {
		return err
	}

	toolRefs := map[string]*types.ToolReference{}
	for _, toolRef := range toolRefList.Items {
		toolRefs[toolRef.ID] = &toolRef
	}

	return Render(rw, req, components.NewStep(*wf, parentID, *step, toolRefs))
}

func DeleteWorkflow(rw http.ResponseWriter, req *http.Request) error {
	var (
		workflowID = req.PathValue("workflow_id")
		c          = webcontext.Client(req.Context())
		ctx        = req.Context()
	)

	if err := c.DeleteWorkflow(ctx, workflowID); err != nil {
		return err
	}

	return nil
}

func DeleteStep(rw http.ResponseWriter, req *http.Request) error {
	var (
		workflowID = req.PathValue("workflow_id")
		stepID     = req.PathValue("id")
		c          = webcontext.Client(req.Context())
		ctx        = req.Context()
	)

	wf, err := c.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}

	step, _ := types.FindStep(&wf.WorkflowManifest, stepID)
	if step == nil {
		return types.NewErrNotFound("failed to find step %s in workflow %s", stepID, workflowID)
	}

	manifest := types.DeleteStep(&wf.WorkflowManifest, stepID)
	if _, err := c.UpdateWorkflow(ctx, workflowID, *manifest); err != nil {
		return err
	}

	return nil
}
