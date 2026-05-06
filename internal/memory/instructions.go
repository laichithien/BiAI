package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"biai/internal/safety"
)

type InstructionSet struct {
	Loaded []InstructionFile `json:"loaded"`
	Text   string            `json:"text"`
}

type InstructionFile struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func LoadInstructions(dataDir, workspace string) InstructionSet {
	var out InstructionSet
	names := []string{"AGENTS.md", "CLAUDE.md", "GEMINI.md"}
	for _, name := range names {
		loadInstructionFile(&out, filepath.Join(dataDir, name), name)
	}
	ws, err := safety.NormalizeWorkspace(workspace)
	if err == nil {
		for _, name := range names {
			loadInstructionFile(&out, filepath.Join(ws, name), name)
		}
	}
	return out
}

func EnsureWorkspaceAgents(workspace string) (string, bool, error) {
	ws, err := safety.NormalizeWorkspace(workspace)
	if err != nil {
		return "", false, err
	}
	path := filepath.Join(ws, "AGENTS.md")
	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	} else if !os.IsNotExist(err) {
		return "", false, err
	}
	template := fmt.Sprintf(`# BiAI Agent Instructions

These instructions are loaded automatically for this workspace.

## Project
- Workspace: %s
- Describe the app, main commands, and coding conventions here.

## Working Rules
- Prefer small, verifiable changes.
- Read existing code before editing.
- Run relevant tests or explain why they could not run.
- Ask for approval before destructive commands or broad file changes.

## Commands
- Test:
- Build:
- Lint:
`, ws)
	if err := os.WriteFile(path, []byte(template), 0o600); err != nil {
		return "", false, err
	}
	return path, true, nil
}

func loadInstructionFile(out *InstructionSet, path, name string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	text := strings.TrimSpace(string(b))
	if text == "" {
		return
	}
	out.Loaded = append(out.Loaded, InstructionFile{Path: path, Name: name})
	if out.Text != "" {
		out.Text += "\n\n"
	}
	out.Text += "## " + path + "\n" + text
}
