# Safety Test Plan: Windows 7 Go Desktop Agent

Muc tieu: dam bao agent khong the doc/xoa/ghi de nham file quan trong, khong bypass workspace boundary, va moi hanh dong nguy hiem deu can user approval co ly do ro.

## 1. Threat Model

| Threat | Example | Required Defense |
|---|---|---|
| Accidental delete | Agent xoa folder source thay vi temp file | Delete approval + soft delete + file count preview |
| Path traversal | `..\..\Windows\System32` | Abs path + workspace boundary |
| Symlink/junction escape | Workspace co junction tro ra `C:\Users` | Resolve final path + block escape |
| Prompt injection in file | File noi "delete project" | Tool policy khong bi thay doi boi file content |
| Missing reason | Agent goi delete voi reason rong | Schema validation fails |
| Protected path delete | `.git`, `.env`, lockfiles | Protected path blocker |
| Silent overwrite | `write_file` de len file co san | Diff approval |
| Command deletes files | Agent chay `del`, `rd`, `Remove-Item`, `git clean` | Command filter detects delete, asks approval, then executes only if approved |
| Command hides delete in shell chain | `echo ok && del *.go` | Shell metacharacter detection + destructive deny |
| Script performs unknown side effects | `.bat`, `.ps1`, `node script.js` | `script_unknown` risk + approval/deny |
| Secret leakage | Doc `.env` roi gui LLM | Secret path/content redaction |
| UI injection | Tool output co quote/script | JSON encode dispatch, HTML escape render |

## 2. Unit Tests

### 2.1 Path Boundary

- `C:\project\a.txt` inside `C:\project` => allow.
- `C:\project\..\project\a.txt` => normalize then allow.
- `C:\project\..\Windows\win.ini` => deny.
- `C:\` target => deny.
- Empty path => deny.
- Relative path with `..\` escaping workspace => deny.
- UNC path `\\server\share` => deny by default.
- Symlink/junction escaping workspace => deny.

### 2.2 Delete Validation

- Delete without `reason` => deny.
- Delete reason shorter than threshold => deny.
- Delete protected path `.git/config` => deny.
- Delete directory with many files => approval requires count/size.
- Delete outside workspace => deny before approval.
- Delete with wildcard `*` => deny.
- Permanent delete while `allowPermanentDelete=false` => deny.
- Soft delete creates recoverable artifact.

### 2.3 Write Validation

- Write new file in workspace => approval optional depending policy.
- Overwrite existing file => approval required with diff.
- Write outside workspace => deny.
- Write `.env` => deny or explicit advanced approval.
- Replace in file with no match => no-op, audit logged.

### 2.4 UI Bridge

- Payload containing quotes/newlines/script tags is JSON encoded.
- UI renderer uses `textContent` for untrusted text.
- Approval id cannot be guessed to approve other run.
- Expired approval cannot execute.

### 2.5 Command Classification

- `dir` in workspace => `read_only`.
- `git status` => `read_only`.
- `go test ./...` => `verify`.
- `npm install` => `package_install` + `network`.
- `del file.txt` => `destructive`.
- `rd /s build` => `destructive`.
- `powershell Remove-Item file.txt` => `destructive`.
- `git clean -fd` => `destructive`.
- `git reset --hard` => `destructive`.
- `echo ok && del file.txt` => `destructive`.
- `cmd /C del file.txt` => `destructive`.
- `powershell -Command Get-ChildItem` => `script_unknown`.
- `reg delete HKCU\Software\Test` => `privileged`/deny.
- Command cwd outside workspace => deny before approval.

## 3. Integration Tests

### 3.1 Safe Read Flow

1. Select temp workspace.
2. Ask agent to list files.
3. Assert `list_directory` runs without approval.
4. Audit log contains read-only tool call.

### 3.2 Delete Deny Flow

1. Create `tmp.txt`.
2. Simulate agent delete call with valid reason.
3. Assert UI receives `approval.required`.
4. Deny approval.
5. Assert file still exists.
6. Assert audit log has `approval=deny`.

### 3.3 Delete Allow Soft Flow

1. Create `tmp.txt`.
2. Simulate delete call.
3. Approve once.
4. Assert original path removed.
5. Assert file exists in `.agent_trash`.
6. Assert audit log records source and trash path.

### 3.4 Protected Delete Flow

1. Create `.git/config` in workspace fixture.
2. Simulate delete call.
3. Assert request is denied before approval.
4. Assert UI receives safety error, not approval prompt.

### 3.5 Command Delete Block Flow

1. Create `tmp.txt`.
2. Simulate `run_command` with `del tmp.txt`.
3. Assert command is classified `destructive`.
4. Assert UI receives `approval.required` before execution.
5. Assert `tmp.txt` still exists.
6. Deny approval.
7. Assert audit log records denied command and file still exists.

### 3.5b Command Delete Approve Flow

1. Create `tmp.txt`.
2. Simulate `run_command` with `del tmp.txt` and valid reason.
3. Assert command is classified `destructive`.
4. Assert UI receives `approval.required`.
5. Approve once.
6. Assert command executes only after approval.
7. Assert audit log records approval, command, exit code and target path if detected.

### 3.6 Command Verify Allow Flow

1. Use temp Go workspace with tiny test.
2. Simulate `run_command` with `go test ./...`.
3. Assert risk is `verify`.
4. Approve once.
5. Assert command runs with cwd inside workspace and timeout.
6. Assert audit log records exit code.

## 4. Manual Windows 7 Smoke Tests

- App starts without admin permission.
- Window shows icon/taskbar entry.
- Chat input works with Vietnamese text.
- LLM streaming updates UI without freezing.
- Selecting workspace via native dialog works.
- Read/list/search tools work on NTFS paths with spaces.
- Delete approval dialog is readable and cannot be bypassed by pressing Enter accidentally.
- Command approval dialog shows command, cwd, risk and reason.
- `del`, `rd`, `Remove-Item`, `git clean -fd` are blocked in normal mode.
- App exits cleanly and log file is flushed.

## 5. Release Gate

Release build must pass:

- Unit tests for `internal/safety`.
- Integration tests for read/delete/write flows.
- Manual smoke on one Windows 7 VM or physical machine.
- Manual smoke on Windows 10/11 if supported.
- Virus scanner false-positive check for packed exe.
