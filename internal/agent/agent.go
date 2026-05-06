package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"biai/internal/safety"
	"biai/internal/tools"
)

type Config struct {
	DataDir string
}

type Agent struct {
	dataDir string
	tools   *tools.Registry
	audit   *safety.AuditLog
}

func New(cfg Config) *Agent {
	audit := safety.NewAuditLog(cfg.DataDir)
	return &Agent{
		dataDir: cfg.DataDir,
		audit:   audit,
		tools:   tools.NewRegistry(audit),
	}
}

type ChatRequest struct {
	Prompt    string `json:"prompt"`
	Workspace string `json:"workspace"`
}

type ChatResponse struct {
	RunID    string         `json:"run_id"`
	Message  string         `json:"message"`
	Events   []tools.Event  `json:"events"`
	Approval *ApprovalDraft `json:"approval,omitempty"`
}

type ApprovalDraft struct {
	ID            string            `json:"id"`
	RunID         string            `json:"run_id"`
	ToolName      string            `json:"tool_name"`
	Risk          string            `json:"risk"`
	Command       string            `json:"command,omitempty"`
	Cwd           string            `json:"cwd,omitempty"`
	Reason        string            `json:"reason"`
	SafetySummary string            `json:"safety_summary"`
	Arguments     map[string]string `json:"arguments"`
}

func (a *Agent) Chat(ctx context.Context, req ChatRequest) ChatResponse {
	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	workspace := strings.TrimSpace(req.Workspace)
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return ChatResponse{RunID: runID, Message: "Nhap noi dung truoc khi gui."}
	}

	events := make([]tools.Event, 0, 2)
	lower := strings.ToLower(prompt)

	switch {
	case strings.HasPrefix(lower, "/list"):
		ev, msg := a.tools.ListDirectory(workspace, strings.TrimSpace(prompt[len("/list"):]))
		events = append(events, ev)
		return ChatResponse{RunID: runID, Message: msg, Events: events}
	case strings.HasPrefix(lower, "/read"):
		ev, msg := a.tools.ReadFile(workspace, strings.TrimSpace(prompt[len("/read"):]))
		events = append(events, ev)
		return ChatResponse{RunID: runID, Message: msg, Events: events}
	case strings.HasPrefix(lower, "/search"):
		ev, msg := a.tools.SearchText(workspace, strings.TrimSpace(prompt[len("/search"):]))
		events = append(events, ev)
		return ChatResponse{RunID: runID, Message: msg, Events: events}
	case strings.HasPrefix(lower, "/cmd"):
		command := strings.TrimSpace(prompt[len("/cmd"):])
		ev, approval, msg := a.tools.PlanCommand(runID, workspace, command, "User requested command execution from chat.")
		events = append(events, ev)
		if approval != nil {
			return ChatResponse{
				RunID:   runID,
				Message: msg,
				Events:  events,
				Approval: &ApprovalDraft{
					ID:            approval.ID,
					RunID:         runID,
					ToolName:      approval.ToolName,
					Risk:          string(approval.Risk),
					Command:       approval.Command,
					Cwd:           approval.Cwd,
					Reason:        approval.ReasonFromAgent,
					SafetySummary: approval.SafetySummary,
					Arguments:     approval.Arguments,
				},
			}
		}
		return ChatResponse{RunID: runID, Message: msg, Events: events}
	default:
		return ChatResponse{
			RunID: runID,
			Message: "MVP hien co cac lenh nhanh: /list [path], /read <file>, /search <text>, /cmd <command>. " +
				"Command nguy hiem se bi loc va hoi approve truoc khi chay.",
		}
	}
}

type ApprovalRequest struct {
	ApprovalID string `json:"approval_id"`
	Decision   string `json:"decision"`
}

type ApprovalResponse struct {
	OK      bool         `json:"ok"`
	Message string       `json:"message"`
	Event   *tools.Event `json:"event,omitempty"`
}

func (a *Agent) DecideApproval(ctx context.Context, req ApprovalRequest) ApprovalResponse {
	ev, msg, ok := a.tools.DecideCommandApproval(req.ApprovalID, req.Decision)
	if ev.Name != "" {
		return ApprovalResponse{OK: ok, Message: msg, Event: &ev}
	}
	return ApprovalResponse{OK: ok, Message: msg}
}
