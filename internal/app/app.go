package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"biai/internal/agent"
	"biai/internal/logging"
	"biai/internal/ui"
)

func Run(ctx context.Context) error {
	dataDir, err := AppDataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	logger := logging.New(dataDir)
	defer logger.Recover("app.Run")
	logger.Printf("starting AgentDesk dataDir=%s", dataDir)

	a := agent.New(agent.Config{
		DataDir: dataDir,
	})

	server, err := ui.NewServer(a, dataDir, logger)
	if err != nil {
		logger.Printf("create UI server failed: %v", err)
		return err
	}
	if err := server.Start(); err != nil {
		logger.Printf("start UI server failed: %v", err)
		return err
	}
	defer server.Close()
	logger.Printf("UI server listening url=%s", server.URL())

	if err := ui.OpenWindow(server.URL(), dataDir); err != nil {
		logger.Printf("window launch failed: %v", err)
		fmt.Printf("Open this URL in a browser: %s\n", server.URL())
		fmt.Printf("Window launch failed: %v\n", err)
	} else {
		logger.Printf("window launch requested")
		fmt.Printf("AgentDesk running at %s\n", server.URL())
	}

	<-ctx.Done()
	logger.Printf("shutdown requested")
	return nil
}

func AppDataDir() (string, error) {
	if dir := os.Getenv("APPDATA"); dir != "" {
		return filepath.Join(dir, "BiAI", "AgentDesk"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".biai", "agentdesk"), nil
}
