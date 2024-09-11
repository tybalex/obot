package cli

import (
	"fmt"
	"os"

	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type Update struct {
	root *Otto
}

func (l *Update) Customize(cmd *cobra.Command) {
	cmd.Use = "update [flags] [ID] [MANIFEST_FILE]"
	cmd.Args = cobra.ExactArgs(2)
}

func (l *Update) Run(cmd *cobra.Command, args []string) error {
	id := args[0]
	data, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}

	var newManifest v1.AgentManifest
	if err := yaml.Unmarshal(data, &newManifest); err != nil {

	}

	agent, err := l.root.Client.UpdateAgent(cmd.Context(), id, newManifest)
	if err != nil {
		return err
	}

	fmt.Printf("Agent updated: %s invoke: %s\n", agent.ID, agent.Links["invoke"])
	return nil
}
