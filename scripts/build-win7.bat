@echo off
setlocal

where go >nul 2>nul
if errorlevel 1 (
  echo Go is not installed or not in PATH.
  exit /b 1
)

echo Building 32-bit Windows 7 binary. Use Go 1.20.x for real Win7 releases.
set GOOS=windows
set GOARCH=386
go build -o agentdesk-win7-386.exe .\cmd\agentdesk
if errorlevel 1 exit /b 1

echo Built agentdesk-win7-386.exe

