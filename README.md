# BiAI AgentDesk

MVP desktop AI-agent shell for Windows 7.

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

On Windows, the app opens a dedicated `mshta.exe` window with no browser address bar. Internally it serves the embedded UI over `127.0.0.1` with a random per-run token.

## MVP Commands

Use these in the chat box:

- `/list .`
- `/read path\to\file.txt`
- `/search keyword`
- `/cmd go test ./...`

Delete-like commands such as `del`, `rd /s`, `Remove-Item`, `git clean -fd`, and `git reset --hard` go through the command filter and require user approval before execution. Some system-level destructive commands are blocked before approval.

## Verify

```sh
go test ./...
go build ./cmd/agentdesk
```

