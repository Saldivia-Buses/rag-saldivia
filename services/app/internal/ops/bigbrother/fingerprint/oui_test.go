package fingerprint

import (
	"net"
	"testing"
)

func TestLookupVendor(t *testing.T) {
	tests := []struct {
		mac  string
		want string
	}{
		{"00:0c:29:ab:cd:ef", "VMware"},
		{"b8:27:eb:11:22:33", "Raspberry Pi"},
		{"00:1e:06:aa:bb:cc", "WAGO"},
		{"00:80:f4:11:22:33", "Schneider Electric"},
		{"00:01:05:11:22:33", "Beckhoff"},
		{"ff:ff:ff:ff:ff:ff", ""},
		{"00:00:00:00:00:00", ""},
	}

	for _, tt := range tests {
		mac, err := net.ParseMAC(tt.mac)
		if err != nil {
			t.Fatalf("parse MAC %s: %v", tt.mac, err)
		}
		got := LookupVendor(mac)
		if got != tt.want {
			t.Errorf("LookupVendor(%s) = %q, want %q", tt.mac, got, tt.want)
		}
	}
}

func TestLookupVendorShortMAC(t *testing.T) {
	got := LookupVendor(net.HardwareAddr{0x00, 0x0c})
	if got != "" {
		t.Errorf("expected empty for short MAC, got %q", got)
	}
}
