package cli

import (
	"github.com/spf13/cobra"
)

type Agents struct {
	root *Otto
}

func (l *Agents) Run(cmd *cobra.Command, args []string) error {
	agents, err := l.root.client.ListAgents(cmd.Context())
	if err != nil {
		return err
	}

	w := newTable("ID", "NAME", "DESCRIPTION", "INVOKE")
	for _, agent := range agents.Items {
		w.WriteRow(agent.ID, agent.Name, agent.Description, agent.Links["invoke"])
	}

	return w.Err()
}
