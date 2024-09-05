package cli

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Runs struct {
	root *Otto
}

func (l *Runs) Customize(cmd *cobra.Command) {
	cmd.Use = "runs [flags] [AGENT_ID] [THREAD_ID]"
	cmd.Args = cobra.MaximumNArgs(2)
}

func (l *Runs) Run(cmd *cobra.Command, args []string) error {
	var opts []client.ListRunsOptions
	if len(args) > 0 {
		opts = append(opts, client.ListRunsOptions{
			AgentID: args[0],
		})
	}
	if len(args) > 1 {
		opts = append(opts, client.ListRunsOptions{
			ThreadID: args[1],
		})
	}
	runs, err := l.root.client.ListRuns(cmd.Context(), opts...)
	if err != nil {
		return err
	}

	w := newTable("ID", "AGENT", "THREAD", "STATE", "INPUT", "OUTPUT", "CREATED")
	maxLength := pterm.GetTerminalWidth() / 5
	for _, run := range runs.Items {
		run.Input = strings.Split(run.Input, "\n")[0]
		run.Output = strings.Split(run.Output, "\n")[0]
		run.Error = strings.Split(run.Error, "\n")[0]

		if len(run.Input) > maxLength {
			run.Input = run.Input[:maxLength] + "..."
		}
		if run.Error != "" {
			run.Output = run.Error
		}
		if len(run.Output) > maxLength {
			run.Output = run.Output[:maxLength] + "..."
		}
		w.WriteRow(run.ID, run.AgentID, run.ThreadID, string(run.State), run.Input, run.Output, humanize.Time(run.Created))
	}

	return w.Err()
}
