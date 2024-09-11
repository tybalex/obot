package controller

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/agents"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/runs"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/slugs"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/threads"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workflow"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workflowexecution"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workflowstep"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	"github.com/gptscript-ai/otto/pkg/services"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func routes(router *router.Router, services *services.Services) error {
	workflows := workflow.New(services.WorkspaceClient, "directory")
	workflowExecution := workflowexecution.New(services.WorkspaceClient)
	workflowStep := workflowstep.New(services.Invoker)
	ingester := knowledge.NewIngester(services.Invoker, services.KnowledgeTool)
	agents := agents.New(services.WorkspaceClient, ingester, "directory", "directory", services.AIHelper)
	threads := threads.New(services.WorkspaceClient, ingester)

	root := router

	// Runs
	root.Type(&v1.Run{}).FinalizeFunc(v1.RunFinalizer, runs.DeleteRunState)
	root.Type(&v1.Run{}).HandlerFunc(runs.Cleanup)

	// Threads
	root.Type(&v1.Thread{}).FinalizeFunc(v1.ThreadFinalizer, threads.RemoveWorkspaces)
	root.Type(&v1.Thread{}).HandlerFunc(threads.Cleanup)
	root.Type(&v1.Thread{}).HandlerFunc(threads.IngestKnowledge)

	// Workflows
	root.Type(&v1.Workflow{}).FinalizeFunc(v1.WorkflowFinalizer, workflows.Finalize)
	root.Type(&v1.Workflow{}).HandlerFunc(workflows.CreateWorkspace)

	// WorkflowExecutions
	root.Type(&v1.WorkflowExecution{}).FinalizeFunc(v1.WorkflowExecutionFinalizer, workflowExecution.Finalize)
	root.Type(&v1.WorkflowExecution{}).HandlerFunc(workflowExecution.Cleanup)
	root.Type(&v1.WorkflowExecution{}).HandlerFunc(workflowExecution.Run)

	// WorkflowSteps
	steps := root.Type(&v1.WorkflowStep{})
	steps.HandlerFunc(workflowStep.SetRunning)
	steps.HandlerFunc(workflowStep.Cleanup)

	running := steps.Middleware(workflowstep.Running)
	running.HandlerFunc(workflowStep.RunInvoke)
	running.HandlerFunc(workflowStep.RunIf)
	running.HandlerFunc(workflowStep.RunForEach)
	running.HandlerFunc(workflowStep.RunWhile)

	// Agents
	root.Type(&v1.Agent{}).FinalizeFunc(v1.AgentFinalizer, agents.RemoveWorkspaces)
	root.Type(&v1.Agent{}).HandlerFunc(agents.Suggestion)
	root.Type(&v1.Agent{}).HandlerFunc(agents.CreateWorkspaces)
	root.Type(&v1.Agent{}).HandlerFunc(agents.IngestKnowledge)

	// Slugs
	root.Type(&v1.Slug{}).HandlerFunc(slugs.SlugGC)
	root.Type(&v1.Agent{}).HandlerFunc(slugs.AssociateWithSlug)
	root.Type(&v1.Workflow{}).HandlerFunc(slugs.AssociateWithSlug)

	return nil
}
