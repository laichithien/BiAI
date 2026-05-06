package memory

import (
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
