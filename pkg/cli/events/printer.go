package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gptscript-ai/otto/apiclient/types"
	"github.com/gptscript-ai/otto/logger"
	"github.com/gptscript-ai/otto/pkg/cli/textio"
)

var log = logger.Package()

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
	Details bool
}

func (v *Verbose) Print(input string, events <-chan types.Progress) error {
	var (
		out       = textio.NewSpinnerPrinter()
		lastType  string
		lastRunID string
	)
	out.Start()
	defer out.Stop()

outer:
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			out.Tick()
		case event, ok := <-events:
			if !ok {
				break outer
			}

			if event.RunID != "" && event.RunID != lastRunID && v.Details {
				lastType = "run"
				out.EnsureNewline()
				out.Print(fmt.Sprintf("\n> Run ID: %s\n", event.RunID))
				lastRunID = event.RunID
			}

			if event.Step != nil && v.Details {
				lastType = "step"
				out.EnsureNewline()
				out.Print(event.Step.Display())
			} else if event.Input != "" {
				out.EnsureNewline()
				if lastType == "content" {
					out.Print("\n")
				}
				lastType = "input"
				out.Print(fmt.Sprintf("> %s\n", color.GreenString(event.Input)))
			} else if event.WaitingOnModel {
				lastType = "waiting"
				out.EnsureNewline()
				out.Print("> Waiting for model... \n")
			} else if event.Error != "" {
				lastType = "error"
				out.EnsureNewline()
				out.Stop()
				log.Errorf("%s", event.Error)
				out.Start()
			} else if event.ToolInput != nil {
				if lastType != "toolInput" {
					out.Print(fmt.Sprintf("> Generating tool input for (%s)...  ", event.ToolInput.Name))
				}
				lastType = "toolInput"
				out.Print(event.ToolInput.Input)
			} else if event.ToolCall != nil {
				out.EnsureNewline()
				out.Print(fmt.Sprintf("> Running tool (%s): %s\n", color.MagentaString(event.ToolCall.Name), color.MagentaString(event.ToolCall.Input)))
			} else if event.Content != "" {
				if lastType != "content" {
					out.EnsureNewline()
					out.Print("\n")
				}
				lastType = "content"
				out.Print(color.CyanString(event.Content))
			}
		}
	}

	out.EnsureNewline()
	return nil
}
