package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Agents struct {
	root  *Otto
	Quiet bool `usage:"Only print IDs of agents" short:"q"`
	Wide  bool `usage:"Print more information" short:"w"`
}

func (l *Agents) Customize(cmd *cobra.Command) {
}

func (l *Agents) Run(cmd *cobra.Command, args []string) error {
	agents, err := l.root.Client.ListAgents(cmd.Context())
	if err != nil {
		return err
	}

	if l.Quiet {
		for _, agent := range agents.Items {
			fmt.Println(agent.ID)
		}
		return nil
	}

	w := newTable("ID", "NAME", "DESCRIPTION", "INVOKE")
	for _, agent := range agents.Items {
		w.WriteRow(agent.ID, agent.Name, truncate(agent.Description, l.Wide), agent.Links["invoke"])
	}

	return w.Err()
}
