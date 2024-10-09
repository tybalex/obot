package cli

import (
	"github.com/otto8-ai/otto8/pkg/server"
	"github.com/otto8-ai/otto8/pkg/services"
	"github.com/spf13/cobra"
)

type Server struct {
	services.Config
}

func (s *Server) Run(cmd *cobra.Command, args []string) error {
	return server.Run(cmd.Context(), s.Config)
}
