package cli

import (
	"fmt"

	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/spf13/cobra"
)

type StepTemplatesDelete struct {
	root  *Otto
	Quiet bool `usage:"Only print IDs of deleted step templates" short:"q"`
}

func (l *StepTemplatesDelete) Customize(cmd *cobra.Command) {
	cmd.Use = "delete [flags] [ID...]"
	cmd.Aliases = []string{"rm", "del", "d"}
}

func (l *StepTemplatesDelete) Run(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if err := l.root.Client.DeleteToolReference(cmd.Context(), arg, v1.ToolReferenceTypeStepTemplate); err != nil {
			return err
		}
		if l.Quiet {
			fmt.Println(arg)
		} else {
			fmt.Println("Step template deleted:", arg)
		}
	}
	return nil
}
