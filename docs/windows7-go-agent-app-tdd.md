# TDD: Windows 7 Go Desktop AI Agent

Ngay tao: 2026-05-06.

Muc tieu: xay mot ung dung desktop chay tren Windows 7, core viet bang Go, UI chat/quan ly dung HTML/CSS/JS nhe trong cua so desktop rieng. Agent co tool doc/liet ke/tim kiem file va co co che bao ve nghiem ngat khi tool co kha nang xoa, ghi de, di chuyen hoac tac dong nguy hiem.

## 1. Important Compatibility Decision

### 1.1 Go Version

Windows 7 khong nen target bang Go moi nhat. Nen build bang **Go 1.20.x** vi cac ban Go sau do da thay doi baseline ho tro Windows cu.

Quy tac project:

- Toolchain chinh: `go1.20.x`.
- Khong dung API Go/runtime yeu cau Windows 10+.
- CI/build release phai co job cross-compile hoac build tren Windows.
- Neu can dependency moi, phai check no co van chay voi Go 1.20 hay khong.

### 1.2 Webview Reality Check

`webview/webview` hien dai tren Windows dua vao WebView2. Windows 7 co the chi chay duoc neu may co WebView2 Runtime phien ban cu tuong thich, va support cua Microsoft cho Windows 7 da ket thuc. Vi vay khong nen coi `webview/webview` la dam bao native Win7 100%.

Kien truc UI nen co 2 backend:

1. **Primary UI backend: WebView2/webview**  
   Dung khi may co runtime phu hop.

2. **Fallback UI backend: local HTML window strategy**  
   Neu WebView2 khong kha dung, mo UI bang browser/IE installed mode hoac dung mot backend native khac duoc pin rieng. Fallback nay co the kem dep hon nhung giu app chay duoc tren Win7.

Neu yeu cau bat buoc la "mot cua so rieng, khong browser tab", can lam mot proof-of-concept tren Windows 7 that truoc khi dong bang dependency.

## 2. High-Level Architecture

```text
Desktop EXE
  |
  +-- App Shell
  |     - window backend: webview primary, fallback optional
  |     - embedded assets: index.html, app.css, app.js
  |     - JS <-> Go direct binding
  |
  +-- Agent Core
  |     - session manager
  |     - LLM client
  |     - context builder
  |     - agentic loop
  |     - tool registry
  |
  +-- Safety Kernel
  |     - workspace policy
  |     - risk classifier
  |     - approval manager
  |     - audit log
  |     - protected path rules
  |
  +-- Local Storage
        - config
        - sessions
        - audit logs
        - model/provider credentials
```

## 3. Packaging Architecture

Ung dung nen duoc dong goi thanh mot `.exe` chinh:

- UI assets nhung bang `go:embed`.
- Config nam o `%APPDATA%\<AppName>\config.json`.
- Logs nam o `%APPDATA%\<AppName>\logs\`.
- Credentials/API keys nam rieng o `%APPDATA%\<AppName>\secrets.json`, co the ma hoa bang Windows DPAPI neu kha thi.
- Khong can local HTTP server cho normal mode.

Neu direct JS binding cua webview khong on dinh tren Win7 fallback, moi bat local loopback server `127.0.0.1` nhu compatibility mode, co random token per session.

## 4. UI Design

### 4.1 Technology

- HTML co ban.
- CSS conservative: flexbox co fallback float/table layout neu can.
- Vanilla JavaScript, tranh modern syntax neu engine cu.
- Khong dung bundler bat buoc.
- Khong dung framework nang.

### 4.2 Layout

```text
+------------------------------------------------------+
| Top bar: project root, model, connection status       |
+-------------------------------+----------------------+
| Chat Panel                    | Context/Safety Panel |
| - messages                    | - active workspace   |
| - tool cards                  | - current tool       |
| - streaming answer            | - pending approvals  |
| - input box                   | - audit events       |
+-------------------------------+----------------------+
```

### 4.3 Required UI States

- `Idle`
- `Thinking`
- `RunningTool`
- `WaitingForApproval`
- `Streaming`
- `Error`
- `Cancelled`

## 5. JS <-> Go Bridge

### 5.1 JS Calls Go

```js
window.agentSendPrompt(text)
window.agentCancelRun(runId)
window.approveToolCall(approvalId, decision, note)
window.selectWorkspace()
window.saveSettings(settings)
```

### 5.2 Go Calls JS

```go
ui.Emit("chat.message.delta", payload)
ui.Emit("tool.started", payload)
ui.Emit("tool.completed", payload)
ui.Emit("approval.required", payload)
ui.Emit("audit.event", payload)
ui.Emit("run.finished", payload)
ui.Emit("error", payload)
```

Khong nen noi chuoi JS bang string raw neu payload co text tu model/tool. Phai JSON encode payload roi goi mot dispatcher duy nhat:

```go
webview.Eval("window.__agentDispatch(" + jsonPayload + ")")
```

## 6. Core Data Schemas

### 6.1 Chat and Runs

```go
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type Message struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id,omitempty"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	ToolCallID string   `json:"tool_call_id,omitempty"`
}

