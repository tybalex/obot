package cli

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gptscript-ai/otto/pkg/api/client"
	"github.com/gptscript-ai/otto/pkg/api/types"
	"github.com/spf13/cobra"
)

type Credentials struct {
	root     *Otto
	Wide     bool   `usage:"Print more information" short:"w"`
	Quiet    bool   `usage:"Only print IDs of credentials" short:"q"`
	Follow   bool   `usage:"Follow the output of credentials" short:"f"`
	ThreadID string `usage:"Specific thread list credentials for" short:"t"`
}

func (l *Credentials) Customize(cmd *cobra.Command) {
	cmd.Use = "credentials [flags]"
	cmd.Aliases = []string{"credential", "cred", "creds", "c"}
}

func (l *Credentials) printCredentialsQuiet(i types.CredentialList) error {
	for _, credential := range i.Items {
		fmt.Println(credential.Name)
	}
	return nil
}

func (l *Credentials) printCredentials(i types.CredentialList) error {
	w := newTable("NAME", "ENV", "EXPIRES")
	for _, credential := range i.Items {
		time := "never"
		if credential.ExpiresAt != nil {
			time = humanize.Time(*credential.ExpiresAt)
		}
		w.WriteRow(credential.Name, strings.Join(credential.EnvVars, ","), time)
	}

	return w.Err()
}

func (l *Credentials) Run(cmd *cobra.Command, args []string) error {
	creds, err := l.root.Client.ListCredentials(cmd.Context(), client.ListCredentialsOptions{
		ThreadID: l.ThreadID,
	})
	if err != nil {
		return err
	}

	if l.ThreadID == "" {
		agents, err := l.root.Client.ListAgents(cmd.Context(), client.ListAgentsOptions{})
		if err != nil {
			return err
		}

		for _, agent := range agents.Items {
			agentCreds, err := l.root.Client.ListCredentials(cmd.Context(), client.ListCredentialsOptions{
				AgentID: agent.ID,
			})
			if err != nil {
				return err
			}
			for _, cred := range agentCreds.Items {
				cred.Name = fmt.Sprintf("%s::%s", agent.ID, cred.Name)
				creds.Items = append(creds.Items, cred)
			}
		}

		wfs, err := l.root.Client.ListWorkflows(cmd.Context(), client.ListWorkflowsOptions{})
		if err != nil {
			return err
		}

		for _, wf := range wfs.Items {
			wfCreds, err := l.root.Client.ListCredentials(cmd.Context(), client.ListCredentialsOptions{
				WorkflowID: wf.ID,
			})
			if err != nil {
				return err
			}
			for _, cred := range wfCreds.Items {
				cred.Name = fmt.Sprintf("%s::%s", wf.ID, cred.Name)
				creds.Items = append(creds.Items, cred)
			}
		}
	}

	if l.Quiet {
		return l.printCredentialsQuiet(creds)
	}

	return l.printCredentials(creds)
}
