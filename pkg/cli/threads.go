package cli

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/spf13/cobra"
)

type Threads struct {
	root  *Otto
	Quiet bool `usage:"Only print IDs of threads" short:"q"`
	Wide  bool `usage:"Print more information" short:"w"`
}

func (l *Threads) Customize(cmd *cobra.Command) {
	cmd.Use = "threads [flags] AGENT_ID"
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Aliases = []string{"thread", "t"}
}

func (l *Threads) Run(cmd *cobra.Command, args []string) error {
	var opts []client.ListThreadsOptions
	if len(args) > 0 {
		opts = append(opts, client.ListThreadsOptions{
			AgentID: args[0],
		})
	}
	threads, err := l.root.Client.ListThreads(cmd.Context(), opts...)
	if err != nil {
		return err
	}

	if l.Quiet {
		for _, thread := range threads.Items {
			fmt.Println(thread.ID)
		}
		return nil
	}

	w := newTable("ID", "AGENT", "STATE", "INPUT", "CREATED")
	for _, thread := range threads.Items {
		thread.Input = truncate(strings.Split(thread.Input, "\n")[0], l.Wide)
		state := "running"
		if thread.LastRunState != "running" {
			state = "waiting"
		}
		w.WriteRow(thread.ID, thread.AgentID, state, thread.Input, humanize.Time(thread.Created))
	}

	return w.Err()
}
