package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"biai/internal/safety"
)

type Event struct {
	Name    string                 `json:"name"`
	OK      bool                   `json:"ok"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type Approval struct {
	ID              string
	RunID           string
	ToolName        string
	Risk            safety.CommandRisk
	Command         string
	Cwd             string
	ReasonFromAgent string
	SafetySummary   string
	Arguments       map[string]string
	Assessment      safety.CommandAssessment
}

type Registry struct {
	audit     *safety.AuditLog
	mu        sync.Mutex
	approvals map[string]Approval
}

func NewRegistry(audit *safety.AuditLog) *Registry {
	return &Registry{
		audit:     audit,
		approvals: make(map[string]Approval),
	}
}

func (r *Registry) ListDirectory(workspace, target string) (Event, string) {
	path, err := safety.ResolveTarget(workspace, target)
	if err != nil {
		return fail("tool.list_directory", err), "Khong the list directory: " + err.Error()
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fail("tool.list_directory", err), "Khong the doc folder: " + err.Error()
	}
	var b strings.Builder
	count := 0
	for _, e := range entries {
		if count >= 200 {
			b.WriteString("... truncated\n")
			break
		}
		name := e.Name()
		if e.IsDir() {
			name += string(os.PathSeparator)
		}
		b.WriteString(name)
		b.WriteByte('\n')
		count++
	}
	r.audit.Write(map[string]interface{}{
		"tool": "list_directory", "risk": "read_only", "target": path, "ok": true,
	})
	return Event{Name: "tool.list_directory", OK: true, Message: "Listed directory", Data: map[string]interface{}{"path": path}}, b.String()
}

func (r *Registry) ReadFile(workspace, target string) (Event, string) {
	path, err := safety.ResolveTarget(workspace, target)
	if err != nil {
		return fail("tool.read_file", err), "Khong the read file: " + err.Error()
	}
	if safety.IsProtectedPath(path) {
		err := errors.New("protected file is blocked")
		return fail("tool.read_file", err), err.Error()
	}
	f, err := os.Open(path)
	if err != nil {
		return fail("tool.read_file", err), "Khong the mo file: " + err.Error()
	}
	defer f.Close()
	var buf bytes.Buffer
	_, err = io.CopyN(&buf, f, 256*1024+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return fail("tool.read_file", err), "Khong the doc file: " + err.Error()
	}
	out := buf.String()
	if len(out) > 256*1024 {
		out = out[:256*1024] + "\n... truncated"
	}
	r.audit.Write(map[string]interface{}{
		"tool": "read_file", "risk": "read_only", "target": path, "ok": true,
	})
	return Event{Name: "tool.read_file", OK: true, Message: "Read file", Data: map[string]interface{}{"path": path}}, out
}

func (r *Registry) SearchText(workspace, query string) (Event, string) {
	query = strings.TrimSpace(query)
	if query == "" {
		err := errors.New("search query is required")
		return fail("tool.search_text", err), err.Error()
	}
	ws, err := safety.NormalizeWorkspace(workspace)
	if err != nil {
		return fail("tool.search_text", err), err.Error()
	}
	var b strings.Builder
	matches := 0
	err = filepath.WalkDir(ws, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || matches >= 100 {
			return nil
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if name == ".git" || name == "node_modules" || name == ".agent_trash" {
				return filepath.SkipDir
			}
			return nil
		}
		if safety.IsProtectedPath(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || bytes.IndexByte(data, 0) >= 0 {
			return nil
		}
		if strings.Contains(strings.ToLower(string(data)), strings.ToLower(query)) {
			rel, _ := filepath.Rel(ws, path)
			b.WriteString(rel)
			b.WriteByte('\n')
			matches++
		}
		return nil
	})
	if err != nil {
		return fail("tool.search_text", err), err.Error()
	}
	r.audit.Write(map[string]interface{}{
		"tool": "search_text", "risk": "read_only", "query": query, "matches": matches, "ok": true,
	})
	if matches == 0 {
		return Event{Name: "tool.search_text", OK: true, Message: "No matches"}, "Khong tim thay ket qua."
	}
	return Event{Name: "tool.search_text", OK: true, Message: "Search completed"}, b.String()
}

func (r *Registry) PlanCommand(runID, workspace, command, reason string) (Event, *Approval, string) {
	ws, err := safety.NormalizeWorkspace(workspace)
	if err != nil {
		return fail("tool.run_command", err), nil, err.Error()
	}
	assessment := safety.AssessCommand(ws, ws, command)
	if assessment.HardDeny {
		r.audit.Write(map[string]interface{}{
			"run_id": runID, "tool": "run_command", "command": command, "risk": assessment.Risk,
			"approval": "blocked", "ok": false, "reason": assessment.Reason,
		})
		msg := "Command bi chan truoc khi chay: " + assessment.Reason
		return Event{Name: "tool.run_command.blocked", OK: false, Message: msg, Data: map[string]interface{}{"assessment": assessment}}, nil, msg
	}
	if assessment.DeleteLike || assessment.RequiresApproval {
		ap := Approval{
			ID:              fmt.Sprintf("appr_%d", time.Now().UnixNano()),
			RunID:           runID,
			ToolName:        "run_command",
			Risk:            assessment.Risk,
			Command:         command,
			Cwd:             ws,
			ReasonFromAgent: reason,
			SafetySummary:   commandSafetySummary(assessment),
			Arguments: map[string]string{
				"command": command,
				"cwd":     ws,
			},
			Assessment: assessment,
		}
		r.mu.Lock()
		r.approvals[ap.ID] = ap
		r.mu.Unlock()
		r.audit.Write(map[string]interface{}{
			"run_id": runID, "tool": "run_command", "command": command, "risk": assessment.Risk,
			"approval": "requested", "ok": false, "reason": assessment.Reason,
		})
		return Event{Name: "approval.required", OK: true, Message: "Command can approval", Data: map[string]interface{}{"approval_id": ap.ID, "assessment": assessment}}, &ap, "Command can user approve truoc khi execute."
	}
	ev, msg := r.executeCommand(runID, ws, command, "auto")
	return ev, nil, msg
}

func (r *Registry) DecideCommandApproval(approvalID, decision string) (Event, string, bool) {
	r.mu.Lock()
	ap, ok := r.approvals[approvalID]
	if ok {
		delete(r.approvals, approvalID)
	}
	r.mu.Unlock()
	if !ok {
		return Event{}, "Approval khong ton tai hoac da het han.", false
	}
	if strings.ToLower(decision) != "allow" {
		r.audit.Write(map[string]interface{}{
			"run_id": ap.RunID, "tool": "run_command", "command": ap.Command, "risk": ap.Risk,
			"approval": "deny", "ok": false,
		})
		return Event{Name: "tool.run_command.denied", OK: false, Message: "User denied command"}, "Da tu choi command. Khong co lenh nao duoc chay.", true
	}
	ev, msg := r.executeCommand(ap.RunID, ap.Cwd, ap.Command, "allow")
	return ev, msg, ev.OK
}

func (r *Registry) executeCommand(runID, cwd, command, approval string) (Event, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)
	if _, err := exec.LookPath("cmd"); err != nil {
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return fail("tool.run_command", errors.New("empty command")), "empty command"
		}
		cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	}
	cmd.Dir = cwd
	cmd.Env = minimalEnv()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)
	output := out.String()
	if len(output) > 12000 {
		output = output[:12000] + "\n... truncated"
	}
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	r.audit.Write(map[string]interface{}{
		"run_id": runID, "tool": "run_command", "command": command, "cwd": cwd,
		"approval": approval, "ok": err == nil, "exit_code": exitCode, "duration_ms": duration.Milliseconds(),
	})
	if err != nil {
		return Event{Name: "tool.run_command.completed", OK: false, Message: err.Error(), Data: map[string]interface{}{"output": output}}, output
	}
	return Event{Name: "tool.run_command.completed", OK: true, Message: "Command completed", Data: map[string]interface{}{"output": output}}, output
}

func commandSafetySummary(a safety.CommandAssessment) string {
	if a.DeleteLike {
		return "Command co dau hieu xoa/destructive. Chi execute sau khi user approve."
	}
	return "Command can approval vi co the co side effect: " + a.Reason
}

func minimalEnv() []string {
	keys := []string{"PATH", "SystemRoot", "TEMP", "TMP", "COMSPEC", "APPDATA", "USERPROFILE"}
	var env []string
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			env = append(env, k+"="+v)
		}
	}
	return env
}

func fail(name string, err error) Event {
	return Event{Name: name, OK: false, Message: err.Error()}
}
