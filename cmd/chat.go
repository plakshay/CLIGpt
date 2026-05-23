package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/lakshaysinghal/cligpt/internal/ai"
	"github.com/lakshaysinghal/cligpt/internal/chat"
	"github.com/lakshaysinghal/cligpt/internal/config"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		client, err := ai.New(ctx, cfg.GeminiAPIKey, cfg.Model)
		if err != nil {
			return err
		}
		session := chat.New(client, cfg.Timeout)

		if err := session.Run(ctx); err != nil {
			return fmt.Errorf("chat session: %w", err)
		}
		return nil
	},
}