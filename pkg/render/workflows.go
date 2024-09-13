package render

import (
	"github.com/gptscript-ai/go-gptscript"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

type StepOptions struct {
	KnowledgeTool string
}

func Step(step *v1.WorkflowStep, opts StepOptions) (_ any, extraEnv []string) {
	if step.Spec.Step.AgentStep != nil {
		tool := gptscript.ToolDef{
			Chat:         true,
			Tools:        step.Spec.Step.Tools,
			Instructions: step.Spec.Step.AgentStep.Prompt.Instructions(),
			Type:         "agent",
			MetaData:     step.Spec.Step.AgentStep.Prompt.Metadata(step.Spec.Step.CodeDependencies),
			Cache:        step.Spec.Step.Cache,
			Temperature:  step.Spec.Step.Temperature,
		}
		if step.Spec.WorkflowKnowledgeWorkspaceID != "" {
			extraEnv = append(extraEnv, "GPTSCRIPT_SCRIPT_ID="+step.Spec.WorkflowKnowledgeWorkspaceID)
			if step.Spec.WorkflowExecutionKnowledgeWorkspaceID != "" {
				extraEnv = append(extraEnv, "OTTO_THREAD_ID="+step.Spec.WorkflowExecutionKnowledgeWorkspaceID)
			}
			if opts.KnowledgeTool != "" {
				tool.Tools = append(tool.Tools, opts.KnowledgeTool)
			}
		}
		return []gptscript.ToolDef{tool}, extraEnv
	} else if step.Spec.Step.ToolStep != nil && step.Spec.Step.ToolStep.Tool != "" {
		if step.Spec.Step.ToolStep.Tool.IsInline() {
			return []gptscript.ToolDef{{
				Chat:         true,
				Tools:        step.Spec.Step.Tools,
				Instructions: step.Spec.Step.ToolStep.Tool.Instructions(),
				MetaData:     step.Spec.Step.ToolStep.Tool.Metadata(step.Spec.Step.CodeDependencies),
				Temperature:  step.Spec.Step.Temperature,
				Cache:        step.Spec.Step.Cache,
			}}, nil
		} else {
			return step.Spec.Step.ToolStep.Tool, nil
		}
	} else {
		return nil, nil
	}
}
