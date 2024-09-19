package cli

import (
	"os"
	"strings"

	"github.com/gptscript-ai/otto/pkg/cli/invokeclient"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Invoke struct {
	Thread string `usage:"Thread name to run the agent in." short:"t"`
	Quiet  *bool  `usage:"Only print output characters" short:"q"`
	root   *Otto
}

func (l *Invoke) GetQuiet() bool {
	if l.Quiet == nil {
		return false
	}
	return *l.Quiet
}

func (l *Invoke) Pre(cmd *cobra.Command, args []string) error {
	if l.Quiet == nil && term.IsTerminal(int(os.Stdout.Fd())) {
		l.Quiet = new(bool)
	}
	return nil
}

func (l *Invoke) Customize(cmd *cobra.Command) {
	cmd.Use = "invoke [flags] AGENT [INPUT...]"
	cmd.Args = cobra.MinimumNArgs(1)
}

func (l *Invoke) Run(cmd *cobra.Command, args []string) error {
	return invokeclient.Invoke(cmd.Context(), l.root.Client, args[0], strings.Join(args[1:], " "), invokeclient.Options{
		ThreadID: l.Thread,
		Quiet:    l.GetQuiet(),
	})
}
