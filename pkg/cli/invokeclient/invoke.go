package invokeclient

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/gptscript-ai/otto/pkg/cli/events"
)

type inputter interface {
	Next(ctx context.Context, previous string, resp *types.InvokeResponse) (string, bool, error)
}

type Options struct {
	ThreadID string
	Quiet    bool
	Async    bool
}

func Invoke(ctx context.Context, c *client.Client, id, input string, opts Options) (err error) {
	var (
		printer           = events.NewPrinter(opts.Quiet)
		inputter inputter = VerboseInputter{
			client: c,
		}
		threadID = opts.ThreadID
	)
	if opts.Quiet {
		inputter = QuietInputter{}
	}

	input, ok, err := inputter.Next(ctx, input, nil)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no input provided")
	}

	for {
		resp, err := c.Invoke(ctx, id, input, client.InvokeOptions{
			ThreadID: threadID,
			Async:    opts.Async,
		})
		if err != nil {
			return err
		}

		threadID = resp.ThreadID

		if opts.Async {
			if opts.Quiet {
				fmt.Println(threadID)
			} else {
				fmt.Printf("Thread ID: %s\n", threadID)
			}
			return nil
		}

		if err := printer.Print(input, resp.Events); err != nil {
			return err
		}

		nextInput, cont, err := inputter.Next(ctx, input, resp)
		if err != nil {
			return err
		} else if !cont {
			break
		}

		input = nextInput
	}

	return nil
}
