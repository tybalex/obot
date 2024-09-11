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
			Instructions: step.Spec.Step.AgentStep.Prompt.Instructions(),
			Type:         "agent",
			MetaData:     step.Spec.Step.AgentStep.Prompt.Metadata(step.Spec.Step.AgentStep.CodeDependencies),
		}}
	} else if step.Spec.Step.ToolStep != nil {
		return []gptscript.ToolDef{{
			Chat:         true,
			Tools:        step.Spec.Step.AgentStep.Tools,
			Instructions: step.Spec.Step.ToolStep.Tool.Instructions(),
			MetaData:     step.Spec.Step.ToolStep.Tool.Metadata(step.Spec.Step.ToolStep.CodeDependencies),
			Type:         "agent",
		}}
	} else {
		return nil
	}
}
