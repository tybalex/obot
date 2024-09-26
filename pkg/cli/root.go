package cli

import (
	"os"

	"github.com/fatih/color"
	"github.com/gptscript-ai/cmd"
	"github.com/gptscript-ai/otto/apiclient"
	log2 "github.com/gptscript-ai/otto/logger"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var log = log2.Package()

type Otto struct {
	Debug  bool `usage:"Enable debug logging"`
	Client *apiclient.Client
}

func (a *Otto) PersistentPre(cmd *cobra.Command, args []string) error {
	if os.Getenv("NO_COLOR") != "" || !term.IsTerminal(int(os.Stdout.Fd())) {
		color.NoColor = true
	}

	if a.Debug {
		log2.SetDebug()
	}
	return nil
}

func New() *cobra.Command {
	root := &Otto{
		Client: &apiclient.Client{
			BaseURL: "http://localhost:8080",
			Token:   os.Getenv("OTTO_TOKEN"),
		},
	}
	return cmd.Command(root,
		&Create{root: root},
		&Agents{root: root},
		&Workflows{root: root},
		&Edit{root: root},
		&Update{root: root},
		&Delete{root: root},
		&Invoke{root: root},
		cmd.Command(&Threads{root: root}, &ThreadPrint{root: root}),
		cmd.Command(&Credentials{root: root}, &CredentialsDelete{root: root}),
		cmd.Command(&Runs{root: root}, &Debug{root: root}, &RunPrint{root: root}),
		cmd.Command(&StepTemplates{root: root},
			&StepTemplatesDelete{root: root},
			&StepTemplateCreate{root: root},
			&StepTemplateUpdate{root: root}),
		&Server{})
}

func (a *Otto) Run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
