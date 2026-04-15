---
title: Package: pkg/remote
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./plc.md
  - ./audit.md
---

## Purpose

SSH and WinRM clients for remote command execution and file transfer on
network devices, plus a fixed `CommandType` allowlist that maps to OS-specific
command strings. **WARNING:** raw transport clients — production use MUST go
through a service that enforces command allowlists, audit logging, and
credential management (BigBrother is the canonical caller). Import this only
when building such a service.

## Public API

| Symbol | Source | Description |
|--------|--------|-------------|
| `SSHConfig` | `ssh.go:33` | Host, Port (22), User, Password, KeyBytes, Timeout, KnownHostsFile |
| `SSHClient` | `ssh.go:27` | Wraps `ssh.Client`; `Exec`, `ReadFile`, `Close`, `Addr` |
| `NewSSHClient(cfg)` | `ssh.go:137` | Connects with TOFU host-key verification |
| `IsReachable(host, port, timeout)` | `ssh.go:280` | TCP probe without auth |
| `WinRMConfig` | `winrm.go:21` | Host, Port (5986 HTTPS), User, Password, Timeout, CACertFile, Insecure |
| `WinRMClient` | `winrm.go:15` | `Exec`, `Addr` |
| `NewWinRMClient(cfg)` | `winrm.go:39` | Always HTTPS; loads CA cert if provided |
| `ExecResult` | `ssh.go:189` | `Stdout`, `Stderr`, `ExitCode` |
| `CommandType` | `collector.go:14` | Enum: `system_info`, `disk_usage`, `memory`, `processes`, `software`, `network`, `uptime` |
| `CommandMap` | `collector.go:35` | `CommandType → {Linux, Windows}` strings |
| `ValidCommands()` / `IsValidCommand(cmd)` | `collector.go:46` | Allowlist accessors |
| `ResolveCommand(cmd, isWindows)` | `collector.go:62` | Lookup OS-specific string |
| `SanitizeOutput(out)` | `collector.go:77` | Strips secret patterns, truncates to 4KB |
| `AllowedSFTPPaths` | `collector.go:94` | Whitelist for `ReadFile` |
| `IsAllowedSFTPPath(path)` | `collector.go:101` | Lookup |
| `MaxOutputSize` | `collector.go:9` | 4096 bytes |

## Usage

```go
cmdStr, ok := remote.ResolveCommand(remote.CmdSystemInfo, isWindows)
if !ok { return remote.ErrUnknownCommand }

c, _ := remote.NewSSHClient(remote.SSHConfig{
    Host: host, User: u, KeyBytes: pem,
})
defer c.Close()
res, err := c.Exec(ctx, cmdStr)
clean := remote.SanitizeOutput(res.Stdout)
```

## Invariants

- SSH uses TOFU (Trust On First Use) host-key verification with a dedicated
  `~/.ssh/sda_known_hosts` file (`pkg/remote/ssh.go:48`). A changed host key
  returns a "MITM detected" error. The literal `"-"` path bypasses verification
  for tests only and emits a `slog.Warn`.
- WinRM is HTTPS-only on port 5986 — `useHTTPS` is hard-coded `true`
  (`pkg/remote/winrm.go:69`).
- `SSHClient.ReadFile` rejects symlinks (`pkg/remote/ssh.go:251`) — defends
  against compromised targets exfiltrating arbitrary files.
- `Exec` on either client honours `ctx`. SSH sends `SIGKILL` on cancel
  (`pkg/remote/ssh.go:217`).
- The `CommandType` allowlist is the only sanctioned way to run remote commands.
  Custom command strings are rejected at the service layer.
- `SanitizeOutput` redacts patterns matching `(password|passwd|secret|key|token|credential|api.?key)\s*[=:]\s*\S+`
  (`pkg/remote/collector.go:74`).

## Importers

`services/bigbrother/internal/handler/devices.go`,
`services/bigbrother/internal/service/remote.go` — sole production callers.
