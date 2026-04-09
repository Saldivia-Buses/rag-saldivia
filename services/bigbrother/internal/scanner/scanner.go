// Package scanner provides network discovery for BigBrother.
// Discovers devices on the LAN via ARP sweep, then enriches with
// fingerprinting (nmap, SNMP, mDNS) in later phases.
package scanner

import (
	"context"
	"net"
)

// Device represents a discovered network device before persistence.
type Device struct {
	IP       net.IP
	MAC      net.HardwareAddr
	Hostname string
	Vendor   string // from OUI lookup
}

// NetworkScanner is the interface for network discovery implementations.
// In production, ARPScanner does real ARP sweeps. In development (WSL2),
// use a mock/stub since there's no physical NIC.
type NetworkScanner interface {
	// Scan discovers devices on the network. Returns all found devices.
	Scan(ctx context.Context) ([]Device, error)
}
