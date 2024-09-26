package cli

import (
	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/cli/events"
	"github.com/spf13/cobra"
)

type ThreadPrint struct {
	root    *Otto
	Quiet   bool `usage:"Only print response content of threads" short:"q"`
	Verbose bool `usage:"Print more information" short:"v"`
	Follow  bool `usage:"Follow the thread events" short:"f"`
}

func (l *ThreadPrint) Customize(cmd *cobra.Command) {
	cmd.Use = "print [flags] [THREAD_ID]"
	cmd.Args = cobra.ExactArgs(1)
}

func (l *ThreadPrint) Run(cmd *cobra.Command, args []string) error {
	var (
		printer = events.NewPrinter(l.Quiet, l.Verbose)
	)

	events, err := l.root.Client.ThreadEvents(cmd.Context(), args[0], client.ThreadEventsOptions{
		Follow: l.Follow,
	})
	if err != nil {
		return err
	}

	return printer.Print("", events)
}
