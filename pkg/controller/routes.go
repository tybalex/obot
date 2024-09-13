package controller

import (
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/agents"
	knowledgehandler "github.com/gptscript-ai/otto/pkg/controller/handlers/knowledge"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/runs"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/slugs"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/threads"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workflowexecution"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workflowstep"
	"github.com/gptscript-ai/otto/pkg/controller/handlers/workspace"
	"github.com/gptscript-ai/otto/pkg/knowledge"
	"github.com/gptscript-ai/otto/pkg/services"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
)

func routes(router *router.Router, services *services.Services) error {
	workflowExecution := workflowexecution.New(services.WorkspaceClient)
	workflowStep := workflowstep.New(services.Invoker)
	ingester := knowledge.NewIngester(services.Invoker, services.KnowledgeTool)
	agents := agents.New(services.WorkspaceClient, ingester, "directory", services.AIHelper)
	threads := threads.New(services.WorkspaceClient, ingester)
	workspace := workspace.New(services.WorkspaceClient, "directory")
	knowledge := knowledgehandler.New(services.WorkspaceClient, ingester, "directory")

	root := router

	// Runs
	root.Type(&v1.Run{}).FinalizeFunc(v1.RunFinalizer, runs.DeleteRunState)
	root.Type(&v1.Run{}).HandlerFunc(runs.Cleanup)

	// Threads
	root.Type(&v1.Thread{}).FinalizeFunc(v1.ThreadFinalizer, knowledge.RemoveWorkspace)
	root.Type(&v1.Thread{}).FinalizeFunc(v1.ThreadFinalizer, workspace.RemoveWorkspace)
	root.Type(&v1.Thread{}).HandlerFunc(threads.Cleanup)
	root.Type(&v1.Thread{}).HandlerFunc(knowledge.IngestKnowledge)

	// Workflows
	root.Type(&v1.Workflow{}).FinalizeFunc(v1.WorkflowFinalizer, workspace.RemoveWorkspace)
	root.Type(&v1.Workflow{}).FinalizeFunc(v1.WorkflowFinalizer, knowledge.RemoveWorkspace)
	root.Type(&v1.Workflow{}).HandlerFunc(knowledge.CreateWorkspace)

	// WorkflowExecutions
	root.Type(&v1.WorkflowExecution{}).FinalizeFunc(v1.WorkflowExecutionFinalizer, workspace.RemoveWorkspace)
	root.Type(&v1.WorkflowExecution{}).FinalizeFunc(v1.WorkflowExecutionFinalizer, knowledge.RemoveWorkspace)
	root.Type(&v1.WorkflowExecution{}).HandlerFunc(workflowExecution.Cleanup)
	// workspace creation is handled by Run, that's why it's not here, but cleanup is in the finalizers
	root.Type(&v1.WorkflowExecution{}).HandlerFunc(knowledge.CreateWorkspace)
	root.Type(&v1.WorkflowExecution{}).HandlerFunc(knowledge.IngestKnowledge)
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
	root.Type(&v1.Agent{}).FinalizeFunc(v1.AgentFinalizer, workspace.RemoveWorkspace)
	root.Type(&v1.Agent{}).FinalizeFunc(v1.AgentFinalizer, knowledge.RemoveWorkspace)
	root.Type(&v1.Agent{}).HandlerFunc(agents.Suggestion)
	root.Type(&v1.Agent{}).HandlerFunc(workspace.CreateWorkspace)
	root.Type(&v1.Agent{}).HandlerFunc(knowledge.CreateWorkspace)
	root.Type(&v1.Agent{}).HandlerFunc(knowledge.IngestKnowledge)

	// Slugs
	root.Type(&v1.Slug{}).HandlerFunc(slugs.SlugGC)
	root.Type(&v1.Agent{}).HandlerFunc(slugs.AssociateWithSlug)
	root.Type(&v1.Workflow{}).HandlerFunc(slugs.AssociateWithSlug)

	return nil
}
