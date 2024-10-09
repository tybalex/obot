package toolreference

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/acorn-io/baaah/pkg/apply"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/apiclient/types"
	"github.com/gptscript-ai/otto/logger"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var log = logger.Package()

type indexEntry struct {
	Reference string `json:"reference,omitempty"`
	All       bool   `json:"all,omitempty"`
}

type index struct {
	Tools         map[string]indexEntry `json:"tools,omitempty"`
	StepTemplates map[string]indexEntry `json:"stepTemplates,omitempty"`
}

type Handler struct {
	gptClient   *gptscript.GPTScript
	registryURL string
}

func New(gptClient *gptscript.GPTScript, registryURL string) *Handler {
	return &Handler{
		gptClient:   gptClient,
		registryURL: registryURL,
	}
}

func isValidTool(tool gptscript.Tool) bool {
	if tool.MetaData["index"] == "false" {
		return false
	}
	return tool.Name != "" && (tool.Type == "" || tool.Type == "tool")
}

func (h *Handler) toolsToToolReferences(ctx context.Context, toolType types.ToolReferenceType, entries map[string]indexEntry) (result []client.Object) {
	for name, entry := range entries {
		if entry.All {
			prg, err := h.gptClient.LoadFile(ctx, "* from "+entry.Reference)
			if err != nil {
				log.Errorf("Failed to load tool %s: %v", entry.Reference, err)
				continue
			}

			tool := prg.ToolSet[prg.EntryToolID]
			if isValidTool(tool) {
				result = append(result, &v1.ToolReference{
					ObjectMeta: metav1.ObjectMeta{
						Name:      normalize(name, tool.Name),
						Namespace: "default",
					},
					Spec: v1.ToolReferenceSpec{
						Type:      toolType,
						Reference: entry.Reference,
					},
				})
			}
			for _, peerToolID := range tool.LocalTools {
				peerTool := prg.ToolSet[peerToolID]
				if isValidTool(peerTool) {
					result = append(result, &v1.ToolReference{
						ObjectMeta: metav1.ObjectMeta{
							Name:      normalize(name, peerTool.Name),
							Namespace: "default",
						},
						Spec: v1.ToolReferenceSpec{
							Type:      toolType,
							Reference: fmt.Sprintf("%s from %s", peerTool.Name, entry.Reference),
						},
					})
				}
			}
		} else {
			result = append(result, &v1.ToolReference{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: "default",
				},
				Spec: v1.ToolReferenceSpec{
					Type:      toolType,
					Reference: entry.Reference,
				},
			})
		}
	}

	return
}

func (h *Handler) readRegistry() (index, error) {
	var input io.ReadCloser

	if strings.HasPrefix(h.registryURL, "file://") {
		file, err := os.Open(strings.TrimPrefix(h.registryURL, "file://"))
		if err != nil {
			return index{}, err
		}
		defer file.Close()
		input = file
	} else {
		resp, err := http.Get(h.registryURL)
		if err != nil {
			return index{}, err
		}
		defer resp.Body.Close()
		input = resp.Body
	}

	data, err := io.ReadAll(input)
	if err != nil {
		log.Errorf("Failed to read index: %v", err)
		return index{}, err
	}

	var index index

	if err := yaml.Unmarshal(data, &index); err != nil {
		log.Errorf("Failed to decode index: %v", err)
		return index, err
	}

	return index, nil
}

func (h *Handler) readFromRegistry(ctx context.Context, c client.Client) error {
	index, err := h.readRegistry()
	if err != nil {
		return err
	}

	var toAdd []client.Object

	toAdd = append(toAdd, h.toolsToToolReferences(ctx, types.ToolReferenceTypeTool, index.Tools)...)
	toAdd = append(toAdd, h.toolsToToolReferences(ctx, types.ToolReferenceTypeStepTemplate, index.StepTemplates)...)

	if len(toAdd) == 0 {
		// Don't accidentally delete all the tool references
		return nil
	}

	return apply.New(c).WithOwnerSubContext("toolreferences").Apply(ctx, nil, toAdd...)
}

func normalize(names ...string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.Join(names, "-"), " ", "-"), "_", "-"))
}

func (h *Handler) PollRegistry(ctx context.Context, c client.Client) {
	if h.registryURL == "" {
		return
	}

	for {
		if err := c.List(ctx, &v1.ToolReferenceList{}, client.InNamespace("default")); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	t := time.NewTicker(1 * time.Hour)
	defer t.Stop()
	for {
		if err := h.readFromRegistry(ctx, c); err != nil {
			log.Errorf("Failed to read from registry: %v", err)
		}

		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}
	}

}

func (h *Handler) Populate(req router.Request, resp router.Response) error {
	toolRef := req.Object.(*v1.ToolReference)
	if toolRef.Generation == toolRef.Status.ObservedGeneration {
		return nil
	}

	// Reset status
	toolRef.Status.ObservedGeneration = toolRef.Generation
	toolRef.Status.Reference = toolRef.Spec.Reference
	toolRef.Status.Tool = nil
	toolRef.Status.Error = ""

	prg, err := h.gptClient.LoadFile(req.Ctx, toolRef.Spec.Reference)
	if err != nil {
		toolRef.Status.Error = err.Error()
		return nil
	}

	tool := prg.ToolSet[prg.EntryToolID]
	toolRef.Status.Tool = &v1.ToolShortDescription{
		Name:        tool.Name,
		Description: tool.Description,
		Metadata:    tool.MetaData,
		Params:      map[string]string{},
	}
	if tool.Arguments != nil {
		for name, param := range tool.Arguments.Properties {
			if param.Value != nil {
				toolRef.Status.Tool.Params[name] = param.Value.Description
			}
		}
	}

	return nil
}
