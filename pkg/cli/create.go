package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/cli/textio"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

type Create struct {
	Quiet            bool              `usage:"Only print ID after successful creation." short:"q"`
	Name             string            `usage:"Name of the agent."`
	Description      string            `usage:"Description of the agent."`
	Ref              string            `usage:"The path segment of the agent in the published URL (defaults to ID of agent)."`
	Tools            []string          `usage:"List of tools the agent can use."`
	CodeDependencies string            `usage:"The code dependencies content for the agent if it using JavaScript (package.json) or Python (requirements.txt)."`
	Steps            []string          `usage:"The steps for a workflow."`
	Output           string            `usage:"The output for a workflow."`
	Params           map[string]string `usage:"The parameters for the agent." hidden:"true"`
	File             string            `usage:"The file to read the agent manifest from." short:"f"`
	Replace          bool              `usage:"If loading from file, replace the agent with the same refName if it exists." short:"r"`
	root             *Otto
}

func (l *Create) Customize(cmd *cobra.Command) {
	cmd.Use = "create [flags]"
}

func parseManifests(data []byte) (result []v1.WorkflowManifest, _ error) {
	var (
		dec = yamlv3.NewDecoder(bytes.NewReader(data))
	)
	for {
		parsed := map[string]any{}
		if err := dec.Decode(&parsed); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		jsonData, err := json.Marshal(parsed)
		if err != nil {
			return nil, err
		}

		var manifest v1.WorkflowManifest
		if err := json.Unmarshal(jsonData, &manifest); err != nil {
			return nil, err
		}
		result = append(result, manifest)
	}

	return
}

func (l *Create) loadFromFile(ctx context.Context) error {
	var (
		data []byte
		err  error
	)

	if strings.HasPrefix(l.File, "http://") || strings.HasPrefix(l.File, "https://") {
		resp, err := http.Get(l.File)
		if err != nil {
			return err
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	} else {
		data, err = os.ReadFile(l.File)
		if err != nil {
			return err
		}
	}

	manifests, err := parseManifests(data)
	if err != nil {
		return err
	}

	for _, newManifest := range manifests {
		if len(newManifest.Steps) > 0 || newManifest.Output != "" {
			if l.Replace && l.Ref != "" {
				workflows, err := l.root.Client.ListWorkflows(ctx, client.ListWorkflowsOptions{
					RefName: l.Ref,
				})
				if err != nil {
					return err
				}
				if len(workflows.Items) > 0 {
					_, err = l.root.Client.UpdateWorkflow(ctx, workflows.Items[0].ID, newManifest)
					return err
				}
			}

			workflow, err := l.root.Client.CreateWorkflow(ctx, newManifest)
			if err != nil {
				return err
			}

			if l.Quiet {
				fmt.Println(workflow.ID)
			} else {
				fmt.Printf("Workflow created: %s\n", workflow.ID)
			}
		} else {
			if l.Replace && l.Ref != "" {
				agents, err := l.root.Client.ListAgents(ctx, client.ListAgentsOptions{
					RefName: l.Ref,
				})
				if err != nil {
					return err
				}
				if len(agents.Items) > 0 {
					_, err = l.root.Client.UpdateAgent(ctx, agents.Items[0].ID, newManifest.AgentManifest)
					return err
				}
			}

			agent, err := l.root.Client.CreateAgent(ctx, newManifest.AgentManifest)
			if err != nil {
				return err
			}

			if l.Quiet {
				fmt.Println(agent.ID)
			} else {
				fmt.Printf("Agent created: %s\n", agent.ID)
			}
		}
	}

	return nil
}

func (l *Create) Run(cmd *cobra.Command, args []string) error {
	if l.File != "" {
		return l.loadFromFile(cmd.Context())
	}

	prompt := strings.Join(args, " ")
	if prompt == "" && !l.Quiet {
		textio.Info("Creating an agent")
		fmt.Println()
		textio.Print("You are about to create an agent. An agent is AI that will respond to user " +
			"input according to the instructions you provide.")
		fmt.Println()
		result, err := textio.Ask("Enter the instructions",
			"You're a friendly assistant")
		if err != nil {
			return err
		}
		prompt = result
	}

	agentManifest := v1.AgentManifest{
		Name:             l.Name,
		Description:      l.Description,
		RefName:          l.Ref,
		Prompt:           v1.Body(prompt),
		Tools:            l.Tools,
		CodeDependencies: l.CodeDependencies,
		Params:           l.Params,
	}

	var (
		id, link string
	)
	if l.Output != "" || len(l.Steps) > 0 {
		wf := v1.WorkflowManifest{
			AgentManifest: agentManifest,
			Output:        l.Output,
		}
		for _, step := range l.Steps {
			wf.Steps = append(wf.Steps, v1.Step{
				Step: step,
			})
		}
		workflow, err := l.root.Client.CreateWorkflow(cmd.Context(), wf)
		if err != nil {
			return err
		}
		id, link = workflow.ID, workflow.Links["invoke"]
	} else {
		agent, err := l.root.Client.CreateAgent(cmd.Context(), agentManifest)
		if err != nil {
			return err
		}
		id, link = agent.ID, agent.Links["invoke"]
	}

	if l.Quiet {
		fmt.Println(id)
	} else {
		fmt.Println()
		textio.Info(fmt.Sprintf("Agent created."))
		textio.Info(fmt.Sprintf(""))
		textio.Info(fmt.Sprintf("ID: %s", id))
		textio.Info(fmt.Sprintf("URL: %s", link))
		fmt.Println()
		textio.Print(fmt.Sprintf("You can now interact with your new agent by running:"))
		fmt.Println()
		fmt.Printf("CLI:  otto invoke %s Hello\n", id)
		fmt.Printf("cURL: curl -d Hello %s\n", link)
	}
	return nil
}
