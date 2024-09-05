package cli

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Threads struct {
	root *Otto
}

func (l *Threads) Customize(cmd *cobra.Command) {
	cmd.Use = "threads [flags] AGENT_ID"
	cmd.Args = cobra.MaximumNArgs(1)
}

func (l *Threads) Run(cmd *cobra.Command, args []string) error {
	var opts []client.ListThreadsOptions
	if len(args) > 0 {
		opts = append(opts, client.ListThreadsOptions{
			AgentID: args[0],
		})
	}
	threads, err := l.root.client.ListThreads(cmd.Context(), opts...)
	if err != nil {
		return err
	}

	w := newTable("ID", "AGENT", "STATE", "INPUT", "CREATED")
	maxLength := pterm.GetTerminalWidth() / 3
	for _, thread := range threads.Items {
		thread.Input = strings.Split(thread.Input, "\n")[0]

		if len(thread.Input) > maxLength {
			thread.Input = thread.Input[:maxLength] + "..."
		}
		state := "running"
		if thread.LastRunState != "running" {
			state = "waiting"
		}
		w.WriteRow(thread.ID, thread.AgentID, state, thread.Input, humanize.Time(thread.Created))
	}

	return w.Err()
}
