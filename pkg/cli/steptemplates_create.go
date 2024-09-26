package cli

import (
	"fmt"

	"github.com/gptscript-ai/otto/pkg/api/types"
	v1 "github.com/gptscript-ai/otto/pkg/storage/apis/otto.gptscript.ai/v1"
	"github.com/spf13/cobra"
)

type StepTemplateCreate struct {
	root  *Otto
	Quiet bool `usage:"Only print IDs of created step template" short:"q"`
}

func (l *StepTemplateCreate) Customize(cmd *cobra.Command) {
	cmd.Use = "create [flags] NAME REFERENCE"
	cmd.Args = cobra.ExactArgs(2)
}

func (l *StepTemplateCreate) Run(cmd *cobra.Command, args []string) error {
	tr, err := l.root.Client.CreateToolReference(cmd.Context(), types.ToolReferenceManifest{
		Name:      args[0],
		ToolType:  v1.ToolReferenceTypeStepTemplate,
		Reference: args[1],
	})
	if err != nil {
		return err
	}
	if l.Quiet {
		fmt.Println(tr.ID)
	} else {
		fmt.Println("Step template created:", tr.ID)
	}
	return nil
}
