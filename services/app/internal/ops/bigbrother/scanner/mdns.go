package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/mdns"
)

// MDNSScanner discovers services advertised via mDNS/DNS-SD on the local network.
type MDNSScanner struct {
	timeout time.Duration
}

// MDNSService represents a discovered mDNS service.
type MDNSService struct {
	Name string // service instance name
	Host string // hostname
	IP   string // IPv4 address
	Port int    // service port
	Info string // TXT record info
}

// Common mDNS service types to scan for.
var defaultServiceTypes = []string{
	"_http._tcp",
	"_https._tcp",
	"_ssh._tcp",
	"_printer._tcp",
	"_ipp._tcp",
	"_smb._tcp",
	"_workstation._tcp",
	"_device-info._tcp",
	"_hap._tcp",        // HomeKit
	"_googlecast._tcp",
	"_airplay._tcp",
}

// NewMDNSScanner creates a scanner with configurable timeout.
func NewMDNSScanner(timeout time.Duration) *MDNSScanner {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &MDNSScanner{timeout: timeout}
}

// Discover scans for mDNS services on the local network.
// Returns all discovered services across all service types.
func (s *MDNSScanner) Discover(ctx context.Context) ([]MDNSService, error) {
	var allServices []MDNSService

	for _, serviceType := range defaultServiceTypes {
		if ctx.Err() != nil {
			break
		}

		services, err := s.discoverService(ctx, serviceType)
		if err != nil {
			slog.Debug("mdns discovery failed for service type", "type", serviceType, "error", err)
			continue
		}
		allServices = append(allServices, services...)
	}

	slog.Info("mdns discovery completed", "services_found", len(allServices))
	return allServices, nil
}

func (s *MDNSScanner) discoverService(ctx context.Context, serviceType string) ([]MDNSService, error) {
	entriesCh := make(chan *mdns.ServiceEntry, 32)

	var services []MDNSService
	done := make(chan struct{})

	go func() {
		defer close(done)
		for entry := range entriesCh {
			ip := ""
			if entry.AddrV4 != nil {
				ip = entry.AddrV4.String()
			}
			if ip == "" {
				continue // skip IPv6-only entries
			}
			services = append(services, MDNSService{
				Name: entry.Name,
				Host: entry.Host,
				IP:   ip,
				Port: entry.Port,
				Info: fmt.Sprintf("%v", entry.InfoFields),
			})
		}
	}()

	params := &mdns.QueryParam{
		Service:             serviceType,
		Domain:              "local",
		Timeout:             s.timeout,
		Entries:             entriesCh,
		WantUnicastResponse: false,
	}

	err := mdns.Query(params)
	close(entriesCh)
	<-done

	if err != nil {
		return nil, err
	}
	return services, nil
}
