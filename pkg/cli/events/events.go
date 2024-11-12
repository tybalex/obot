package events

import (
	"context"

	"github.com/otto8-ai/otto8/apiclient"
	"github.com/otto8-ai/otto8/apiclient/types"
)

type Printer interface {
	Print(input string, events <-chan types.Progress) error
}

func NewPrinter(ctx context.Context, c *apiclient.Client, quiet, details bool) Printer {
	if quiet {
		return &Quiet{
			Client: c,
			Ctx:    ctx,
		}
	}
	return &Verbose{
		Details: details,
		Client:  c,
		Ctx:     ctx,
	}
}
