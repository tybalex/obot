package invokeclient

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/api/types"
)

type responsePrinter interface {
	Print(input string, resp *types.InvokeResponse) error
}

type inputter interface {
	Next(ctx context.Context, previous string, resp *types.InvokeResponse) (string, bool, error)
}

type Options struct {
	ThreadID   string
	Quiet      bool
	EmptyInput bool
}

func Invoke(ctx context.Context, c *client.Client, id, input string, opts Options) (err error) {
	var (
		printer  responsePrinter = &Verbose{}
		inputter inputter        = VerboseInputter{
			client: c,
		}
	)
	if opts.Quiet {
		printer = &Quiet{}
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
			ThreadID: opts.ThreadID,
		})
		if err != nil {
			return err
		}

		if err := printer.Print(input, resp); err != nil {
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
