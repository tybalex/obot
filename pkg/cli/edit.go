package cli

import (
	"fmt"

	"github.com/gptscript-ai/otto/pkg/cli/edit"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type Edit struct {
	root   *Otto
	Prompt bool `usage:"Edit just the prompt for the agent" short:"p"`
}

func (l *Edit) Customize(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(1)
}

func (l *Edit) Run(cmd *cobra.Command, args []string) error {
	agent, err := l.root.Client.GetAgent(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if l.Prompt {
		err = edit.Edit(cmd.Context(), []byte(agent.Prompt), ".txt", func(data []byte) error {
			agent.Prompt = v1.Body(data)
			_, err := l.root.Client.UpdateAgent(cmd.Context(), agent.ID, agent.AgentManifest)
			return err
		})
		if err != nil {
			return err
		}
		fmt.Printf("Agent updated: %s\n", agent.ID)
		return nil
	}

	data, err := yaml.Marshal(agent.AgentManifest)
	if err != nil {
		return err
	}

	err = edit.Edit(cmd.Context(), data, ".yaml", func(data []byte) error {
		var newManifest v1.AgentManifest
		if err := yaml.Unmarshal(data, &newManifest); err != nil {
			return err
		}

		_, err := l.root.Client.UpdateAgent(cmd.Context(), agent.ID, newManifest)
		return err
	})
	if err != nil {
		return err
	}
	fmt.Printf("Agent updated: %s\n", agent.ID)
	return nil
}
