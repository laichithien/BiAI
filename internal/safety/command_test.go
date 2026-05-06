package safety

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAssessCommandDeleteRequiresApproval(t *testing.T) {
	ws := t.TempDir()
	cases := []string{
		"del tmp.txt",
		"rd /s build",
		"powershell Remove-Item tmp.txt",
		"git clean -fd",
		"git reset --hard",
		"cmd /C del tmp.txt",
		"echo ok && del tmp.txt",
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			got := AssessCommand(ws, ws, tc)
			if !got.DeleteLike {
				t.Fatalf("expected delete-like command")
			}
			if got.HardDeny {
				t.Fatalf("expected approval gate, got hard deny: %s", got.Reason)
			}
			if !got.RequiresApproval {
				t.Fatalf("expected approval required")
			}
		})
	}
}

func TestAssessCommandHardDenySystemCommands(t *testing.T) {
	ws := t.TempDir()
	cases := []string{
		"format C:",
		"diskpart",
		"reg delete HKCU\\Software\\Test",
		"del C:\\Windows\\win.ini",
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			got := AssessCommand(ws, ws, tc)
			if !got.HardDeny {
				t.Fatalf("expected hard deny, got %#v", got)
			}
		})
	}
}

func TestAssessCommandReadOnly(t *testing.T) {
	ws := t.TempDir()
	got := AssessCommand(ws, ws, "git status")
	if got.Risk != CommandReadOnly || got.RequiresApproval {
		t.Fatalf("expected read-only without approval, got %#v", got)
	}
}

func TestAssessCommandOutsideCwdDenied(t *testing.T) {
	ws := t.TempDir()
	outside := filepath.Dir(ws)
	got := AssessCommand(ws, outside, "git status")
	if !got.HardDeny {
		t.Fatalf("expected hard deny for cwd outside workspace")
	}
}

func TestResolveTargetBlocksTraversal(t *testing.T) {
	ws := t.TempDir()
	outsideFile := filepath.Join(filepath.Dir(ws), "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveTarget(ws, filepath.Join("..", filepath.Base(outsideFile)))
	if err != ErrOutsideWorkspace {
		t.Fatalf("expected ErrOutsideWorkspace, got %v", err)
	}
}
