package expression

import v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"

type Workflow struct {
	execution *v1.WorkflowExecution
	steps     *Steps
}

func (w Workflow) Keys() ([]string, error) {
	return []string{
		"input",
		"steps",
	}, nil
}

func (w Workflow) Get(s string) (any, bool, error) {
	switch s {
	case "input":
		return &StringWrapper{value: w.execution.Spec.Input}, true, nil
	case "steps":
		return w.steps, true, nil
	}
	return nil, false, nil
}
