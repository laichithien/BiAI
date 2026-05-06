# Architecture Evidence Map

File nay map tung ket luan kien truc sang bang chung trong repo cong khai da clone.

## OpenAI Codex CLI

| Claim | Evidence |
|---|---|
| Repo instruction file la primitive chinh | `docs/research-repos/openai-codex/AGENTS.md`, `docs/research-repos/openai-codex/docs/agents_md.md` |
| Sandbox/approval la first-class | `docs/research-repos/openai-codex/docs/sandbox.md`, `docs/research-repos/openai-codex/docs/execpolicy.md`, `docs/research-repos/openai-codex/AGENTS.md` |
| Non-interactive execution ton tai | `docs/research-repos/openai-codex/docs/exec.md`, `docs/research-repos/openai-codex/docs/install.md` |
| Config/profile la boundary rieng | `docs/research-repos/openai-codex/docs/config.md`, `docs/research-repos/openai-codex/docs/example-config.md`, `docs/research-repos/openai-codex/codex-rs/config.md` |
| Skills/slash commands/MCP la extension surfaces | `docs/research-repos/openai-codex/docs/skills.md`, `docs/research-repos/openai-codex/docs/slash_commands.md`, `docs/research-repos/openai-codex/AGENTS.md` |

## Claude Code

| Claim | Evidence |
|---|---|
| Public repo la package/docs/plugins/examples, khong phai full runtime source | `docs/research-repos/claude-code-public/README.md`, repo tree |
| Settings/permissions co the quan ly enterprise | `docs/research-repos/claude-code-public/examples/settings/README.md`, `examples/mdm/README.md` |
| Bash sandbox khong bao phu tat ca tool | `docs/research-repos/claude-code-public/examples/settings/README.md` |
| Hooks co the validate Bash command | `docs/research-repos/claude-code-public/examples/hooks/bash_command_validator_example.py` |
| Plugins dong goi commands/agents/hooks/skills/MCP | `docs/research-repos/claude-code-public/plugins/README.md` |
| Multi-agent PR review la pattern cong khai | `docs/research-repos/claude-code-public/plugins/code-review/README.md`, `plugins/code-review/commands/code-review.md` |

## Gemini CLI

| Claim | Evidence |
|---|---|
| Project context qua `GEMINI.md` | `docs/research-repos/gemini-cli/GEMINI.md`, `docs/research-repos/gemini-cli/docs/cli/gemini-md.md` |
| Checkpoint/rewind/session management la first-class | `docs/research-repos/gemini-cli/docs/cli/checkpointing.md`, `docs/cli/rewind.md`, `docs/cli/session-management.md` |
| Policy engine co TOML va ap duoc theo subagent/tool | `docs/research-repos/gemini-cli/docs/reference/policy-engine.md`, `docs/core/subagents.md` |
| Hooks co schema stdin/stdout va event lifecycle | `docs/research-repos/gemini-cli/docs/hooks/reference.md` |
| Subagents co local/remote/extension variants | `docs/research-repos/gemini-cli/docs/core/subagents.md`, `docs/core/remote-agents.md`, `docs/extensions/index.md` |
| Sandbox duoc test theo none/docker/podman/seatbelt | `docs/research-repos/gemini-cli/docs/cli/sandbox.md`, `docs/integration-tests.md` |
| Evals la phan repo | `docs/research-repos/gemini-cli/evals/README.md` |

## OpenClaw

| Claim | Evidence |
|---|---|
| Core CLI/gateway/provider/plugin/MCP la cac boundary rieng | `docs/research-repos/openclaw/tsdown.config.ts`, `docs/research-repos/openclaw/AGENTS.md` |
| SDK expose agents/sessions/tools/artifacts/models/environments | `docs/research-repos/openclaw/packages/sdk/src/index.test.ts`, `packages/sdk/src/index.e2e.test.ts` |
| Tool model tach catalog/effective/invoke | `docs/research-repos/openclaw/packages/sdk/src/index.test.ts`, `packages/sdk/src/index.e2e.test.ts` |
| Security query tach theo boundary | `docs/research-repos/openclaw/.github/codeql/` |
| Sandbox co smoke test/deployment note | `docs/research-repos/openclaw/docker-compose.yml`, `.github/workflows/sandbox-common-smoke.yml` |
| Secrets/auth profiles khong nam trong repo | `docs/research-repos/openclaw/AGENTS.md` |

## Harnesses

| Claim | Evidence |
|---|---|
| SWE-agent xem tool/interface la Agent-Computer Interface | `docs/research-repos/swe-agent/docs/background/aci.md` |
| SWE-agent config dinh nghia tools/templates/models/env | `docs/research-repos/swe-agent/docs/config/config.md`, `docs/config/tools.md`, `docs/config/templates.md` |
| Trajectory/replay/demo la primitive eval | `docs/research-repos/swe-agent/docs/config/demonstrations.md`, `docs/usage/trajectories.md` |
| mini-SWE-agent giu runtime/harness nho hon | `docs/research-repos/mini-swe-agent/README.md`, `minisweagent/`, `tests/test_data/*.traj.json` |
| SWE-bench la harness theo instance + verifier | `docs/research-repos/swe-bench/README.md`, repo `swebench/` |
| OpenHands co runtime/gui/server, skills/microagents, trajectory analytics | `docs/research-repos/openhands/README.md`, `skills/README.md`, `openhands/analytics/EVENTS.md` |
