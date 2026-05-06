# Agent Engineering Standards

Bo tieu chuan nay rut ra tu nghien cuu Codex CLI, Claude Code, Gemini CLI, OpenClaw va cac harness cong khai. Dung nhu checklist khi thiet ke, review hoac implement mot coding agent CLI.

## 1. Architecture Baseline

Mot agent CLI dat chuan nen co cac module doc lap:

- **CLI/UI**: interactive TUI, non-interactive command, status/progress, diff view.
- **Agent core**: turn loop, planning, context selection, memory, summarization.
- **Provider adapters**: OpenAI, Anthropic, Gemini, local model; khong de provider API ro ri vao business logic.
- **Tool registry**: filesystem, shell, git, test, web/browser, MCP, custom tools.
- **Policy broker**: sandbox, approvals, command/path/domain policy.
- **Patch engine**: tao/sua file bang diff co the review.
- **Harness**: run record, transcript, replay, eval runner, metrics.
- **Extension layer**: MCP, plugins, hooks, slash commands, skills/subagents.

## 2. Must-Have Capabilities

| Capability | Requirement |
|---|---|
| Repo instructions | Support `AGENTS.md`; optionally read `CLAUDE.md`, `GEMINI.md` with precedence ro |
| File ops | Read/search/list/write qua tool co audit |
| Shell ops | Command execution co timeout, cwd, env filtering, approval gate |
| Git awareness | Detect dirty worktree, khong revert user changes, summarize diff |
| Test verification | Discover/run test commands tu instruction file hoac project metadata |
| Sandbox | Workspace-write by default; network/destructive ops gated |
| MCP | Client support + per-server permission |
| Transcript | Luu message/tool/output/diff theo run |
| Replay | Co the replay hoac inspect run sau khi loi |
| Non-interactive mode | Can cho CI/harness/eval |

## 2.1 Evidence-Based Additions

Nhung capability sau nen coi la tieu chuan cao hon baseline vi xuat hien ro trong cac repo da clone:

- **Policy engine as code**: Gemini CLI co policy TOML; OpenClaw co effective tool policy. Nen de policy doc duoc, review duoc va test duoc.
- **Tool effective view**: OpenClaw tach `tools.catalog` va `tools.effective`; agent/UI nen biet tool nao ton tai va tool nao dang duoc phep trong session hien tai.
- **Checkpoint/rewind**: Gemini CLI co checkpointing/rewind/session management. Agent CLI nen co checkpoint toi thieu truoc khi sua file lon.
- **Hook contracts**: Claude Code/Gemini CLI deu co hook surface. Hook nen co schema stdin/stdout, timeout va audit log.
- **Trajectory/replay**: SWE-agent/mini-SWE-agent dat trajectory lam artifact trung tam. Production agent nen luu duoc transcript co the replay/debug.
- **Boundary-specific security tests**: OpenClaw chia CodeQL/security checks theo provider/plugin/MCP/gateway/secrets/network. Agent platform nen co test theo boundary thay vi chi unit test chung.

## 3. Safety Standards

### 3.1 Sandbox Modes

- `read-only`: chi doc/search, khong ghi file, khong shell side effects.
- `workspace-write`: duoc ghi trong workspace, shell gated theo command risk.
- `full-access`: chi cho local trusted/dev container; can banner ro.

Network nen la flag rieng:

- `network=false`: mac dinh cho task coding offline.
- `network=approval`: hoi khi can fetch docs/package.
- `network=true`: chi cho profile tin cay.

### 3.2 Command Risk Classes

| Class | Vi du | Policy |
|---|---|---|
| Read-only | `ls`, `rg`, `sed`, `git diff` | Allow |
| Local verify | `npm test`, `pytest`, `go test` | Allow hoac ask tuy repo |
| Package install | `npm install`, `pip install` | Ask |
| Network fetch | `curl`, `wget`, `git clone` | Ask/deny theo network |
| Destructive VCS | `git reset --hard`, `git clean -fd`, force push | Deny unless explicit user request |
| Secret/prod | cloud deploy, prod db, `kubectl delete` | Deny by default |

### 3.3 Secret and Prompt-Injection Controls

- Redact `.env`, keys, tokens before dua vao model.
- Treat web/file content as untrusted data.
- Tool output khong duoc tu cap quyen moi cho chinh no.
- MCP server can trust label: first-party, workspace, user-installed, remote.

