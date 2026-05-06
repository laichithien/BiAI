# Cloned Research Repositories

Ngay clone/kiem tra: 2026-05-06.

Tat ca repo duoi day duoc clone shallow (`--depth 1`) vao `docs/research-repos/` de doi chieu truc tiep voi docs/code cong khai.

| Repo local | Upstream | Commit | Ly do clone |
|---|---|---:|---|
| `docs/research-repos/openai-codex` | https://github.com/openai/codex | `41505bc` | Codex CLI, sandbox, approvals, config, skills, MCP |
| `docs/research-repos/claude-code-public` | https://github.com/anthropics/claude-code | `5c0e4f9` | Claude Code public package/docs/plugins/examples/settings; khong phai full runtime source |
| `docs/research-repos/gemini-cli` | https://github.com/google-gemini/gemini-cli | `82f6ea5` | Gemini CLI source/docs, hooks, policy engine, sandbox, evals, subagents |
| `docs/research-repos/openclaw` | https://github.com/openclaw/openclaw | `06c490f8` | OpenClaw gateway/SDK/plugin/provider/MCP/security boundary architecture |
| `docs/research-repos/swe-agent` | https://github.com/SWE-agent/SWE-agent | `0f4f3bb` | Agent-Computer Interface, tools config, trajectory/replay, benchmark harness |
| `docs/research-repos/mini-swe-agent` | https://github.com/SWE-agent/mini-swe-agent | `bc85a45` | Minimal SWE-agent harness/runtime patterns |
| `docs/research-repos/swe-bench` | https://github.com/SWE-bench/SWE-bench | `f7bbbb2` | Software-engineering eval harness |
| `docs/research-repos/openhands` | https://github.com/All-Hands-AI/OpenHands | `3ec19e0` | Open agent runtime, skills/microagents, server/UI, trajectories, evaluations |
| `docs/research-repos/modelcontextprotocol` | https://github.com/modelcontextprotocol/modelcontextprotocol | `640b51f` | MCP spec/docs baseline |

## Important Notes

- OpenClaw search results tren web rat nhieu nhieu va co the co website khong dang tin. Tai lieu nay chi dung repo GitHub `openclaw/openclaw` lam nguon code.
- Claude Code public repo co settings, plugins, docs va examples. Kien truc runtime cua Claude Code van phai doi chieu voi official docs/SDK va hanh vi cong khai, khong suy dien nhu full source.
- Cac repo shallow clone khong bao gom full history. Neu can audit evolution/PR context, chay fetch full history rieng.
