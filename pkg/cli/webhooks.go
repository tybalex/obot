package cli

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type Webhooks struct {
	root  *Otto8
	Quiet bool `usage:"Only print IDs of agents" short:"q"`
	Wide  bool `usage:"Print more information" short:"w"`
}

func (l *Webhooks) Customize(cmd *cobra.Command) {
	cmd.Aliases = []string{"webhook", "wh"}
}

func (l *Webhooks) Run(cmd *cobra.Command, args []string) error {
	whs, err := l.root.Client.ListWebhooks(cmd.Context())
	if err != nil {
		return err
	}

	if l.Quiet {
		for _, webhook := range whs.Items {
			fmt.Println(webhook.ID)
		}
		return nil
	}

	w := newTable("ID", "DESCRIPTION", "WORKFLOW", "CREATED")
	for _, wh := range whs.Items {
		w.WriteRow(wh.ID, truncate(wh.Description, l.Wide), wh.WorkflowID, humanize.Time(wh.Created.Time))
	}

	return w.Err()
}
