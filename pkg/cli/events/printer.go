package events

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

func (q *Quiet) Print(input string, events <-chan types.Progress) error {
	var lastContent string
	for event := range events {
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

func (v *Verbose) Print(input string, events <-chan types.Progress) error {
	var (
		spinner              = textio.NewSpinnerPrinter()
		printGeneratingInput bool
		lastType             string
		lastRunID            string
	)
	spinner.Start()
	defer spinner.Stop()

outer:
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			spinner.Tick()
		case event, ok := <-events:
			if !ok {
				break outer
			}

			if event.RunID != lastRunID {
				lastType = "run"
				spinner.EnsureNewline()
				spinner.Print(fmt.Sprintf("\n> Run ID: %s\n", event.RunID))
				lastRunID = event.RunID
			}

			if event.Step != nil {
				lastType = "step"
				spinner.EnsureNewline()
				spinner.Print("\n")
				spinner.Print(event.Step.Display())
			} else if event.Input != "" {
				if lastType != "step" && lastType != "run" {
					spinner.Print("\n")
				}
				lastType = "input"
				spinner.EnsureNewline()
				spinner.Print(fmt.Sprintf("> Input: %s\n", event.Input))
			} else if event.WaitingOnModel {
				lastType = "waiting"
				spinner.EnsureNewline()
				spinner.Print("> Waiting for model... \n")
			} else if event.Error != "" {
				lastType = "error"
				spinner.EnsureNewline()
				spinner.Stop()
				log.Errorf("%s", event.Error)
				spinner.Start()
			} else if event.Tool.PartialInput != "" {
				lastType = "tool"
				if !printGeneratingInput {
					spinner.Print(fmt.Sprintf("> Generating tool input for (%s)...  ", event.Tool.GeneratingInputForName))
					printGeneratingInput = true
				}
				spinner.Print(event.Tool.PartialInput)
			} else if event.Tool.Name != "" {
				lastType = "tool"
				if printGeneratingInput {
					spinner.Print("\n")
					printGeneratingInput = false
				}
				spinner.Print(fmt.Sprintf("> Running tool (%s): %s\n", event.Tool.Name, event.Tool.Input))
			} else if event.Content != "" {
				lastType = "content"
				if printGeneratingInput {
					spinner.Print("\n")
				}
				spinner.Print(event.Content)
			}
		}
	}

	spinner.EnsureNewline()
	return nil
}
