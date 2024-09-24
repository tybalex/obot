package events

import "github.com/gptscript-ai/otto/pkg/api/types"

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
