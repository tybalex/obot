package events

import (
	"github.com/gptscript-ai/otto/apiclient/types"
)

type Printer interface {
	Print(input string, events <-chan types.Progress) error
}

func NewPrinter(quiet, details bool) Printer {
	if quiet {
		return &Quiet{}
	}
	return &Verbose{
		Details: details,
	}
}
