package remote

import (
	"regexp"
	"strings"
)

const (
	// MaxOutputSize is the maximum size of command output (4KB).
	// Larger outputs are truncated.
	MaxOutputSize = 4096
)

// CommandType defines the allowed enum-based commands.
// Only these commands can be executed — no custom strings.
type CommandType string

const (
	CmdSystemInfo CommandType = "system_info"
	CmdDiskUsage  CommandType = "disk_usage"
	CmdMemory     CommandType = "memory"
	CmdProcesses  CommandType = "processes"
	CmdSoftware   CommandType = "software"
	CmdNetwork    CommandType = "network"
	CmdUptime     CommandType = "uptime"
)

type commandPair struct {
	Linux   string
	Windows string
}

// CommandMap maps enum values to fixed OS-specific commands.
// These are the ONLY commands that can be executed remotely.
var CommandMap = map[CommandType]commandPair{
	CmdSystemInfo: {"uname -a", "systeminfo"},
	CmdDiskUsage:  {"df -h", "wmic logicaldisk get size,freespace,caption"},
	CmdMemory:     {"free -m", "wmic memorychip get capacity"},
	CmdProcesses:  {"ps aux --sort=-%mem | head -20", "tasklist /FO TABLE"},
	CmdSoftware:   {"dpkg -l", "wmic product get name,version"},
	CmdNetwork:    {"ip addr", "ipconfig /all"},
	CmdUptime:     {"uptime", "systeminfo | findstr Boot"},
}

// ValidCommands returns all valid command type values.
func ValidCommands() []CommandType {
	cmds := make([]CommandType, 0, len(CommandMap))
	for k := range CommandMap {
		cmds = append(cmds, k)
	}
	return cmds
}

// IsValidCommand returns true if the command type is recognized.
func IsValidCommand(cmd CommandType) bool {
	_, ok := CommandMap[cmd]
	return ok
}

// ResolveCommand returns the OS-specific command string for the given type.
// isWindows determines which variant to return.
func ResolveCommand(cmd CommandType, isWindows bool) (string, bool) {
	pair, ok := CommandMap[cmd]
	if !ok {
		return "", false
	}
	if isWindows {
		return pair.Windows, true
	}
	return pair.Linux, true
}

// sensitivePatterns are stripped from command output to prevent leaking secrets.
var sensitivePatterns = regexp.MustCompile(`(?i)(password|passwd|secret|key|token|credential|api.?key)\s*[=:]\s*\S+`)

// SanitizeOutput truncates output to MaxOutputSize and strips sensitive patterns.
func SanitizeOutput(output string) string {
	// Strip sensitive patterns
	output = sensitivePatterns.ReplaceAllString(output, "[REDACTED]")

	// Truncate to max size
	if len(output) > MaxOutputSize {
		output = output[:MaxOutputSize] + "\n... [truncated to 4KB]"
	}

	// Trim trailing whitespace
	output = strings.TrimRight(output, "\n\r\t ")

	return output
}

// AllowedSFTPPaths defines the only paths that can be read via SFTP.
// Any path not in this map is rejected at the service layer.
var AllowedSFTPPaths = map[string]bool{
	"/etc/os-release": true,
	"/etc/hostname":   true,
	"/etc/hosts":      true,
}

// IsAllowedSFTPPath returns true if the path is in the SFTP allowlist.
func IsAllowedSFTPPath(path string) bool {
	return AllowedSFTPPaths[path]
}
