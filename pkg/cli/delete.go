package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Delete struct {
	root *Otto
}

func (l *Delete) Customize(cmd *cobra.Command) {
	cmd.Use = "delete [flags] ID"
	cmd.Args = cobra.ExactArgs(1)
}

func (l *Delete) Run(cmd *cobra.Command, args []string) error {
	id := args[0]

	if err := l.root.client.DeleteAgent(cmd.Context(), id); err != nil {
		return err
	}

	fmt.Printf("Agent deleted: %s\n", id)
	return nil
}
