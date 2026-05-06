package safety

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrOutsideWorkspace = errors.New("path is outside workspace")

func NormalizeWorkspace(workspace string) (string, error) {
	if strings.TrimSpace(workspace) == "" {
		return "", errors.New("workspace is required")
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func ResolveTarget(workspace, target string) (string, error) {
	ws, err := NormalizeWorkspace(workspace)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(target) == "" {
		target = "."
	}
	var p string
	if filepath.IsAbs(target) {
		p = target
	} else {
		p = filepath.Join(ws, target)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	if !IsWithinWorkspace(ws, abs) {
		return "", ErrOutsideWorkspace
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		real = filepath.Clean(real)
		if !IsWithinWorkspace(ws, real) {
			return "", ErrOutsideWorkspace
		}
		return real, nil
	}
	return abs, nil
}

func IsWithinWorkspace(workspace, target string) bool {
	ws := filepath.Clean(workspace)
	t := filepath.Clean(target)
	if strings.EqualFold(ws, t) {
		return true
	}
	rel, err := filepath.Rel(ws, t)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func IsProtectedPath(path string) bool {
	clean := strings.ToLower(filepath.ToSlash(filepath.Clean(path)))
	base := strings.ToLower(filepath.Base(clean))
	if base == ".git" || strings.Contains(clean, "/.git/") {
		return true
	}
	if base == ".svn" || base == ".hg" || strings.Contains(clean, "/.svn/") || strings.Contains(clean, "/.hg/") {
		return true
	}
	if base == ".env" || strings.HasPrefix(base, ".env.") {
		return true
	}
	switch base {
	case "package-lock.json", "pnpm-lock.yaml", "yarn.lock", "go.sum":
		return true
	}
	for _, suffix := range []string{".key", ".pem", ".pfx", ".cert"} {
		if strings.HasSuffix(base, suffix) {
			return true
		}
	}
	if strings.HasPrefix(clean, "c:/windows") || strings.HasPrefix(clean, "c:/program files") {
		return true
	}
	return false
}
