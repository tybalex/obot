package v1

import (
	"maps"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gptscript-ai/go-gptscript"
)

type Manifest struct {
	ID          string
	Name        string
	Slug        string
	Description string
	Prompt      string
	Tools       []string
	Agents      []string
	Params      map[string]string
}

func (m Manifest) GetParams() *openapi3.Schema {
	var args []string
	for _, k := range slices.Sorted(maps.Keys(m.Params)) {
		args = append(args, k)
		args = append(args, m.Params[k])
	}

	return gptscript.ObjectSchema(args...)
}
