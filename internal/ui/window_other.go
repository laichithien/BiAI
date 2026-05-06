//go:build !windows

package ui

import (
	"os/exec"
	"runtime"
)

func OpenWindow(url, dataDir string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
