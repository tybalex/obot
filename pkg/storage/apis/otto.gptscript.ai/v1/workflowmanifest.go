package v1

type WorkflowManifest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Slug        string            `json:"slug,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	Steps       []Step            `json:"steps,omitempty"`
}

type Step struct {
	Name    string    `json:"name,omitempty"`
	Input   StepInput `json:"input,omitempty"`
	Tool    string    `json:"tool,omitempty"`
	If      *If       `json:"if,omitempty"`
	While   *While    `json:"while,omitempty"`
	ForEach *ForEach  `json:"forEach,omitempty"`

	*AgentStep
	*ToolStep
	Tools            []string `json:"tools,omitempty"`
	Temperature      *float32 `json:"temperature,omitempty"`
	CodeDependencies string   `json:"codeDependencies,omitempty"`
}

type AgentStep struct {
	Prompt Body  `json:"prompt,omitempty"`
	Cache  *bool `json:"cache,omitempty"`
}

type ToolStep struct {
	Tool Body `json:"tool,omitempty"`
}

type StepInput struct {
	Content string            `json:"content,omitempty"`
	Args    map[string]string `json:"args,omitempty"`
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

type ForEach struct {
	Items string `json:"items,omitempty"`
	Var   string `json:"var,omitempty"`
	Steps []Step `json:"steps,omitempty"`
}
