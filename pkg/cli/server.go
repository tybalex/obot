package cli

import (
	"github.com/gptscript-ai/otto/pkg/server"
	"github.com/gptscript-ai/otto/pkg/services"
	"github.com/spf13/cobra"
)

type Server struct {
	services.Config
}

func (s *Server) Run(cmd *cobra.Command, args []string) error {
	return server.Run(cmd.Context(), s.Config)
}
