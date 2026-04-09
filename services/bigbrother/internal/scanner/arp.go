package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/mdlayher/arp"
)

// ARPScanner discovers devices via Layer 2 ARP sweep.
// Requires CAP_NET_RAW capability. Will not work in WSL2 (no physical NIC).
type ARPScanner struct {
	iface   *net.Interface
	timeout time.Duration
}

// NewARPScanner creates a scanner for the given network interface.
func NewARPScanner(ifaceName string, timeout time.Duration) (*ARPScanner, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("interface %q: %w", ifaceName, err)
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &ARPScanner{iface: iface, timeout: timeout}, nil
}

// Scan performs an ARP sweep on the interface's subnet.
// Sends ARP requests to all IPs in the subnet and collects responses.
func (s *ARPScanner) Scan(ctx context.Context) ([]Device, error) {
	client, err := arp.Dial(s.iface)
	if err != nil {
		return nil, fmt.Errorf("arp dial on %s: %w", s.iface.Name, err)
	}
	defer client.Close()

	addrs, err := s.iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("get addrs for %s: %w", s.iface.Name, err)
	}

	var subnet *net.IPNet
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			subnet = ipnet
			break
		}
	}
	if subnet == nil {
		return nil, fmt.Errorf("no IPv4 subnet found on %s", s.iface.Name)
	}

	// Enumerate all IPs in the subnet
	ips := enumerateSubnet(subnet)
	slog.Info("ARP scan starting", "interface", s.iface.Name, "subnet", subnet.String(), "hosts", len(ips))

	// Set deadline for the entire scan
	if err := client.SetDeadline(time.Now().Add(s.timeout)); err != nil {
		return nil, fmt.Errorf("set deadline: %w", err)
	}

	// Send ARP requests to all IPs
	for _, ip := range ips {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		addr, ok := netip.AddrFromSlice(ip.To4())
		if !ok {
			continue
		}
		if err := client.Request(addr); err != nil {
			// Non-fatal: some IPs may be unreachable
			continue
		}
	}

	// Collect responses
	var devices []Device
	seen := make(map[string]bool)

	for {
		if ctx.Err() != nil {
			break
		}

		pkt, _, err := client.Read()
		if err != nil {
			// Timeout or other read error — scan is done
			break
		}

		if pkt.Operation != arp.OperationReply {
			continue
		}

		key := pkt.SenderHardwareAddr.String()
		if seen[key] {
			continue
		}
		seen[key] = true

		devices = append(devices, Device{
			IP:  net.IP(pkt.SenderIP.AsSlice()),
			MAC: pkt.SenderHardwareAddr,
		})
	}

	slog.Info("ARP scan completed", "found", len(devices))
	return devices, nil
}

// enumerateSubnet returns all host IPs in a subnet (excluding network and broadcast).
func enumerateSubnet(subnet *net.IPNet) []net.IP {
	var ips []net.IP
	ip := subnet.IP.Mask(subnet.Mask)

	for ip := cloneIP(ip); subnet.Contains(ip); incIP(ip) {
		// Skip network address (all host bits 0) and broadcast (all host bits 1)
		if isNetworkOrBroadcast(ip, subnet) {
			continue
		}
		ips = append(ips, cloneIP(ip))
	}

	// Cap at 1024 hosts to prevent scanning huge subnets
	if len(ips) > 1024 {
		ips = ips[:1024]
	}

	return ips
}

func cloneIP(ip net.IP) net.IP {
	c := make(net.IP, len(ip))
	copy(c, ip)
	return c
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isNetworkOrBroadcast(ip net.IP, subnet *net.IPNet) bool {
	// Network address: all host bits are 0
	// Broadcast address: all host bits are 1
	for i := range ip {
		hostBits := ip[i] & ^subnet.Mask[i]
		if hostBits != 0 && hostBits != ^subnet.Mask[i] {
			return false
		}
	}
	return true
}
