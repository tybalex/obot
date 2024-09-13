package render

import (
	"context"
	"fmt"
	"sort"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/storage"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/gptscript-ai/otto/pkg/workspace"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultAgentParams = []string{
	"message", "Message to send to the agent",
}

type AgentOptions struct {
	Thread        *v1.Thread
	KnowledgeTool string
}

func agentKnowledgeEnv(agent *v1.Agent, thread *v1.Thread) (envs []string) {
	if agent.GetKnowledgeWorkspaceStatus().HasKnowledge {
		envs = append(envs, fmt.Sprintf("GPTSCRIPT_SCRIPT_ID=%s",
			workspace.KnowledgeIDFromWorkspaceID(agent.GetKnowledgeWorkspaceStatus().KnowledgeWorkspaceID)))
		if thread != nil && thread.GetKnowledgeWorkspaceStatus().HasKnowledge {
			envs = append(envs, fmt.Sprintf("OTTO_THREAD_ID=%s", workspace.KnowledgeIDFromWorkspaceID(thread.GetKnowledgeWorkspaceStatus().KnowledgeWorkspaceID)))
		}
	}
	return envs
}

func Agent(ctx context.Context, db storage.Client, agent *v1.Agent, opts AgentOptions) (_ []gptscript.ToolDef, extraEnv []string, _ error) {
	t := []gptscript.ToolDef{{
		Name:         agent.Spec.Manifest.Name,
		Description:  agent.Spec.Manifest.Description,
		Chat:         true,
		Tools:        agent.Spec.Manifest.Tools,
		Instructions: agent.Spec.Manifest.Prompt.Instructions(),
		MetaData:     agent.Spec.Manifest.Prompt.Metadata(agent.Spec.Manifest.CodeDependencies),
		Temperature:  agent.Spec.Manifest.Temperature,
		Cache:        agent.Spec.Manifest.Cache,
		Type:         "agent",
	}}

	if envs := agentKnowledgeEnv(agent, opts.Thread); len(envs) > 0 {
		extraEnv = envs
		if opts.KnowledgeTool != "" {
			t[0].Tools = append(t[0].Tools, opts.KnowledgeTool)
		}
	}

	if len(agent.Spec.Manifest.Agents) == 0 {
		return t, extraEnv, nil
	}

	agents, err := ByName(ctx, db, agent.Namespace)
	if err != nil {
		return nil, nil, err
	}

	for _, agentRef := range agent.Spec.Manifest.Agents {
		agent, ok := agents[agentRef]
		if !ok {
			continue
		}

		toolDef := gptscript.ToolDef{
			Name:        agent.Spec.Manifest.Name,
			Description: agent.Spec.Manifest.Description,
			Arguments:   agent.Spec.Manifest.GetParams(),
		}
		if toolDef.Name == "" {
			toolDef.Name = agentRef
		}
		if toolDef.Description == "" {
			toolDef.Description = "Send message to agent named " + toolDef.Name
		}
		if len(agent.Spec.Manifest.Params) == 0 {
			toolDef.Arguments = gptscript.ObjectSchema(DefaultAgentParams...)
		}
		toolDef.Instructions = fmt.Sprintf(`#!${OTTO_BIN} invoke -t "${OTTO_THREAD_ID}.%s" "%s" "${GPTSCRIPT_INPUT}"`, agentRef, agent.Name)

		t[0].Tools = append(t[0].Tools, toolDef.Name)
		t = append(t, toolDef)
	}

	return t, extraEnv, nil
}

func ByName(ctx context.Context, db storage.Client, namespace string) (map[string]v1.Agent, error) {
	var agents v1.AgentList
	err := db.List(ctx, &agents, &kclient.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(agents.Items, func(i, j int) bool {
		return agents.Items[i].Name < agents.Items[i].Name
	})

	result := map[string]v1.Agent{}
	for _, agent := range agents.Items {
		result[agent.Name] = agent
		if _, ok := result[agent.Spec.Manifest.Slug]; !ok {
			result[agent.Spec.Manifest.Slug] = agent
		}
		if _, ok := result[agent.Spec.Manifest.Name]; !ok {
			result[agent.Spec.Manifest.Name] = agent
		}
	}

	return result, nil
}
