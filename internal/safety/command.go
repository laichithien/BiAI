package safety

import (
	"path/filepath"
	"regexp"
	"strings"
)

type CommandRisk string

const (
	CommandReadOnly       CommandRisk = "read_only"
	CommandVerify         CommandRisk = "verify"
	CommandBuildWrite     CommandRisk = "write_build_artifacts"
	CommandPackageInstall CommandRisk = "package_install"
	CommandNetwork        CommandRisk = "network"
	CommandDestructive    CommandRisk = "destructive"
	CommandScriptUnknown  CommandRisk = "script_unknown"
	CommandPrivileged     CommandRisk = "privileged"
)

type CommandAssessment struct {
	Command          string      `json:"command"`
	Cwd              string      `json:"cwd"`
	Risk             CommandRisk `json:"risk"`
	RequiresApproval bool        `json:"requires_approval"`
	HardDeny         bool        `json:"hard_deny"`
	DeleteLike       bool        `json:"delete_like"`
	Reason           string      `json:"reason"`
	Markers          []string    `json:"markers"`
	TargetHints      []string    `json:"target_hints"`
}

var spaceRE = regexp.MustCompile(`\s+`)

func AssessCommand(workspace, cwd, command string) CommandAssessment {
	cmd := strings.TrimSpace(command)
	assessment := CommandAssessment{
		Command: cmd,
		Cwd:     cwd,
		Risk:    CommandScriptUnknown,
		Reason:  "unknown command shape",
	}
	if cmd == "" {
		assessment.HardDeny = true
		assessment.Reason = "empty command"
		return assessment
	}
	if _, err := ResolveTarget(workspace, cwd); err != nil {
		assessment.HardDeny = true
		assessment.Reason = "command cwd is outside workspace"
		return assessment
	}

	lower := strings.ToLower(cmd)
	markers := findCommandMarkers(lower)
	assessment.Markers = markers
	assessment.TargetHints = extractTargetHints(cmd)

	if containsAny(lower, []string{"format ", "diskpart", "cipher /w", "shutdown", "reg delete"}) {
		assessment.Risk = CommandPrivileged
		assessment.HardDeny = true
		assessment.Reason = "privileged or system destructive command"
		return assessment
	}
	if targetsSystemPath(lower) || hasDangerousWideDelete(lower) {
		assessment.Risk = CommandDestructive
		assessment.DeleteLike = true
		assessment.HardDeny = true
		assessment.Reason = "delete command targets a protected or too broad path"
		return assessment
	}
	if len(markers) > 0 {
		assessment.Risk = CommandDestructive
		assessment.DeleteLike = true
		assessment.RequiresApproval = true
		assessment.Reason = "command contains delete/destructive marker"
		return assessment
	}
	if containsShellMeta(cmd) {
		assessment.Risk = CommandScriptUnknown
		assessment.RequiresApproval = true
		assessment.Reason = "command uses shell metacharacters"
		return assessment
	}

	fields := spaceRE.Split(cmd, -1)
	exe := strings.ToLower(filepath.Base(fields[0]))
	exe = strings.TrimSuffix(exe, ".exe")
	switch exe {
	case "dir", "type":
		assessment.Risk = CommandReadOnly
		assessment.Reason = "read-only Windows command"
	case "git":
		if len(fields) > 1 && (fields[1] == "status" || fields[1] == "diff" || fields[1] == "log") {
			assessment.Risk = CommandReadOnly
			assessment.Reason = "read-only git command"
		} else {
			assessment.Risk = CommandScriptUnknown
			assessment.RequiresApproval = true
			assessment.Reason = "git command may change repository state"
		}
	case "go":
		if len(fields) > 1 && fields[1] == "test" {
			assessment.Risk = CommandVerify
			assessment.RequiresApproval = true
			assessment.Reason = "test command"
		} else if len(fields) > 1 && fields[1] == "env" {
			assessment.Risk = CommandReadOnly
			assessment.Reason = "read-only go env"
		} else {
			assessment.Risk = CommandBuildWrite
			assessment.RequiresApproval = true
			assessment.Reason = "go command may write build artifacts"
		}
	case "npm", "pip":
		if containsAny(lower, []string{" install", " add", " update"}) {
			assessment.Risk = CommandPackageInstall
			assessment.RequiresApproval = true
			assessment.Reason = "package manager command may use network and change dependencies"
		} else {
			assessment.Risk = CommandScriptUnknown
			assessment.RequiresApproval = true
			assessment.Reason = "package manager script may have side effects"
		}
	case "curl", "wget":
		assessment.Risk = CommandNetwork
		assessment.RequiresApproval = true
		assessment.Reason = "network command"
	case "powershell", "cmd", "wscript", "cscript":
		assessment.Risk = CommandScriptUnknown
		assessment.RequiresApproval = true
		assessment.Reason = "shell/script host command"
	default:
		assessment.Risk = CommandScriptUnknown
		assessment.RequiresApproval = true
		assessment.Reason = "not on read-only allowlist"
	}
	return assessment
}

func findCommandMarkers(lower string) []string {
	patterns := []string{
		"del ", "erase ", "rd ", "rmdir ", "remove-item", "rm ", "rimraf",
		"git clean", "git reset --hard", "git checkout -- .",
		"cmd /c del", "cmd /c rd",
	}
	var found []string
	for _, p := range patterns {
		if strings.Contains(lower, p) || strings.HasPrefix(lower, strings.TrimSpace(p)+" ") {
			found = append(found, strings.TrimSpace(p))
		}
	}
	return found
}

func extractTargetHints(command string) []string {
	fields := spaceRE.Split(strings.TrimSpace(command), -1)
	var out []string
	for i, f := range fields {
		lf := strings.ToLower(f)
		if lf == "del" || lf == "erase" || lf == "rd" || lf == "rmdir" || lf == "rm" || lf == "remove-item" {
			for _, candidate := range fields[i+1:] {
				if strings.HasPrefix(candidate, "-") || strings.HasPrefix(candidate, "/") {
					continue
				}
				out = append(out, strings.Trim(candidate, `"'`))
				break
			}
		}
	}
	return out
}

func containsAny(s string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

func containsShellMeta(s string) bool {
	return strings.ContainsAny(s, "&|<>%!`")
}

func targetsSystemPath(lower string) bool {
	return containsAny(lower, []string{
		"c:\\windows", "c:/windows", "c:\\program files", "c:/program files",
	})
}

func hasDangerousWideDelete(lower string) bool {
	return containsAny(lower, []string{
		"del *", "erase *", "rd /s c:\\", "rmdir /s c:\\", "remove-item -recurse c:\\", "rm -rf /",
	})
}
