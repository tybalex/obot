package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/cli/textio"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Invoke struct {
	Thread string `usage:"Thread name to run the agent in." short:"t"`
	Quiet  *bool  `usage:"Only print output characters" short:"q"`
	root   *Otto
}

func (l *Invoke) PersistentPre(cmd *cobra.Command, args []string) error {
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
	input := strings.Join(args[1:], " ")

	resp, err := l.root.Client.Invoke(cmd.Context(), args[0], input, client.InvokeOptions{
		ThreadID: l.Thread,
	})
	if err != nil {
		return err
	}

	if l.Quiet != nil && *l.Quiet {
		var lastContent string
		for event := range resp.Events {
			if event.Error != "" {
				return fmt.Errorf("%s", event.Error)
			}
			if event.Content != "" {
				lastContent = event.Content
				fmt.Print(event.Content)
			}
		}
		if !strings.HasSuffix(lastContent, "\n") {
			fmt.Println()
		}
		return nil
	}

	var (
		spinner              = textio.NewSpinnerPrinter()
		printGeneratingInput bool
	)
	spinner.Start()
	defer spinner.Stop()

	spinner.Print(fmt.Sprintf("> Input: %s\n", input))
	spinner.EnsureNewline()
outer:
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			spinner.Tick()
		case event, ok := <-resp.Events:
			if !ok {
				break outer
			}

			if event.WaitingOnModel {
				spinner.EnsureNewline()
				spinner.Print("> Waiting for model... \n")
			}

			if event.Error != "" {
				spinner.EnsureNewline()
				spinner.Stop()
				log.Errorf("%s", event.Error)
				spinner.Start()
			}

			if event.Tool.PartialInput != "" {
				if !printGeneratingInput {
					spinner.Print(fmt.Sprintf("> Generating tool input for (%s)...  ", event.Tool.GeneratingInputForName))
					printGeneratingInput = true
				}
				spinner.Print(event.Tool.PartialInput)
			}

			if event.Content != "" {
				if printGeneratingInput {
					spinner.Print("\n")
					printGeneratingInput = false
				}
				spinner.Print(event.Content)
			}

			if event.Tool.Name != "" {
				if printGeneratingInput {
					spinner.Print("\n")
					printGeneratingInput = false
				}
				spinner.Print(fmt.Sprintf("> Running tool (%s): %s\n", event.Tool.Name, event.Tool.Input))
			}
		}
	}

	spinner.EnsureNewline()
	return nil
}