type AgentRun struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	Status      RunStatus `json:"status"`
	UserPrompt  string    `json:"user_prompt"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at,omitempty"`
	Workspace   string    `json:"workspace"`
	Model       string    `json:"model"`
	ToolCalls   []ToolCallRecord `json:"tool_calls"`
	Approvals   []ApprovalRecord `json:"approvals"`
}
```

### 6.2 Tool Schema

```go
type ToolRisk string

const (
	RiskReadOnly    ToolRisk = "read_only"
	RiskWrite       ToolRisk = "write"
	RiskDelete      ToolRisk = "delete"
	RiskDestructive ToolRisk = "destructive"
	RiskNetwork     ToolRisk = "network"
	RiskSecret      ToolRisk = "secret"
)

type ToolSpec struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema  any        `json:"input_schema"`
	Risks       []ToolRisk  `json:"risks"`
	RequiresApproval bool   `json:"requires_approval"`
}

type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Reason    string         `json:"reason"`
}

type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	OK         bool   `json:"ok"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}
```

### 6.3 Approval Schema

```go
type ApprovalDecision string

const (
	ApprovalPending ApprovalDecision = "pending"
	ApprovalAllow   ApprovalDecision = "allow"
	ApprovalDeny    ApprovalDecision = "deny"
)

type ApprovalRequest struct {
	ID              string         `json:"id"`
	RunID           string         `json:"run_id"`
	ToolCallID      string         `json:"tool_call_id"`
	ToolName        string         `json:"tool_name"`
	Risk            ToolRisk       `json:"risk"`
	TargetPaths     []string       `json:"target_paths"`
	ReasonFromAgent string         `json:"reason_from_agent"`
	SafetySummary    string         `json:"safety_summary"`
	Arguments        map[string]any `json:"arguments"`
	CreatedAt        time.Time      `json:"created_at"`
	ExpiresAt        time.Time      `json:"expires_at"`
}

type ApprovalRecord struct {
	RequestID string           `json:"request_id"`
	Decision  ApprovalDecision `json:"decision"`
	UserNote  string           `json:"user_note,omitempty"`
	DecidedAt time.Time        `json:"decided_at"`
}
```

## 7. Tool Registry

### 7.1 Initial Safe Tools

| Tool | Risk | Approval | Notes |
|---|---|---:|---|
| `list_directory` | `read_only` | No | Chi trong workspace |
| `read_file` | `read_only` | No | Gioi han file size, block secrets |
| `search_text` | `read_only` | No | Gioi han result count |
| `get_file_info` | `read_only` | No | Metadata only |

### 7.2 Controlled Write Tools

| Tool | Risk | Approval | Notes |
|---|---|---:|---|
| `write_file` | `write` | Yes by default | Hien diff truoc khi apply |
| `replace_in_file` | `write` | Yes by default | Hien preview |
| `create_file` | `write` | Ask if path exists | Khong ghi de ngam |
| `rename_path` | `destructive` | Yes | Can reason |
| `delete_path` | `delete` | Always yes | Bat buoc reason + confirm |
| `run_command` | `destructive/network/unknown` | Depends on classifier | Disabled in MVP; xem command safety |

### 7.3 Delete Tool Contract

Agent khong duoc goi tool xoa neu khong co ly do ro rang.

```go
type DeletePathArgs struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Mode   string `json:"mode"` // "trash" | "permanent"
}
```

Policy:

- `Reason` bat buoc, toi thieu 20 ky tu.
- Mac dinh `Mode = "trash"` hoac soft delete vao `.agent_trash/`.
- Permanent delete bi tat mac dinh.
- Khong bao gio xoa ngoai workspace.
- Khong xoa protected paths.
- Khong xoa wildcard/rong/root path.
- User phai approve tren UI voi path, reason, size, file count.

## 8. Safety Kernel

### 8.1 Workspace Boundary

Moi tool file system phai resolve path bang `filepath.Abs` va kiem tra:

```go
func IsWithinWorkspace(workspace, target string) bool
```

Khong chap nhan:

