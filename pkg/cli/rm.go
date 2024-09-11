package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

type Delete struct {
	root *Otto
}

func (l *Delete) Customize(cmd *cobra.Command) {
	cmd.Use = "delete [flags] ID..."
	cmd.Aliases = []string{"rm", "del"}
}

func (l *Delete) Run(cmd *cobra.Command, args []string) error {
	var errs []error
	for _, id := range args {
		if len(id) < 1 {
			continue
		}

		switch id[0:2] {
		//case "t1":
		//	if err := l.root.client.DeleteThread(cmd.Context(), id); err != nil {
		//		errs = append(errs, err)
		//	} else {
		//		log.Errorf("Thread deleted: %s\n", id)
		//	}
		default:
			fallthrough
		case "a1":
			if err := l.root.Client.DeleteAgent(cmd.Context(), id); err != nil {
				errs = append(errs, err)
			} else {
				log.Infof("Agent deleted: %s\n", id)
			}
			//case "r1":
			//	if err := l.root.client.DeleteRun(cmd.Context(), id); err != nil {
			//		errs = append(errs, err)
			//	} else {
			//		log.Errorf("Run deleted: %s\n", id)
			//	}
		}
	}

	return errors.Join(errs...)
}
