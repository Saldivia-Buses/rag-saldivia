package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMPScanner queries managed devices via SNMP for system information.
type SNMPScanner struct {
	timeout   time.Duration
	community string // v2c community string (default "public")
}

// SNMPResult holds system info retrieved via SNMP.
type SNMPResult struct {
	SysDescr  string // .1.3.6.1.2.1.1.1.0 — system description
	SysName   string // .1.3.6.1.2.1.1.5.0 — hostname
	SysUptime uint32 // .1.3.6.1.2.1.1.3.0 — uptime in timeticks
}

// Well-known SNMP OIDs
const (
	oidSysDescr  = ".1.3.6.1.2.1.1.1.0"
	oidSysName   = ".1.3.6.1.2.1.1.5.0"
	oidSysUptime = ".1.3.6.1.2.1.1.3.0"
)

// NewSNMPScanner creates an SNMP scanner with v2c defaults.
func NewSNMPScanner(community string, timeout time.Duration) *SNMPScanner {
	if community == "" {
		community = "public"
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &SNMPScanner{timeout: timeout, community: community}
}

// Query retrieves system info from a single host via SNMP v2c GET.
// Returns nil result (not error) if host doesn't respond to SNMP.
func (s *SNMPScanner) Query(ctx context.Context, ip string) (*SNMPResult, error) {
	client := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      161,
		Community: s.community,
		Version:   gosnmp.Version2c,
		Timeout:   s.timeout,
		Retries:   1,
	}

	if err := client.ConnectIPv4(); err != nil {
		return nil, fmt.Errorf("snmp connect %s: %w", ip, err)
	}
	defer client.Conn.Close()

	oids := []string{oidSysDescr, oidSysName, oidSysUptime}
	result, err := client.Get(oids)
	if err != nil {
		// SNMP not available on this host — not an error
		slog.Debug("snmp query failed, host may not support SNMP", "ip", ip, "error", err)
		return nil, nil
	}

	r := &SNMPResult{}
	for _, v := range result.Variables {
		switch v.Name {
		case oidSysDescr:
			if s, ok := v.Value.([]byte); ok {
				r.SysDescr = string(s)
			} else if s, ok := v.Value.(string); ok {
				r.SysDescr = s
			}
		case oidSysName:
			if s, ok := v.Value.([]byte); ok {
				r.SysName = string(s)
			} else if s, ok := v.Value.(string); ok {
				r.SysName = s
			}
		case oidSysUptime:
			if u, ok := v.Value.(uint32); ok {
				r.SysUptime = u
			}
		}
	}

	slog.Debug("snmp query succeeded", "ip", ip, "sysName", r.SysName, "sysDescr", r.SysDescr)
	return r, nil
}

// BulkQuery queries multiple hosts for system info.
// Skips hosts that don't respond. Returns results only for responding hosts.
func (s *SNMPScanner) BulkQuery(ctx context.Context, ips []string) map[string]*SNMPResult {
	results := make(map[string]*SNMPResult)
	for _, ip := range ips {
		if ctx.Err() != nil {
			break
		}
		r, err := s.Query(ctx, ip)
		if err != nil {
			slog.Debug("snmp bulk query skip", "ip", ip, "error", err)
			continue
		}
		if r != nil {
			results[ip] = r
		}
	}
	return results
}
