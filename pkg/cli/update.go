package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

	agent, err := l.root.client.UpdateAgent(cmd.Context(), id, data)
	if err != nil {
		return err
	}

	fmt.Printf("Agent updated: %s invoke: %s\n", agent.ID, agent.Links["invoke"])
	return nil
}
