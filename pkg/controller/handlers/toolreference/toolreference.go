package toolreference

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

type Handler struct {
	gptClient *gptscript.GPTScript
}

func New(gptClient *gptscript.GPTScript) *Handler {
	return &Handler{
		gptClient: gptClient,
	}
}

func (h *Handler) Populate(req router.Request, resp router.Response) error {
	toolRef := req.Object.(*v1.ToolReference)
	if toolRef.Generation == toolRef.Status.ObservedGeneration {
		return nil
	}

	// Reset status
	toolRef.Status.ObservedGeneration = toolRef.Generation
	toolRef.Status.Tool = nil
	toolRef.Status.Error = ""

	nodes, err := h.gptClient.Parse(req.Ctx, toolRef.Spec.Reference)
	if err != nil {
		toolRef.Status.Error = err.Error()
		return nil
	}

	for _, node := range nodes {
		if node.ToolNode != nil {
			toolRef.Status.Tool = &v1.ToolShortDescription{
				Name:        node.ToolNode.Tool.Name,
				Description: node.ToolNode.Tool.Description,
				Params:      map[string]string{},
			}
			if node.ToolNode.Tool.Arguments != nil {
				for name, param := range node.ToolNode.Tool.Arguments.Properties {
					if param.Value != nil {
						toolRef.Status.Tool.Params[name] = param.Value.Description
					}
				}
			}
			break
		}
	}

	return nil
}
