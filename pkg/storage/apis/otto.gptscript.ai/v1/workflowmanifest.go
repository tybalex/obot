package v1

import (
	"strings"
)

type WorkflowManifest struct {
	AgentManifest `json:",inline"`
	Steps         []Step `json:"steps,omitempty"`
	Output        string `json:"output,omitempty"`
}

type Step struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	If          *If       `json:"if,omitempty"`
	While       *While    `json:"while,omitempty"`
	Template    *Template `json:"template,omitempty"`
	Tools       []string  `json:"tools,omitempty"`
	Agents      []string  `json:"agents,omitempty"`
	Workflows   []string  `json:"workflows,omitempty"`

	Step        string   `json:"step,omitempty"`
	Cache       *bool    `json:"cache,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
}

type Template struct {
	Name string            `json:"name,omitempty"`
	Args map[string]string `json:"args,omitempty"`
}

func oneline(s string) string {
	l := strings.Split(s, "\n")[0]
	if len(l) > 80 {
		return l[:80] + "..."
	}
	return l
}

func (s Step) Display() string {
	preamble := strings.Builder{}
	preamble.WriteString("> Step(")
	preamble.WriteString(s.ID)
	preamble.WriteString("): ")
	if s.Name != "" {
		preamble.WriteString(s.Name)
	}
	if s.While != nil {
		preamble.WriteString(" while ")
		preamble.WriteString(oneline(s.While.Condition))
	}
	if s.If != nil {
		preamble.WriteString(" if ")
		preamble.WriteString(oneline(s.If.Condition))
	}
	if s.Step != "" {
		preamble.WriteString(" ")
		preamble.WriteString(oneline(s.Step))
	}
	return preamble.String()
}

type SubFlow struct {
	Input    string `json:"input,omitempty"`
	Workflow string `json:"workflow,omitempty"`
}

type If struct {
	Condition string `json:"condition,omitempty"`
	Steps     []Step `json:"steps,omitempty"`
	Else      []Step `json:"else,omitempty"`
}

type While struct {
	Condition string `json:"condition,omitempty"`
	MaxLoops  int    `json:"maxLoops,omitempty"`
	Steps     []Step `json:"steps,omitempty"`
}
