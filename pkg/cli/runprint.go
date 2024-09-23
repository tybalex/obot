package cli

import (
	"github.com/gptscript-ai/otto/pkg/cli/events"
	"github.com/spf13/cobra"
)

type RunPrint struct {
	root  *Otto
	Quiet bool `usage:"Only print the response content of the runs" short:"q"`
}

func (l *RunPrint) Customize(cmd *cobra.Command) {
	cmd.Use = "print [flags] [RUN_ID]"
	cmd.Args = cobra.ExactArgs(1)
}

func (l *RunPrint) Run(cmd *cobra.Command, args []string) error {
	debug, err := l.root.Client.RunEvents(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	printer := events.NewPrinter(l.Quiet)
	return printer.Print("", debug)
}
