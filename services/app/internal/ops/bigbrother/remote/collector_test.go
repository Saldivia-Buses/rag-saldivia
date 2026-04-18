package remote

import (
	"strings"
	"testing"
)

func TestIsValidCommand(t *testing.T) {
	tests := []struct {
		cmd  CommandType
		want bool
	}{
		{CmdSystemInfo, true},
		{CmdDiskUsage, true},
		{CmdMemory, true},
		{CmdProcesses, true},
		{CmdSoftware, true},
		{CmdNetwork, true},
		{CmdUptime, true},
		{"rm -rf /", false},
		{"", false},
		{"custom_command", false},
	}

	for _, tt := range tests {
		if got := IsValidCommand(tt.cmd); got != tt.want {
			t.Errorf("IsValidCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestResolveCommand(t *testing.T) {
	linux, ok := ResolveCommand(CmdSystemInfo, false)
	if !ok || linux != "uname -a" {
		t.Errorf("Linux CmdSystemInfo = %q, ok=%v", linux, ok)
	}

	win, ok := ResolveCommand(CmdSystemInfo, true)
	if !ok || win != "systeminfo" {
		t.Errorf("Windows CmdSystemInfo = %q, ok=%v", win, ok)
	}

	_, ok = ResolveCommand("invalid", false)
	if ok {
		t.Error("expected false for invalid command")
	}
}

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(string) bool
	}{
		{
			"normal output",
			"Linux server 5.15.0 #1 SMP x86_64",
			func(s string) bool { return s == "Linux server 5.15.0 #1 SMP x86_64" },
		},
		{
			"strip password",
			"DB_PASSWORD=secret123\nother data",
			func(s string) bool { return strings.Contains(s, "[REDACTED]") && !strings.Contains(s, "secret123") },
		},
		{
			"strip api key",
			"api_key: sk-1234abcd",
			func(s string) bool { return strings.Contains(s, "[REDACTED]") && !strings.Contains(s, "sk-1234") },
		},
		{
			"truncate large output",
			strings.Repeat("x", 10000),
			func(s string) bool { return len(s) < 5000 && strings.Contains(s, "[truncated to 4KB]") },
		},
		{
			"trim trailing whitespace",
			"data\n\n\n",
			func(s string) bool { return s == "data" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeOutput(tt.input)
			if !tt.check(got) {
				t.Errorf("SanitizeOutput failed: got %q", got)
			}
		})
	}
}

func TestIsAllowedSFTPPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/etc/os-release", true},
		{"/etc/hostname", true},
		{"/etc/hosts", true},
		{"/etc/shadow", false},
		{"/etc/passwd", false},
		{"/root/.ssh/id_rsa", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsAllowedSFTPPath(tt.path); got != tt.want {
			t.Errorf("IsAllowedSFTPPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestValidCommands(t *testing.T) {
	cmds := ValidCommands()
	if len(cmds) != 7 {
		t.Errorf("expected 7 commands, got %d", len(cmds))
	}
}
