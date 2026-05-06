package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureWorkspaceAgentsCreatesTemplateOnce(t *testing.T) {
	workspace := t.TempDir()
	path, created, err := EnsureWorkspaceAgents(workspace)
	if err != nil {
		t.Fatalf("EnsureWorkspaceAgents failed: %v", err)
	}
	if !created {
		t.Fatalf("expected AGENTS.md to be created")
	}
	if filepath.Base(path) != "AGENTS.md" {
		t.Fatalf("unexpected path: %s", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(b), "BiAI Agent Instructions") {
		t.Fatalf("template missing title: %s", string(b))
	}

	path2, created2, err := EnsureWorkspaceAgents(workspace)
	if err != nil {
		t.Fatalf("second EnsureWorkspaceAgents failed: %v", err)
	}
	if created2 {
		t.Fatalf("existing AGENTS.md should not be recreated")
	}
	if path2 != path {
		t.Fatalf("expected same path, got %s and %s", path, path2)
	}
}

func TestLoadInstructionsReadsGlobalAndWorkspaceFiles(t *testing.T) {
	dataDir := t.TempDir()
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "AGENTS.md"), []byte("global rules"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "GEMINI.md"), []byte("workspace rules"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := LoadInstructions(dataDir, workspace)
	if len(got.Loaded) != 2 {
		t.Fatalf("expected 2 instruction files, got %d: %#v", len(got.Loaded), got.Loaded)
	}
	if !strings.Contains(got.Text, "global rules") || !strings.Contains(got.Text, "workspace rules") {
		t.Fatalf("missing loaded instruction text: %s", got.Text)
	}
}
