package v1

import (
	"maps"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gptscript-ai/go-gptscript"
)

type AgentManifest struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Slug        string            `json:"slug,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`
	Tools       []string          `json:"tools,omitempty"`
	Agents      []string          `json:"agents,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (m AgentManifest) GetParams() *openapi3.Schema {
	var args []string
	for _, k := range slices.Sorted(maps.Keys(m.Params)) {
		args = append(args, k)
		args = append(args, m.Params[k])
	}

	return gptscript.ObjectSchema(args...)
}
