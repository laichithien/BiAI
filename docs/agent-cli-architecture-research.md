# Research: Agent CLI Architecture

Ngay nghien cuu: 2026-05-06.

Pham vi: OpenAI Codex CLI, Anthropic Claude Code, Google Gemini CLI, OpenClaw va lop harness dung de chay/evaluate/quan sat agent. Muc tieu la rut ra dac trung kien truc, tinh nang can co, tieu chuan an toan va cac mau thiet ke co the ap dung cho mot agent coding CLI.

Repo cong khai da clone ve workspace de doi chieu truc tiep:

- `docs/research-repos/openai-codex`
- `docs/research-repos/claude-code-public`
- `docs/research-repos/gemini-cli`
- `docs/research-repos/openclaw`
- `docs/research-repos/swe-agent`
- `docs/research-repos/mini-swe-agent`
- `docs/research-repos/swe-bench`
- `docs/research-repos/openhands`
- `docs/research-repos/modelcontextprotocol`

## 1. Tom Tat Dieu Hanh

Mot agent CLI hien dai khong chi la wrapper quanh LLM. Kien truc chung da hoi tu thanh 7 lop:

1. **Interface loop**: terminal TUI/REPL, non-interactive mode, IDE/editor integration, slash commands.
2. **Context engine**: doc huong dan repo (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`), history, memory, file discovery, summarization/compaction.
3. **Tool runtime**: shell, read/write file, search, web, git, test runner, browser/Playwright, MCP/external tools.
4. **Planning and execution policy**: task decomposition, plan updates, subagents, checkpoints, resumability.
5. **Safety layer**: sandbox, permission prompt, allowlist/denylist, network gate, secret redaction, destructive-command guard.
6. **Observability and harness**: transcript, event stream, logs, replay, metrics, benchmark/evaluation harness.
7. **Extensibility**: MCP, plugins/extensions, hooks, slash commands, custom agents/skills/profiles.

Ket luan quan trong:

- **MCP da tro thanh chuan tich hop de facto** cho tools/context ben ngoai. Codex, Claude Code va Gemini CLI deu co cau chuyen MCP rieng.
- **Permissioning khong the tach khoi UX**. Agent tot can cho nguoi dung thay lenh sap chay, file sap sua, rui ro network/destructive command va co co che chap thuan nhanh.
- **Harness la lop khac voi CLI**. CLI phuc vu nguoi dung; harness phuc vu viec chay co kiem soat, danh gia, replay, benchmark, audit va chuan hoa moi truong.
- **Repo-local instruction file la primitive cot loi**. Moi he sinh thai dat ten khac nhau, nhung deu can mot "contract" trong repo de gan coding style, test command, convention va ranh gioi an toan.

## 2. OpenAI Codex CLI

Nguon chinh: OpenAI Codex docs va repo `openai/codex`.

### 2.1 Vai Tro

Codex CLI la coding agent chay trong terminal, tap trung vao:

- Doc, sua va tao file trong workspace.
- Chay lenh/test/build theo approval policy.
- Lam viec voi Git va PR/review workflow.
- Ho tro `AGENTS.md` de nap huong dan repo.
- Cau hinh sandbox, approval mode, MCP server, profiles va model.
- Co che resumable conversation/history va compact context.

### 2.2 Kien Truc Suy Ra

```text
User terminal / IDE
        |
        v
Codex CLI runtime
  - TUI / non-interactive exec
  - prompt composer
  - context loader: AGENTS.md + files + history
  - planner / turn manager
        |
        v
Tool broker
  - shell exec
  - file patching
  - web/search where available
  - MCP clients
  - git/test/build commands
        |
        v
Policy layer
  - sandbox mode
  - approval policy
  - command classification
  - writable roots / network rules
        |
        v
Workspace + external services
```

### 2.3 Thanh Phan Chinh

**Bang chung tu repo**

- `docs/research-repos/openai-codex/AGENTS.md` co instruction ve sandbox env vars nhu `CODEX_SANDBOX_NETWORK_DISABLED` va `CODEX_SANDBOX=seatbelt`.
- `docs/research-repos/openai-codex/docs/exec.md` va `docs/execpolicy.md` tach non-interactive execution va execution policy.
- `docs/research-repos/openai-codex/docs/config.md`, `docs/example-config.md`, `codex-rs/config.md` cho thay config/profile la primitive rieng.
- `docs/research-repos/openai-codex/docs/skills.md` va `docs/slash_commands.md` cho thay extensibility khong chi qua MCP.
- `docs/research-repos/openai-codex/AGENTS.md` nhac truc tiep den MCP connection manager trong `codex-rs/codex-mcp`.

**Instruction hierarchy**

- `AGENTS.md` la file huong dan repo. Thuc te day la "developer documentation for agents": coding style, commands, tests, conventions, caveats.
- Agent can merge nhieu nguon instruction: system/developer/user, global config, repo-local instruction, turn context.
- Quy tac tot: instruction gan code nen de trong repo; instruction ca nhan/deployment nen de trong config nguoi dung.

**Configuration**

- Codex dung config file local, profile va cac tuy chon runtime de chon model, reasoning effort, sandbox/approval policy, MCP server.
- Cau hinh theo profile giup tach workflow: research read-only, normal coding, high-trust local automation, CI-like execution.

**Sandbox and approvals**

- Sandbox quy dinh agent duoc doc/ghi/chay lenh trong pham vi nao.
- Approval policy quy dinh khi nao phai hoi nguoi dung truoc khi chay command, dung network, ghi ngoai workspace hoac lam thao tac rui ro.
- Day la lop can co neu agent co quyen shell that. Khong the chi dua vao prompt.

**Execution model**

- Agent lap vong: hieu yeu cau -> doc context -> lap ke hoach -> sua file -> chay test -> tra ket qua.
- Chinh sua file nen co audit trail ro rang, ideally qua patch/diff.
- Command outputs can duoc tom tat cho model va nguoi dung; output dai nen compact de tranh phinh context.

**Extensibility**

- MCP cho phep them tool/context nhu GitHub, docs, database, browser, issue tracker.
- Subagents/agent profiles co the tach task doc-only, code-edit, verification.

### 2.4 Diem Manh

- Security posture ro: sandbox + approvals la first-class.
- `AGENTS.md` don gian, repo-native, de version control.
- Workflow CLI gan voi thoi quen engineer: `rg`, shell, git diff, test runner.
- Phu hop lam coding agent tong quat trong repo co san.

### 2.5 Rui Ro/Khoang Trong

- Chat-driven agent de bi phinh context khi task dai; can compaction/replay tot.
- MCP/tool ecosystem can policy rieng cho tung tool, neu khong agent co the bypass sandbox qua tool ben ngoai.
- Neu approval UX qua on ao, nguoi dung se chap thuan may moc; neu qua long, rui ro tang.

## 3. Anthropic Claude Code

Nguon chinh: Claude Code docs, Anthropic engineering docs, repo/SDK cong khai va bao cao phan tich cong khai. Khong sao chep hay tai tao noi dung ro ri.

### 3.1 Vai Tro

Claude Code la agentic coding tool chay trong terminal, co kha nang:

- Hieu codebase, sua file, chay command.
- Quan ly memory/instructions qua `CLAUDE.md`.
- Slash commands, hooks, settings, permissions.
- MCP integration.
- Subagents va skills de chuyen mon hoa hanh vi.
- SDK/headless mode de nhung vao automation.

### 3.2 Kien Truc Suy Ra

```text
Terminal / IDE / SDK
        |
        v
Claude Code agent host
  - interactive conversation loop
  - memory loader: CLAUDE.md + user/project settings
  - tool planner
  - slash command router
  - subagent/skill selection
        |
        v
Permission + hook system
  - tool permission decisions
  - lifecycle hooks
  - command allow/deny rules
        |
        v
Tools
  - filesystem
  - shell
  - search
  - MCP servers
  - VCS/test tooling
```

### 3.3 Thanh Phan Chinh

**Bang chung tu repo cong khai**

- `docs/research-repos/claude-code-public/README.md` mo ta Claude Code la agentic coding tool trong terminal/IDE/GitHub.
- `examples/settings/*.json` va `examples/mdm/*` cho thay permission/settings co the quan ly ca o muc enterprise.
- `examples/settings/README.md` ghi ro `sandbox` property ap dung cho Bash tool, khong mac dinh bao phu Read/Write/Web/MCP/hooks/internal commands. Day la chi tiet quan trong khi thiet ke threat model.
- `examples/hooks/bash_command_validator_example.py` la mau hook chan/validate Bash command.
- `plugins/README.md` xac nhan plugin co the dong goi commands, agents, hooks, skills va MCP servers.
- `plugins/code-review` la mau dung nhieu agent song song de review PR, trong do co buoc thu thap `CLAUDE.md` lien quan va validate findings bang subagents.

**Memory via `CLAUDE.md`**

- `CLAUDE.md` dong vai tro giong repo contract cho agent.
- Nen ghi ro build/test/lint commands, style, dependency policy, khu vuc cam dung, conventions va cach verify.

**Permissions**

- Claude Code co co che permissions/settings de dieu khien tool use.
- Pattern can hoc: permission khong chi la "yes/no", ma gom scope theo tool, duong dan, command, project va mode.

**Hooks**

- Hooks cho phep chen logic tai cac diem vong doi: truoc/sau tool call, notification, session events.
- Gia tri kien truc: enterprise/co quan co the ep lint, audit, secret scanning, logging, policy check ma khong sua core agent.

**Subagents and skills**

- Subagents giup chuyen mon hoa context va tool permissions cho task cu the, vi du review, test, migration, docs.
- Skills dong goi workflow/knowledge co the tai su dung.
- Nguyen tac: delegate task bounded, co ownership ro, output co the kiem tra.

**SDK/headless automation**

- SDK cho phep dung agent trong script/CI/internal tools.
- Khi vao automation, can them timeout, max turns, output schema, replay log va fail-fast policy.

### 3.4 Diem Manh

- He sinh thai extensibility ro: MCP, hooks, slash commands, subagents, skills.
- Memory/project instruction co tinh thuc dung cao.
- Phu hop ca interactive coding va automation.

### 3.5 Rui Ro/Khoang Trong

- Hook/plugin manh co the mo rong attack surface; can signature/trust boundary.
- Subagent neu khong gioi han ownership de gay conflict, duplicate work, ton token.
- Settings phuc tap can preset an toan mac dinh.

## 4. Google Gemini CLI

Nguon chinh: Google Gemini CLI docs va repo `google-gemini/gemini-cli`.

### 4.1 Vai Tro

Gemini CLI la open-source AI agent CLI cua Google, tap trung vao terminal developer workflow:

- Query/sua codebase.
- Dung tools nhu read/write file, shell, web/search va MCP.
- Memory/instruction qua `GEMINI.md`.
- Sandbox execution.
- Authentication/cloud integration voi Google ecosystem.
- Extension system.

### 4.2 Kien Truc Suy Ra

```text
CLI package / terminal UI
        |
        v
Gemini CLI core
  - prompt + context management
  - GEMINI.md discovery
  - model client
  - extension loader
        |
        v
Tool registry
  - built-in tools
  - MCP tools
  - shell/file tools
  - web/Google search tools
        |
        v
Sandbox / auth / deployment config
        |
        v
Workspace + Google services + MCP servers
```

### 4.3 Thanh Phan Chinh

**Bang chung tu repo**

- `docs/research-repos/gemini-cli/docs/cli/gemini-md.md`: project context file.
- `docs/research-repos/gemini-cli/docs/cli/checkpointing.md`, `rewind.md`, `session-management.md`: checkpoint/resume/rewind la first-class.
- `docs/research-repos/gemini-cli/docs/reference/policy-engine.md`: policy engine bang TOML, co the ap vao tool/subagent.
- `docs/research-repos/gemini-cli/docs/hooks/reference.md`: hook schema, stdin/stdout contract, tool/model/lifecycle hooks.
- `docs/research-repos/gemini-cli/docs/core/subagents.md`: built-in/custom subagents, tool allowlists, inline MCP server per subagent, remote subagents qua Agent-to-Agent.
- `docs/research-repos/gemini-cli/docs/cli/sandbox.md` va `docs/integration-tests.md`: sandbox none/docker/podman/seatbelt va integration tests theo sandbox modes.
- `docs/research-repos/gemini-cli/evals/README.md`: evals la phan repo, gan voi chat/CLI regression.

**Open-source distribution**

- Vi Gemini CLI la open source, co the audit tool registry, package layout, extension mechanism va sandbox behavior tot hon cac tool dong.
- Day la loi the lon cho org can fork, self-host policy, hoac viet extension noi bo.

**`GEMINI.md`**

- Cung mau voi `AGENTS.md`/`CLAUDE.md`: huong dan repo-local cho agent.
- Khi thiet ke agent rieng, nen support nhieu ten file pho bien hoac co adapter de tranh lock-in.

**Extensions**

- Extension co the dong goi tool, command, prompt/context va config.
- Extension can manifest ro rang, versioning, permission declaration, compatibility range.

**Sandboxing**

- Gemini CLI docs de cap deployment/sandbox. Diem can rut ra: sandbox phai gan voi platform runtime, khong chi gan voi LLM.
- Nen co mac dinh read/write workspace-only, network gated, command risk classification.

### 4.4 Diem Manh

- Open-source, de hoc mau kien truc va tich hop.
- Gan manh voi Google Search/Cloud ecosystem.
- `GEMINI.md` va extension system phu hop repo/team workflow.

### 4.5 Rui Ro/Khoang Trong

- Ket noi cloud/search manh lam tang yeu cau privacy va data boundary.
- Extension ecosystem can trust model ro.
- Open-source code khong tu dong dam bao sandbox production-grade; can threat model rieng.

## 5. OpenClaw

Nguon chinh: repo cong khai `openclaw/openclaw` da clone tai `docs/research-repos/openclaw`.

### 5.1 Vai Tro

OpenClaw la mot he thong agent/gateway/plugin lon hon mot CLI don le. Gia tri nghien cuu nam o viec no gom nhieu primitive cua agent platform:

- Core CLI va gateway.
- SDK cho agent/session/tool invocation.
- Provider plugin/runtime hooks.
- Tool catalog/effective tools/invoke.
- Subagent registry va hooks.
- MCP process/tool boundary.
- Sandbox smoke tests va runtime boundary tests.
- Secret/auth profiles tach khoi repo.
- CodeQL/security queries theo boundary: provider, plugin, MCP, gateway, network SSRF, secrets.

### 5.2 Bang Chung Tu Repo

- `tsdown.config.ts` khai bao nhieu runtime entry: `agents/auth-profiles.runtime`, `agents/model-catalog.runtime`, `subagent-registry.runtime`, `plugins/provider-runtime.runtime`, `mcp/plugin-tools-serve`, `agents/pi-embedded-runner/effective-tool-policy`.
- `packages/sdk/src/index.e2e.test.ts` va `packages/sdk/src/index.test.ts` the hien API Gateway: `agents.list`, `agents.create`, `agent`, `agent.wait`, `tools.catalog`, `tools.effective`, `tools.invoke`, `artifacts.*`, `models.status`, `environments.create`.
- `.github/codeql/*` chia security/quality query theo boundary: agent runtime, channel runtime, provider runtime, plugin trust, MCP process/tool, network SSRF, auth secrets, session diagnostics, UI control plane.
- `docker-compose.yml` co note ve `agents.defaults.sandbox` va Docker CLI trong image.
- `AGENTS.md` ghi ro ownership boundary: core owns generic loop; provider plugins own auth/catalog/runtime hooks; extension-owned behavior stays extension-owned.
- `AGENTS.md` ghi secrets trong `~/.openclaw/credentials/` va auth profiles trong `~/.openclaw/agents/<agentId>/agent/auth-profiles.json`.

```text
OpenClaw
  - CLI/control plane
  - Gateway/session API
  - SDK client
  - agent runtime / embedded runner
  - provider catalog/runtime plugins
  - tool catalog/effective policy/invocation
  - MCP process/tool serving
  - artifact/session/event stream
  - sandbox/security boundary checks
```

### 5.3 Gia Tri Rut Ra

- Mot agent platform tot nen co **provider abstraction**: model/provider co auth, catalog va runtime hooks rieng.
- Tool calling nen tach thanh **capability registry** thay vi hard-code vao prompt.
- `tools.catalog`, `tools.effective`, `tools.invoke` la bo API tot de tach "tool ton tai" voi "tool duoc phep trong session hien tai".
- Gateway API nen co `agent.run`, `agent.wait`, events va artifacts de automation khong phu thuoc vao TUI.
- Boundary testing/security query nen chia theo component ownership: provider, plugin, MCP, gateway, runtime, secrets, network.

### 5.4 Rui Ro/Khoang Trong

- Surface lon hon CLI nen attack surface cung lon: gateway, plugin, MCP, provider, channels, UI.
- Provider/plugin extensibility manh can trust model va version policy ro.
- Neu gateway co `tools.invoke`, permission phai gan voi session/user/agent, khong chi voi tool name.

## 6. Harness Layer

Tu "harness" trong agent engineering co nhieu nghia. Trong pham vi nay, can tach 5 loai:

Bang chung tu repo:

- SWE-agent dung khai niem **Agent-Computer Interface (ACI)**: tool/interface design anh huong truc tiep den ket qua agent.
- SWE-agent co config YAML de dinh nghia tools, prompts/templates, demonstrations, model behavior va environment.
- SWE-agent/mini-SWE-agent ghi trajectory JSON va co replay/demo conversion, rat quan trong cho reproducibility.
- SWE-bench la eval harness theo instance: repo snapshot, issue, patch/test verification.
- OpenHands co runtime/gui/server, skills/microagents, conversation trajectory, analytics/error terminal states va repo benchmark rieng.
- Gemini CLI co `evals/`, checkpointing, telemetry va integration tests tren nhieu sandbox mode.

### 6.1 Execution Harness

Chay agent trong moi truong co kiem soat:

- Workspace tam thoi hoac container.
- Timeout, max turns, max cost.
- Network policy.
- Secret isolation.
- Reproducible dependencies.
- Log moi tool call va diff.

Ap dung: CLI production, CI, benchmark, automated issue fixing.

### 6.2 Evaluation Harness

Danh gia agent tren tap task:

- Input: repo snapshot + issue/task + expected behavior/test.
- Runner: khoi tao moi truong, chay agent, chay verifier.
- Metrics: pass rate, cost, latency, turns, tool-call count, human approvals, regression rate.
- Artifacts: patch, transcript, stdout/stderr, final answer, test logs.

Vi du lien quan: SWE-bench style harness, Scion-style harness, research ve "agent harnesses".

### 6.3 Tool Harness

Chuan hoa tool interface:

- Tool schema.
- Input validation.
- Output truncation/summarization.
- Error normalization.
- Permission decision.
- Replay support.

MCP la mot dang tool/context harness cross-application.

### 6.4 Policy Harness

Bao quanh agent bang policy bat buoc:

- Pre-command classifier.
- Path allowlist/denylist.
- Secret scanner truoc khi gui context ra model.
- Prompt-injection detector cho web/file content.
- License/compliance filter.
- Audit log immutable.

### 6.5 Human-in-the-loop Harness

Quan ly diem can nguoi dung quyet:

- Approval prompt co diff/command/context.
- Batch approvals voi scope nho.
- Checkpoint/revert.
- Review mode.
- Escalation khi command destructive, network, auth, production resource.

## 7. Ma Tran Tinh Nang

| Nang luc | Codex CLI | Claude Code | Gemini CLI | OpenClaw | Chuan chung |
|---|---:|---:|---:|---:|---|
| Terminal interactive agent | Co | Co | Co | Co | Bat buoc |
| Non-interactive/headless | Co | Co/SDK | Co | Tuy du an | Nen co |
| Repo instruction file | `AGENTS.md` | `CLAUDE.md` | `GEMINI.md` | Nen support nhieu ten | Bat buoc |
| File read/write tools | Co | Co | Co | Co | Bat buoc |
| Shell command tool | Co | Co | Co | Co | Bat buoc nhung gated |
| Sandbox | Co | Permission/sandbox policy | Co | Tuy du an | Bat buoc |
| Approval policy | Co | Co | Co | Tuy du an | Bat buoc |
| MCP | Co | Co | Co | Nen co | Bat buoc cho extensibility |
| Hooks | Han che/tuy runtime | Manh | Extension-driven | Tuy du an | Nen co |
| Subagents/skills | Co/Profiles | Co | Extension/commands | Tuy du an | Nen co cho task lon |
| Web/search | Tuy moi truong | Tuy tool | Manh voi Google | Tuy provider | Optional, gated |
| Checkpoint/replay | Nen co | Co mot phan qua transcript/session | Nen co | Can them | Bat buoc cho harness |
| Provider abstraction | Chu yeu OpenAI | Chu yeu Anthropic | Chu yeu Gemini | Manh | Nen co neu xay agent rieng |

## 8. Tieu Chuan Ky Thuat Chung Rut Ra

### 8.1 Context Contract

Agent phai co co che nap instruction theo thu tu uu tien ro:

1. System/developer policy.
2. Organization/team policy.
3. User config.
4. Repo instruction file.
5. Task prompt.
6. Runtime evidence: files, command output, test logs.

Can log duoc instruction nao da duoc nap, tu file nao, o commit/path nao.

### 8.2 Tool Contract

Moi tool can khai bao:

- `name`, `description`, input schema, output schema.
- Quyen yeu cau: read, write, execute, network, secret access.
- Scope: duong dan, domain, command prefix, resource.
- Side effects.
- Timeout/retry policy.
- Output truncation policy.

### 8.3 Permission Contract

Permission nen gom:

- `read-only`: doc/search khong sua.
- `workspace-write`: duoc sua workspace, command rui ro phai hoi.
- `full-access`: chi dung o moi truong tin cay.
- `network`: flag rieng, khong gop chung voi file write.
- `destructive`: flag rieng cho delete/reset/rebase/production commands.

### 8.4 Patch Contract

Chinh sua code nen sinh duoc:

- Diff co the review.
- Danh sach file thay doi.
- Ly do thay doi.
- Cach verify.
- Khong revert thay doi cua nguoi dung khong lien quan.

### 8.5 Harness Contract

Moi lan agent chay nen co run record:

```json
{
  "run_id": "uuid",
  "agent": "name/version",
  "model": "provider/model",
  "repo": "path or remote",
  "commit": "sha",
  "task": "text or issue id",
  "policy": {
    "sandbox": "workspace-write",
    "network": false,
    "approval": "on-request"
  },
  "artifacts": {
    "transcript": "path",
    "patch": "path",
    "logs": "path"
  },
  "metrics": {
    "turns": 0,
    "tool_calls": 0,
    "latency_ms": 0,
    "cost_usd": 0,
    "tests_passed": null
  }
}
```

## 9. Kien Truc De Xuat Neu Xay Agent CLI Rieng

```text
packages/
  cli/
    tui/
    commands/
  core/
    agent_loop/
    context/
    planning/
    memory/
  tools/
    filesystem/
    shell/
    git/
    browser/
    mcp/
  policy/
    sandbox/
    approvals/
    command_classifier/
    secret_scanner/
  harness/
    runner/
    replay/
    eval/
    metrics/
  providers/
    openai/
    anthropic/
    gemini/
    local/
```

### Luong Xu Ly Chuan

1. Khoi tao session, doc config va repo instruction files.
2. Tao context toi thieu: task + tree + relevant files, khong dump ca repo.
3. Lap plan ngan, xac dinh tool can dung.
4. Moi tool call qua policy broker.
5. File write qua patch engine, luu diff.
6. Chay verify theo instruction repo.
7. Neu loi, lap vong sua co gioi han.
8. Ket thuc bang summary + changed files + tests + residual risks.
9. Luu transcript/run record cho replay/eval.

## 10. Cac Mau Thiet Ke Nen Ap Dung

- **Policy-before-tool**: moi tool call di qua policy broker.
- **Context as evidence**: model chi ket luan dua tren file/log da doc, khong doan tren ten file.
- **Small patch loop**: sua nho, test nhanh, lap lai.
- **Instruction file compatibility**: support `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`.
- **MCP as boundary**: external capabilities vao qua MCP/tool registry co schema va permission.
- **Transcript-first observability**: co the replay duoc mot run ma khong can nho lai hoi thoai.
- **Human approval as scoped token**: approval nen co scope va thoi han, khong phai "yes forever".

## 11. Anti-Patterns

- Dua shell full-access cho model ma khong sandbox.
- Nap toan bo repo vao context thay vi retrieval co muc tieu.
- Hard-code provider API vao tool loop.
- Tool output khong truncate, lam tran context.
- Approval prompt chi hien "Allow?" ma khong hien command/diff/scope.
- Plugin/hook khong co trust boundary.
- Khong ghi transcript/diff/test logs nen khong audit duoc.
- Benchmark chi do pass/fail, bo qua chi phi, latency, approvals va rui ro.

## 12. Nhan Dinh Rieng Ve "Leaked Claude Code"

Co cac bai viet/phan tich cong khai noi ve viec code hoac bundle cua Claude Code bi soi/phan tich. Tai lieu nay khong dung noi dung bi ro ri lam nguon chinh. Nhung o muc kien truc, nhung diem duoc cong khai qua tai lieu chinh thuc da du de rut ra:

- Agent host co tool loop va permission layer.
- Memory/project instructions la primitive quan trong.
- Hooks va MCP la extension boundary.
- Subagents/skills la cach dong goi chuyen mon.
- SDK/headless mode bien CLI thanh runtime co the nhung vao automation.

Neu can audit sau hon cho muc dich noi bo, nen chi dung nguon co giay phep ro rang, repo chinh thuc, SDK public va runtime behavior do chinh minh quan sat trong moi truong hop le.
