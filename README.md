# BiAI AgentDesk

Production-oriented desktop AI-agent shell for Windows 7.

BiAI AgentDesk is a lightweight local agent app with an OpenAI-compatible `/v1`
provider, persistent multi-session chat, project instruction loading, guarded
workspace tools, command approval gates, audit logs, and a Windows 7-compatible
release build path.

## Windows 7 Build Rule

Build release binaries with Go `1.20.x`. Newer Go toolchains may compile, but Windows 7 runtime compatibility is not guaranteed.

```bat
set GOOS=windows
set GOARCH=386
go build -o agentdesk-win7-386.exe .\cmd\agentdesk
```

For 64-bit Windows 7:

```bat
set GOOS=windows
set GOARCH=amd64
go build -o agentdesk-win7-amd64.exe .\cmd\agentdesk
```

## Run

```bat
agentdesk-win7-386.exe
```

On Windows, the normal release executable is built as a GUI app and opens a dedicated `mshta.exe` window with no browser address bar. Internally it serves the embedded UI over `127.0.0.1` with a random per-run token. If `mshta.exe` exits immediately, the app falls back to the default browser.

If double-clicking appears to do nothing, check:

```text
%APPDATA%\BiAI\AgentDesk\agentdesk.log
%APPDATA%\BiAI\AgentDesk\startup-error.log
%APPDATA%\BiAI\AgentDesk\crash.log
```

Release also includes `*-debug-console.exe`. Run that from `cmd.exe` to see startup logs directly.

## Agent Commands

Use these in the chat box:

- `/list .`
- `/read path\to\file.txt`
- `/search keyword`
- `/cmd go test ./...`
- `/status`
- `/init`

Delete-like commands such as `del`, `rd /s`, `Remove-Item`, `git clean -fd`, and `git reset --hard` go through the command filter and require user approval before execution. Some system-level destructive commands are blocked before approval.

`/init` creates an `AGENTS.md` file in the selected workspace. On every AI turn,
the agent loads user/global and workspace `AGENTS.md`, `CLAUDE.md`, and
`GEMINI.md` instruction files, then combines them with dynamic machine context
such as OS, architecture, hostname, app data path, process cwd, and selected
workspace.

Conversation history is stored per session under:

```text
%APPDATA%\BiAI\AgentDesk\sessions
```

Audit logs and crash/startup diagnostics are stored under:

```text
%APPDATA%\BiAI\AgentDesk
```

## Verify

```sh
go test ./...
go build ./cmd/agentdesk
```
