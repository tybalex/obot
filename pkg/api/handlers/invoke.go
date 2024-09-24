package handlers

import (
	"net/http"

	"github.com/gptscript-ai/otto/pkg/api"
	"github.com/gptscript-ai/otto/pkg/invoke"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/system"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type InvokeHandler struct {
	invoker *invoke.Invoker
}

func NewInvokeHandler(invoker *invoke.Invoker) *InvokeHandler {
	return &InvokeHandler{
		invoker: invoker,
	}
}

func (i *InvokeHandler) Invoke(req api.Context) error {
	var (
		id       = req.PathValue("id")
		agentID  string
		wfID     string
		agent    v1.Agent
		wf       v1.Workflow
		ref      v1.Reference
		threadID = req.PathValue("thread")
		async    = req.URL.Query().Get("async") == "true"
	)

	if threadID == "" {
		threadID = req.Request.Header.Get("X-Otto-Thread-Id")
	}

	if system.IsAgentID(id) {
		agentID = id
	} else if system.IsWorkflowID(id) {
		wfID = id
	} else {
		if err := req.Get(&ref, id); apierrors.IsNotFound(err) {
		} else if err != nil {
			return err
		} else if ref.Spec.AgentName != "" {
			agentID = ref.Spec.AgentName
		} else if ref.Spec.WorkflowName != "" {
			wfID = ref.Spec.WorkflowName
		}
	}

	if agentID != "" {
		if err := req.Get(&agent, agentID); err != nil {
			return err
		}
	} else if wfID != "" {
		if err := req.Get(&wf, wfID); err != nil {
			return err
		}
	} else {
		return apierrors.NewBadRequest("invalid id, most be agent or workflow id")
	}

	input, err := req.Body()
	if err != nil {
		return err
	}

	var resp *invoke.Response

	if agentID != "" {
		resp, err = i.invoker.Agent(req.Context(), req.Storage, &agent, string(input), invoke.Options{
			ThreadName: threadID,
			Background: async,
		})
		if err != nil {
			return err
		}
	} else {
		resp, err = i.invoker.Workflow(req.Context(), req.Storage, &wf, string(input), invoke.WorkflowOptions{
			ThreadName: threadID,
			Background: async,
		})
		if err != nil {
			return err
		}
	}

	req.ResponseWriter.Header().Set("X-Otto-Thread-Id", resp.Thread.Name)

	if async {
		req.WriteHeader(http.StatusCreated)
		return req.Write(map[string]string{
			"threadID": resp.Thread.Name,
		})
	}

	return req.WriteEvents(resp.Events)
}
