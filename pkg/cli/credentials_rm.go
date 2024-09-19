package cli

import (
	"fmt"
	"strings"

	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/system"
	"github.com/spf13/cobra"
)

type CredentialsRm struct {
	root   *Otto
	Wide   bool `usage:"Print more information" short:"w"`
	Quiet  bool `usage:"Only print IDs of credentials" short:"q"`
	Follow bool `usage:"Follow the output of credentials" short:"f"`
}

func (l *CredentialsRm) Customize(cmd *cobra.Command) {
	cmd.Use = "rm [flags] [CRED_NAME...]"
	cmd.Aliases = []string{"delete", "del", "d"}
}

func (l *CredentialsRm) Run(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		opt := client.DeleteCredentialsOptions{}
		scope, name, ok := strings.Cut(arg, "::")
		if !ok {
			name = scope
		} else if system.IsThreadID(scope) {
			opt.ThreadID = scope
		} else if system.IsAgentID(scope) {
			opt.AgentID = scope
		} else if system.IsWorkflowID(scope) {
			opt.WorkflowID = scope
		}
		if err := l.root.Client.DeleteCredential(cmd.Context(), name, opt); err != nil {
			return err
		}
		if !l.Quiet {
			fmt.Printf("Credential %s deleted\n", arg)
		}
	}
	return nil
}
