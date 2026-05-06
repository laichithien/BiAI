package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"

	"biai/internal/app"
	"biai/internal/platform"
)

func main() {
	defer func() {
		if v := recover(); v != nil {
			writeCrashLog(v)
			log.Fatalf("panic: %v", v)
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.Run(ctx); err != nil {
		writeFatalLog(err)
		platform.ShowError("BiAI AgentDesk failed to start", err.Error())
		log.Fatal(err)
	}
}

func writeFatalLog(err error) {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir, _ = os.UserHomeDir()
	}
	if dir == "" {
		return
	}
	path := filepath.Join(dir, "BiAI", "AgentDesk")
	_ = os.MkdirAll(path, 0o700)
	_ = os.WriteFile(filepath.Join(path, "startup-error.log"), []byte(fmt.Sprintf("startup error: %v\n", err)), 0o600)
}

func writeCrashLog(v interface{}) {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir, _ = os.UserHomeDir()
	}
	if dir == "" {
		return
	}
	path := filepath.Join(dir, "BiAI", "AgentDesk")
	_ = os.MkdirAll(path, 0o700)
	_ = os.WriteFile(filepath.Join(path, "crash.log"), []byte(fmt.Sprintf("panic: %v\n%s", v, string(debug.Stack()))), 0o600)
}
