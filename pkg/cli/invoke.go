package cli

import (
	"fmt"
	"strings"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/spf13/cobra"
)

type Invoke struct {
	Thread string `usage:"Thread name to run the agent in." short:"t"`
	root   *Otto
}

func (l *Invoke) Customize(cmd *cobra.Command) {
	cmd.Use = "invoke [flags] AGENT [INPUT...]"
	cmd.Args = cobra.MinimumNArgs(1)
}

func (l *Invoke) Run(cmd *cobra.Command, args []string) error {
	resp, err := l.root.Client.Invoke(cmd.Context(), args[0], strings.Join(args[1:], " "), client.InvokeOptions{
		ThreadID: l.Thread,
	})
	if err != nil {
		return err
	}

	var buf strings.Builder
	for event := range resp.Events {
		buf.WriteString(event.Content)
		fmt.Print(event.Content)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		fmt.Println()
	}
	return nil
}