- Path ngoai workspace.
- Symlink/junction thoat khoi workspace.
- Relative path nguy hiem nhu `..\..\`.
- Drive root nhu `C:\`.
- Windows system paths.

### 8.2 Protected Paths

Default protected:

- `.git/`
- `.svn/`
- `.hg/`
- `node_modules/` neu delete recursive.
- `vendor/` neu delete recursive.
- `.env`, `.env.*`
- `*.key`, `*.pem`, `*.pfx`, `*.cert`
- `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `go.sum` khi delete.
- `C:\Windows\`
- `C:\Program Files\`
- `C:\Program Files (x86)\`
- user profile root.

Protected path delete chi duoc khi user bat `advanced_unsafe_mode` trong settings va van can confirm rieng.

### 8.3 Risk Classifier

Moi tool call di qua:

```text
ToolCall
  -> validate schema
  -> normalize paths
  -> classify risk
  -> check workspace boundary
  -> check protected paths
  -> require approval if needed
  -> execute
  -> audit log
```

### 8.4 Approval UX

Approval dialog phai hien:

- Tool name.
- Target path(s).
- Risk level.
- Agent reason.
- File count/total size neu delete folder.
- Preview/diff neu write.
- Nut `Allow once`.
- Nut `Deny`.
- Optional note cua user.

Khong co `Always allow delete`.

## 8.5 Command Tool Safety

Shell/command execution la tool nguy hiem nhat vi no co the xoa file gian tiep ma khong goi `delete_path`. MVP nen **khong bat `run_command` mac dinh**. Khi them command tool, no phai di qua policy rieng.

### 8.5.1 Command Tool Contract

```go
type RunCommandArgs struct {
	Command        string `json:"command"`
	Cwd            string `json:"cwd"`
	Reason         string `json:"reason"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}
```

Rules:

- `Reason` bat buoc cho moi command.
- `Cwd` bat buoc nam trong workspace.
- Khong cho shell interactive.
- Timeout bat buoc, default 60s, max 300s.
- Environment variables duoc whitelist, khong pass toan bo host env neu khong can.
- Output truncate truoc khi dua lai model.

### 8.5.2 Command Risk Classes

| Class | Examples | Policy |
|---|---|---|
| `read_only` | `dir`, `type file.txt`, `git status`, `git diff`, `go env` | Allow if cwd inside workspace |
| `verify` | `go test`, `go test ./...`, `npm test`, `python -m pytest` | Ask first until user trusts project |
| `write_build_artifacts` | `go build`, `npm run build`, `tsc` | Ask if writes output |
| `package_install` | `npm install`, `go get`, `pip install` | Ask + network notice |
| `network` | `curl`, `wget`, `git clone`, package managers | Ask or deny by config |
| `destructive` | `del`, `erase`, `rd`, `rmdir`, `Remove-Item`, `rm`, `git clean`, `git reset --hard` | Require user approval before execute; critical patterns hard-deny |
| `script_unknown` | `powershell -Command ...`, `.bat`, `.cmd`, `.ps1`, `node script.js` | Ask + show warning; optional deny |
| `privileged` | `runas`, service control, registry edits, installer commands | Deny by default |

### 8.5.3 Delete Command Approval Gate

Moi command tool call bat buoc qua command filter:

```text
RunCommandArgs
  -> parse/normalize command
  -> classify command risk
  -> detect delete/write/network/script markers
  -> if delete-like:
       extract target paths if possible
       validate cwd/workspace boundary
       check protected paths
       require reason
       create ApprovalRequest
       wait user approval
       execute only if approved
  -> audit
```

Neu command co dau hieu xoa thi **khong duoc execute truoc approval**. Vi du:

- `del file.txt`
- `erase file.txt`
- `rd /s build`
- `rmdir /s build`
- `powershell Remove-Item file.txt`
- `rm -rf build`
- `git clean -fd`
- `git reset --hard`
- `cmd /C del file.txt`
- `echo ok && del file.txt`

Approval request cho delete-like command phai co:

- Full command.
- Cwd.
- Agent reason.
- Detected delete markers.
- Target paths neu parse duoc.
- Protected path warning neu co.
- Khuyen nghi dung `delete_path` soft-delete thay vi command delete neu phu hop.

Chi khi user bam `Allow once` thi command moi duoc execute. Neu user `Deny`, command khong chay va run nhan tool result bi deny.

### 8.5.4 Hard Deny Patterns

Mot so command qua nguy hiem nen block truoc approval, tru khi user bat `advanced_unsafe_mode` va type manual confirmation:

- `del`, `erase`, `rd`, `rmdir` with any non-trivial path.
- `powershell Remove-Item`, `rm`, `rimraf`.
- `git reset --hard`, `git clean -fd`, `git checkout -- .`.
- `format`, `diskpart`, `cipher /w`, `shutdown`, `reg delete`.
- Commands targeting `C:\Windows`, `Program Files`, user profile root, drive root.
- Any command containing wildcard delete patterns like `del *`, `rd /s`, `Remove-Item -Recurse`.

Even in advanced mode, command delete should prefer routing through `delete_path` so the app can calculate file count, protected path checks and soft-delete behavior.

### 8.5.5 Safer Execution Strategy

Do not run user/model command through `cmd.exe /C` unless needed. Prefer direct exec with parsed argv:

```text
command string -> parse -> executable + args -> classify -> execute
```

On Windows, parsing is hard. For MVP:

- Support allowlisted direct commands first: `go`, `git`, `npm`, `node`, `python`, `pytest`.
- Treat raw `cmd /C ...` and `powershell ...` as `script_unknown` or `destructive` depending content.
- Require approval for any shell metacharacter: `&`, `&&`, `|`, `>`, `>>`, `<`, `%`, `!`, backticks.

### 8.5.6 Approval UX For Commands

Approval dialog must show:

- Full command.
- Parsed executable and args if available.
- Working directory.
- Risk class.
- Agent reason.
- Files likely affected if inferable.
- Network/destructive warning badges.

Buttons:

- `Allow once`.
- `Deny`.
- No `always allow` for destructive/script/network commands.

### 8.5.7 Command Audit

Audit log for command:

```json
{
  "time": "2026-05-06T10:00:00Z",
  "run_id": "run_123",
  "tool": "run_command",
  "command": "go test ./...",
  "cwd": "C:\\project",
  "risk": "verify",
  "approval": "allow",
  "exit_code": 0,
  "duration_ms": 1240
}
```

## 9. Agentic Loop

```text
User prompt
  -> append message
  -> build context
  -> call LLM
  -> if final answer: stream to UI
  -> if tool call:
       validate tool call
       if approval required:
          pause run
          emit approval.required
          wait user decision
       execute tool
       append tool result
       continue LLM loop
  -> finish run
```

Gioi han:

- `max_turns_per_run`: 20.
- `max_tool_calls_per_run`: 50.
- `max_delete_calls_per_run`: 1 by default.
- `max_file_read_bytes`: configurable, default 256 KB.
- `max_search_results`: 100.

## 10. Audit Log

Moi tool call phai ghi audit JSONL:

```json
{
  "time": "2026-05-06T10:00:00Z",
  "run_id": "run_123",
  "tool_call_id": "tool_123",
  "tool": "delete_path",
  "risk": "delete",
  "target_paths": ["C:\\project\\tmp.txt"],
  "reason": "Remove generated temp file after build cleanup",
  "approval": "allow",
  "ok": true
}
```

Audit logs append-only. UI co tab xem lich su tool calls.

## 11. Secure Defaults

- File read/list/search only trong workspace.
- Write/delete/mac dinh can approval.
- Delete mac dinh la soft delete.
- Permanent delete disabled.
- Network disabled cho tool ban dau.
- Shell tool chua nen them o MVP.
- Secrets redaction truoc khi gui content cho LLM.
- Tool output truncate truoc khi dua vao model.
- User co nut cancel run.

## 12. MVP Scope

### Phase 1

- Desktop window + embedded UI.
- Chat UI.
- Settings: API key, model, workspace.
- LLM streaming.
- Tools: `list_directory`, `read_file`, `search_text`.
- Audit log.

### Phase 2

- `write_file`, `replace_in_file` voi diff approval.
- Soft delete tool voi approval.
- Protected path rules.
- Session persistence.

### Phase 3

- Provider abstraction: OpenAI/Anthropic/Gemini/local.
- Optional `run_command` tool with command classifier, approval and denylist.
- Tool plugins.
- MCP optional.
- Eval harness cho tool safety.
- Windows installer.

## 13. Recommended Project Layout

```text
cmd/agentdesk/
  main.go
internal/app/
  app.go
  config.go
internal/ui/
  bridge.go
  assets.go
  window.go
internal/agent/
  loop.go
  messages.go
  llm.go
internal/tools/
  registry.go
  filesystem.go
  delete.go
  write.go
internal/safety/
  policy.go
  risk.go
  approvals.go
  protected_paths.go
  audit.go
web/
  index.html
  app.css
  app.js
```

## 14. Acceptance Criteria

- App khoi dong tren Windows 7 test machine.
- UI hien chat va stream output.
- Agent doc/list/search file trong workspace.
- Agent khong doc duoc file ngoai workspace.
- Agent khong the xoa neu khong co `reason`.
- Delete request hien approval dialog.
- Deny delete khong thay doi file.
- Allow delete chuyen file vao trash/soft-delete area.
- Protected paths bi chan truoc approval.
- Command tool disabled by default.
- Destructive commands are denied or require explicit advanced confirmation.
- Audit log ghi day du tool call, approval va ket qua.
