package cli

import (
	"fmt"
	"os"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/spf13/cobra"
)

type Create struct {
	Replace bool `usage:"Replace the existing agent with the same name, if it exists."`
	Quiet   bool `usage:"Only print ID after successful creation." short:"q"`
	root    *Otto
}

func (l *Create) Customize(cmd *cobra.Command) {
	cmd.Use = "create [flags] MANIFEST_FILE"
	cmd.Args = cobra.ExactArgs(1)
}

func (l *Create) Run(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	var replaced bool
	agent, err := l.root.client.CreateAgent(cmd.Context(), data, client.CreateOptions{
		Replace: l.Replace,
		ReplacedCallback: func() {
			replaced = true
		},
	})
	if err != nil {
		return err
	}

	if l.Quiet {
		fmt.Println(agent.ID)
	} else if replaced {
		fmt.Printf("Agent replaced: %s invoke: %s\n", agent.ID, agent.Links["invoke"])
	} else {
		fmt.Printf("Agent created: %s invoke: %s\n", agent.ID, agent.Links["invoke"])
	}
	return nil
}
