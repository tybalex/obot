package cli

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/gptscript-ai/otto/apiclient"
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
	var opts []apiclient.ListThreadsOptions
	if len(args) > 0 {
		opts = append(opts, apiclient.ListThreadsOptions{
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

	w := newTable("ID", "DESC", "AGENT/WF", "LASTRUN", "PREVTHREAD", "STATE", "CREATED")
	for _, thread := range threads.Items {
		agentWF := thread.AgentID
		if agentWF == "" {
			agentWF = thread.WorkflowID
		}
		w.WriteRow(thread.ID, thread.Description, agentWF, thread.LastRunID, thread.PreviousThreadID, string(thread.LastRunState), humanize.Time(thread.Created.Time))
	}

	return w.Err()
}