## 4. Context Standards

### 4.1 Instruction File Template

Nen tao `AGENTS.md` voi cau truc:

```md
# Agent Instructions

## Project
Short description and architecture.

## Commands
- Install:
- Test:
- Lint:
- Build:

## Code Style
...

## Safety
- Do not edit:
- Ask before:

## Verification
Required checks before final answer.
```

### 4.2 Context Selection Rules

- Bat dau bang repo tree nho + search targeted.
- Doc file lien quan truoc khi sua.
- Khong dua binary/large generated files vao context.
- Output dai phai summarize va giu artifact path.
- Neu ket luan dua tren suy luan, ghi ro la suy luan.

## 5. Tool Standards

Moi tool phai co manifest:

```yaml
name: shell.exec
description: Run a shell command in a controlled workspace
permissions:
  - execute
  - filesystem.read
  - filesystem.write_optional
side_effects: true
timeout_ms: 120000
input_schema:
  cwd: string
  cmd: string
output_policy:
  max_chars: 12000
  preserve_exit_code: true
```

Tool runtime can:

- Validate input schema.
- Normalize errors.
- Capture exit code.
- Enforce timeout.
- Truncate output predictably.
- Attach artifact for full log.

## 6. Patch and Git Standards

- Moi file edit can di qua patch/diff engine.
- Khong sua unrelated files.
- Khong format ca repo neu task hep.
- Truoc khi final, chay `git diff --stat` va test lien quan neu co.
- Final answer can neu: files changed, verification, residual risk.

## 7. Harness Standards

### 7.1 Run Record

Bat buoc log:

- Agent version.
- Model/provider.
- Repo path + commit.
- Sandbox/approval/network mode.
- Task prompt.
- Tool calls.
- File diffs.
- Test commands/results.
- Final answer.
- Cost/latency neu co.

### 7.2 Eval Metrics

Do toi thieu:

- Task success/pass rate.
- Regression rate.
- Average turns.
- Tool-call count.
- Wall-clock latency.
- Token/cost.
- Approval count.
- Number of changed files/lines.
- Failure class: compile, test, wrong behavior, unsafe action, timeout.

### 7.3 Reproducibility

- Pin dependency versions.
- Use container/devcontainer where possible.
- Record env vars whitelist, OS, runtime versions.
- Keep verifier independent from agent.
- Reset workspace per task.

## 8. Extensibility Standards

### 8.1 MCP

- Per-server config.
- Explicit tool list.
- Permission per server/tool.
- Output size limits.
- Disable remote MCP by default in untrusted repos.

### 8.2 Hooks

Hook points nen gom:

- `before_tool_call`
- `after_tool_call`
- `before_file_write`
- `after_file_write`
- `before_final`
- `on_error`

Hooks can timeout, exit-code semantics va audit log.

### 8.3 Subagents/Skills

- Subagent phai co role, tools, scope, output contract.
- Khong de nhieu subagent sua cung file neu khong co owner ro.
- Skills nen co README/instructions + examples + tests neu tao tool.

## 9. Acceptance Checklist

Dung checklist nay truoc khi coi mot agent CLI la production-ready:

- [ ] Co sandbox mac dinh khong phai full-access.
- [ ] Approval prompt hien command/diff/scope.
- [ ] Co instruction file repo-local.
- [ ] Co MCP/tool permission boundary.
- [ ] Co transcript va run record.
- [ ] Co non-interactive mode cho harness.
- [ ] Co timeout va max turns.
- [ ] Co secret redaction.
- [ ] Co policy cho destructive commands.
- [ ] Co test verification workflow.
- [ ] Co replay/debug artifact.
- [ ] Co provider abstraction hoac ly do ro neu khong can.
- [ ] Co extension trust model.
- [ ] Co benchmark/eval suite toi thieu.

## 10. Recommended Defaults

- Sandbox: `workspace-write`.
- Network: `approval`.
- Approval: ask for destructive, install, network, outside-workspace writes.
- Context: instruction files + targeted search, no full repo dump.
- Patch: small diffs, verify after edit.
- Logging: transcript on by default, secrets redacted.
- Harness: every run emits JSON record + patch artifact.
