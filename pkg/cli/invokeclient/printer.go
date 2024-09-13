package invokeclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/cli/textio"
	"github.com/gptscript-ai/otto/pkg/mvl"
)

var log = mvl.Package()

type Quiet struct {
}

func (q *Quiet) Print(input string, resp *types.InvokeResponse) error {
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

type Verbose struct {
}

func (v *Verbose) Print(input string, resp *types.InvokeResponse) error {
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
