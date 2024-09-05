package cli

import (
	"github.com/gptscript-ai/otto/pkg/controller"
	"github.com/spf13/cobra"
)

type Controller struct {
	controller.Options
}

func (c Controller) Run(cmd *cobra.Command, args []string) error {
	ctr, err := controller.New(cmd.Context(), c.Options)
	if err != nil {
		return err
	}
	return ctr.Start(cmd.Context())
}
