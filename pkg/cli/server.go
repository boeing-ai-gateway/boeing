package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/boeing-ai-gateway/boeing/pkg/server"
	"github.com/boeing-ai-gateway/boeing/pkg/services"
	"github.com/spf13/cobra"
)

type Server struct {
	services.Config
}

func (s *Server) Customize(cmd *cobra.Command) {
	cmd.Hidden = true
}

func (s *Server) Run(cmd *cobra.Command, _ []string) error {
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	return server.Run(ctx, s.Config)
}
