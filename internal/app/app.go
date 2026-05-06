package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"biai/internal/agent"
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

	a := agent.New(agent.Config{
		DataDir: dataDir,
	})

	server, err := ui.NewServer(a, dataDir)
	if err != nil {
		return err
	}
	if err := server.Start(); err != nil {
		return err
	}
	defer server.Close()

	if err := ui.OpenWindow(server.URL(), dataDir); err != nil {
		fmt.Printf("Open this URL in a browser: %s\n", server.URL())
		fmt.Printf("Window launch failed: %v\n", err)
	} else {
		fmt.Printf("AgentDesk running at %s\n", server.URL())
	}

	<-ctx.Done()
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
