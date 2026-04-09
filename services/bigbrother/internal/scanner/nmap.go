package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

// NmapScanner wraps the nmap binary for port scanning and OS fingerprinting.
// Requires `nmap` to be installed in the container (Alpine + nmap package).
type NmapScanner struct {
	timeout time.Duration
}

// NewNmapScanner creates a scanner with configurable timeout.
func NewNmapScanner(timeout time.Duration) *NmapScanner {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	return &NmapScanner{timeout: timeout}
}

// NmapResult holds the fingerprinting results for a single host.
type NmapResult struct {
	IP       string
	Hostname string
	OS       string
	Ports    []PortInfo
}

// PortInfo represents a discovered open port.
type PortInfo struct {
	Port     int
	Protocol string // "tcp" or "udp"
	Service  string // service name (e.g., "ssh", "http")
	Version  string // service version if detected
	State    string // "open", "closed", "filtered"
}

// ScanHosts runs nmap service + OS detection against the given IP addresses.
// Uses -sV (service version) -O (OS detection) -T4 (aggressive timing).
// Returns results for hosts that responded.
func (s *NmapScanner) ScanHosts(ctx context.Context, ips []string) ([]NmapResult, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets(ips...),
		nmap.WithServiceInfo(),
		nmap.WithOSDetection(),
		nmap.WithTimingTemplate(nmap.TimingAggressive),
		nmap.WithSkipHostDiscovery(), // we already know hosts are up from ARP
	)
	if err != nil {
		return nil, fmt.Errorf("create nmap scanner: %w", err)
	}

	slog.Info("nmap scan starting", "targets", len(ips))
	result, warnings, err := scanner.Run()
	if len(*warnings) > 0 {
		slog.Warn("nmap warnings", "warnings", *warnings)
	}
	if err != nil {
		return nil, fmt.Errorf("nmap scan: %w", err)
	}

	var results []NmapResult
	for _, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		r := NmapResult{
			IP: host.Addresses[0].Addr,
		}

		// Hostname
		if len(host.Hostnames) > 0 {
			r.Hostname = host.Hostnames[0].Name
		}

		// OS detection — pick the most likely match
		if len(host.OS.Matches) > 0 {
			r.OS = host.OS.Matches[0].Name
		}

		// Open ports
		for _, port := range host.Ports {
			r.Ports = append(r.Ports, PortInfo{
				Port:     int(port.ID),
				Protocol: port.Protocol,
				Service:  port.Service.Name,
				Version:  strings.TrimSpace(fmt.Sprintf("%s %s", port.Service.Product, port.Service.Version)),
				State:    port.State.State,
			})
		}

		results = append(results, r)
	}

	slog.Info("nmap scan completed", "scanned", len(ips), "responded", len(results))
	return results, nil
}

// ScanSingleHost is a convenience wrapper for scanning a single host.
func (s *NmapScanner) ScanSingleHost(ctx context.Context, ip string) (*NmapResult, error) {
	results, err := s.ScanHosts(ctx, []string{ip})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("host %s did not respond to nmap", ip)
	}
	return &results[0], nil
}

// DetectModbus checks if a host has Modbus TCP (port 502) open.
func DetectModbus(ports []PortInfo) bool {
	for _, p := range ports {
		if p.Port == 502 && p.State == "open" {
			return true
		}
	}
	return false
}

// DetectOPCUA checks if a host has OPC-UA (port 4840) open.
func DetectOPCUA(ports []PortInfo) bool {
	for _, p := range ports {
		if p.Port == 4840 && p.State == "open" {
			return true
		}
	}
	return false
}

// DetectSSH checks if a host has SSH (port 22) open.
func DetectSSH(ports []PortInfo) bool {
	for _, p := range ports {
		if p.Port == 22 && p.State == "open" {
			return true
		}
	}
	return false
}

// DetectWinRM checks if a host has WinRM HTTPS (port 5986) open.
func DetectWinRM(ports []PortInfo) bool {
	for _, p := range ports {
		if p.Port == 5986 && p.State == "open" {
			return true
		}
	}
	return false
}

// DetectSNMP checks if a host has SNMP (port 161) open.
func DetectSNMP(ports []PortInfo) bool {
	for _, p := range ports {
		if p.Port == 161 {
			return true
		}
	}
	return false
}

// DetectHTTP checks if a host has HTTP/HTTPS open.
func DetectHTTP(ports []PortInfo) bool {
	for _, p := range ports {
		if (p.Port == 80 || p.Port == 443 || p.Port == 8080 || p.Port == 8443) && p.State == "open" {
			return true
		}
	}
	return false
}

// HasOpenPort checks if a specific port is open on the host.
func HasOpenPort(ports []PortInfo, port int) bool {
	for _, p := range ports {
		if p.Port == port && p.State == "open" {
			return true
		}
	}
	return false
}

// FilterIP returns only IPv4 addresses from a mixed list.
func FilterIP(addrs []net.Addr) []string {
	var ips []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}
	return ips
}
