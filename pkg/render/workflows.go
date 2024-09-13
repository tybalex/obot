package render

import (
	"github.com/gptscript-ai/go-gptscript"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func Step(step *v1.WorkflowStep) any {
	if step.Spec.Step.AgentStep != nil {
		return []gptscript.ToolDef{{
			Chat:         true,
			Tools:        step.Spec.Step.Tools,
			Instructions: step.Spec.Step.AgentStep.Prompt.Instructions(),
			Type:         "agent",
			MetaData:     step.Spec.Step.AgentStep.Prompt.Metadata(step.Spec.Step.CodeDependencies),
			Cache:        step.Spec.Step.Cache,
			Temperature:  step.Spec.Step.Temperature,
		}}
	} else if step.Spec.Step.ToolStep != nil && step.Spec.Step.ToolStep.Tool != "" {
		if step.Spec.Step.ToolStep.Tool.IsInline() {
			return []gptscript.ToolDef{{
				Chat:         true,
				Tools:        step.Spec.Step.Tools,
				Instructions: step.Spec.Step.ToolStep.Tool.Instructions(),
				MetaData:     step.Spec.Step.ToolStep.Tool.Metadata(step.Spec.Step.CodeDependencies),
				Temperature:  step.Spec.Step.Temperature,
				Cache:        step.Spec.Step.Cache,
			}}
		} else {
			return step.Spec.Step.ToolStep.Tool
		}
	} else {
		return nil
	}
}
