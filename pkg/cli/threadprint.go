package cli

import (
	"github.com/gptscript-ai/otto/apiclient"
	"github.com/gptscript-ai/otto/pkg/cli/events"
	"github.com/spf13/cobra"
)

type ThreadPrint struct {
	root    *Otto
	Quiet   bool `usage:"Only print response content of threads" short:"q"`
	Verbose bool `usage:"Print more information" short:"v"`
}

func (l *ThreadPrint) Customize(cmd *cobra.Command) {
	cmd.Use = "print [flags] [THREAD_ID]"
	cmd.Args = cobra.ExactArgs(1)
}

func (l *ThreadPrint) Run(cmd *cobra.Command, args []string) error {
	var (
		printer = events.NewPrinter(l.Quiet, l.Verbose)
	)

	events, err := l.root.Client.ThreadEvents(cmd.Context(), args[0], apiclient.ThreadEventsOptions{})
	if err != nil {
		return err
	}

	return printer.Print("", events)
}
