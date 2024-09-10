package render

import (
	"github.com/gptscript-ai/go-gptscript"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func Step(step *v1.WorkflowStep) []gptscript.ToolDef {
	if step.Spec.Step.AgentStep != nil {
		return []gptscript.ToolDef{{
			Chat:         true,
			Tools:        step.Spec.Step.AgentStep.Tools,
			Instructions: step.Spec.Step.AgentStep.Prompt,
			Type:         "agent",
		}}
	} else if step.Spec.Step.ToolStep != nil {
		return []gptscript.ToolDef{{
			Chat:         true,
			Tools:        step.Spec.Step.AgentStep.Tools,
			Instructions: step.Spec.Step.ToolStep.Tool,
			MetaData:     step.Spec.Step.ToolStep.Metadata,
			Type:         "agent",
		}}
	} else {
		return nil
	}
}
