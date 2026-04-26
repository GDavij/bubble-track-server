package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/bubbletrack/server/internal/config"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
	"github.com/bubbletrack/server/internal/logger"
	"github.com/bubbletrack/server/internal/tui"
	"github.com/bubbletrack/server/internal/tui/websocket"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()
	log := logger.New(cfg.Server.Env)
	slog.SetDefault(log)

	pool, err := repository.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := repository.RunMigrations(ctx, pool); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	graphRepo := repository.NewPostgresGraphRepository(pool)
	graphAdapter := tui.NewGraphAdapter(graphRepo)
	chatClient := tui.NewChatAPIClient(cfg.Server.APIURL)

	var wsClient *websocket.Client
	if cfg.Server.WsURL != "" {
		wsClient = websocket.NewClient(cfg.Server.WsURL, 3, true)
	}

	model := tui.NewModelWithChatService(graphAdapter, chatClient, cfg.Server.DefaultUserID, log, wsClient)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error("failed to start TUI", "error", err)
		os.Exit(1)
	}
}
