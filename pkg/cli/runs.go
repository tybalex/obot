package cli

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Runs struct {
	root  *Otto
	Wide  bool `usage:"Print more information" short:"w"`
	Quiet bool `usage:"Only print IDs of runs" short:"q"`
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
	runs, err := l.root.Client.ListRuns(cmd.Context(), opts...)
	if err != nil {
		return err
	}

	if l.Quiet {
		for _, run := range runs.Items {
			cmd.Println(run.ID)
		}
		return nil
	}

	w := newTable("ID", "AGENT", "THREAD", "STATE", "INPUT", "OUTPUT", "CREATED")
	for _, run := range runs.Items {
		run.Input = truncate(strings.Split(run.Input, "\n")[0], l.Wide)
		run.Output = truncate(strings.Split(run.Output, "\n")[0], l.Wide)
		run.Error = truncate(strings.Split(run.Error, "\n")[0], l.Wide)
		if run.Error != "" {
			run.Output = run.Error
		}
		w.WriteRow(run.ID, run.AgentID, run.ThreadID, string(run.State), run.Input, run.Output, humanize.Time(run.Created))
	}

	return w.Err()
}

func truncate(text string, wide bool) string {
	if wide {
		return text
	}
	maxLength := pterm.GetTerminalWidth() / 3
	if len(text) > maxLength {
		return text[:maxLength] + "..."
	}
	return text
}
