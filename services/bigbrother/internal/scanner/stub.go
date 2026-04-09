package scanner

import (
	"context"
	"net"
)

// StubScanner returns a fixed set of devices for development/testing.
// Use when running in WSL2 or environments without physical NIC.
type StubScanner struct {
	Devices []Device
}

// NewStubScanner creates a stub scanner with sample devices.
func NewStubScanner() *StubScanner {
	return &StubScanner{
		Devices: []Device{
			{IP: net.ParseIP("192.168.1.1"), MAC: mustParseMAC("aa:bb:cc:dd:ee:01"), Hostname: "gateway"},
			{IP: net.ParseIP("192.168.1.10"), MAC: mustParseMAC("aa:bb:cc:dd:ee:10"), Hostname: "plc-line1"},
			{IP: net.ParseIP("192.168.1.20"), MAC: mustParseMAC("aa:bb:cc:dd:ee:20"), Hostname: "server-main"},
			{IP: net.ParseIP("192.168.1.30"), MAC: mustParseMAC("aa:bb:cc:dd:ee:30"), Hostname: "printer-office"},
			{IP: net.ParseIP("192.168.1.40"), MAC: mustParseMAC("aa:bb:cc:dd:ee:40"), Hostname: "camera-entrada"},
		},
	}
}

func (s *StubScanner) Scan(_ context.Context) ([]Device, error) {
	return s.Devices, nil
}

func mustParseMAC(s string) net.HardwareAddr {
	mac, err := net.ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return mac
}
