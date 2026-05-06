package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"biai/internal/config"
	"biai/internal/llm"
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
		return a.runLLM(ctx, runID, workspace, prompt)
	}
}

func (a *Agent) runLLM(ctx context.Context, runID, workspace, prompt string) ChatResponse {
	cfg, cfgErr := config.LoadUserConfig(a.dataDir)
	sec, secErr := config.LoadUserSecrets(a.dataDir)
	if cfgErr != nil {
		return ChatResponse{RunID: runID, Message: "Khong doc duoc config: " + cfgErr.Error()}
	}
	if secErr != nil {
		return ChatResponse{RunID: runID, Message: "Khong doc duoc token: " + secErr.Error()}
	}
	llmCfg := llm.ChatConfig{BaseURL: cfg.LLMBaseURL, Token: sec.APIToken, Model: cfg.Model}
	events := make([]tools.Event, 0, 4)
	messages := []llm.Message{
		{Role: "system", Content: "You are BiAI AgentDesk, a practical local coding assistant. Use tools when you need workspace evidence. Never invent file contents. For shell commands, explain why; destructive commands require user approval by the app."},
		{Role: "user", Content: prompt},
	}
	for turn := 0; turn < 4; turn++ {
		msg, err := llm.Complete(ctx, llmCfg, messages, toolDefinitions())
		if err != nil {
			return ChatResponse{
				RunID:   runID,
				Message: "Chua goi duoc AI: " + err.Error() + "\n\nKiem tra API URL, Token, Model; sau do bam Tai model va Luu.",
				Events:  events,
			}
		}
		if len(msg.ToolCalls) == 0 {
			if msg.Content == "" {
				msg.Content = "Model khong tra ve noi dung."
			}
			return ChatResponse{RunID: runID, Message: msg.Content, Events: events}
		}
		messages = append(messages, msg)
		for _, call := range msg.ToolCalls {
			ev, content, approval := a.executeLLMTool(runID, workspace, call)
			events = append(events, ev)
			if approval != nil {
				return ChatResponse{
					RunID:    runID,
					Message:  "Command can user approve truoc khi execute.",
					Events:   events,
					Approval: approval,
				}
			}
			messages = append(messages, llm.Message{
				Role:       "tool",
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Content:    content,
			})
		}
	}
	return ChatResponse{RunID: runID, Message: "Agent dung lai sau nhieu tool calls. Hay thu lai voi yeu cau cu the hon.", Events: events}
}

func (a *Agent) executeLLMTool(runID, workspace string, call llm.ToolCall) (tools.Event, string, *ApprovalDraft) {
	var args map[string]string
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		ev := tools.Event{Name: "tool." + call.Function.Name, OK: false, Message: "invalid tool arguments: " + err.Error()}
		return ev, ev.Message, nil
	}
	switch call.Function.Name {
	case "list_directory":
		ev, out := a.tools.ListDirectory(workspace, args["path"])
		return ev, out, nil
	case "read_file":
		ev, out := a.tools.ReadFile(workspace, args["path"])
		return ev, out, nil
	case "search_text":
		ev, out := a.tools.SearchText(workspace, args["query"])
		return ev, out, nil
	case "run_command":
		reason := args["reason"]
		if reason == "" {
			reason = "Model requested command execution."
		}
		ev, approval, msg := a.tools.PlanCommand(runID, workspace, args["command"], reason)
		if approval != nil {
			return ev, msg, &ApprovalDraft{
				ID:            approval.ID,
				RunID:         runID,
				ToolName:      approval.ToolName,
				Risk:          string(approval.Risk),
				Command:       approval.Command,
				Cwd:           approval.Cwd,
				Reason:        approval.ReasonFromAgent,
				SafetySummary: approval.SafetySummary,
				Arguments:     approval.Arguments,
			}
		}
		return ev, msg, nil
	default:
		ev := tools.Event{Name: "tool." + call.Function.Name, OK: false, Message: "unknown tool"}
		return ev, ev.Message, nil
	}
}

func toolDefinitions() []llm.ToolDefinition {
	return []llm.ToolDefinition{
		toolDef("list_directory", "List files and folders under the workspace.", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "Workspace-relative path, default ."},
			},
		}),
		toolDef("read_file", "Read a text file inside the workspace.", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "Workspace-relative file path"},
			},
			"required": []string{"path"},
		}),
		toolDef("search_text", "Search text in workspace files.", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{"type": "string", "description": "Text to search for"},
			},
			"required": []string{"query"},
		}),
		toolDef("run_command", "Run a command in the workspace. Destructive commands are approval-gated.", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{"type": "string", "description": "Command to run"},
				"reason":  map[string]interface{}{"type": "string", "description": "Why this command is needed"},
			},
			"required": []string{"command", "reason"},
		}),
	}
}

func toolDef(name, desc string, params map[string]interface{}) llm.ToolDefinition {
	var d llm.ToolDefinition
	d.Type = "function"
	d.Function.Name = name
	d.Function.Description = desc
	d.Function.Parameters = params
	return d
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
