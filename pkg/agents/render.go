package agents

import (
	"context"
	"fmt"
	"sort"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/otto/pkg/storage"
	"github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultAgentParams = []string{
	"message", "Message to send to the agent",
}

func Render(ctx context.Context, db storage.Client, namespace string, manifest v1.Manifest) ([]gptscript.ToolDef, error) {
	t := []gptscript.ToolDef{{
		Name:         manifest.Name,
		Description:  manifest.Description,
		Chat:         true,
		Tools:        manifest.Tools,
		Arguments:    manifest.GetParams(),
		Instructions: manifest.Prompt,
		Type:         "agent",
	}}

	if len(manifest.Agents) == 0 {
		return t, nil
	}

	agents, err := ByName(ctx, db, namespace)
	if err != nil {
		return nil, err
	}

	for _, agentRef := range manifest.Agents {
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

	return t, nil
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
