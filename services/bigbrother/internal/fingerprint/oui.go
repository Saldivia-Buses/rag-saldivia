// Package fingerprint provides device classification and identification.
package fingerprint

import (
	"net"
	"strings"
)

// Common MAC vendor prefixes (first 3 bytes = OUI).
// This is a curated subset for fast offline lookup. For comprehensive
// lookup, use github.com/klauspost/oui with the full IEEE database.
var knownVendors = map[string]string{
	"00:1a:2b": "Siemens",
	"00:0c:29": "VMware",
	"00:50:56": "VMware",
	"00:15:5d": "Microsoft Hyper-V",
	"08:00:27": "VirtualBox",
	"b8:27:eb": "Raspberry Pi",
	"dc:a6:32": "Raspberry Pi",
	"e4:5f:01": "Raspberry Pi",
	"00:1e:06": "WAGO",
	"00:30:de": "WAGO",
	"00:80:f4": "Schneider Electric",
	"00:0e:8c": "Siemens AG",
	"00:60:65": "Honeywell",
	"00:20:4a": "Proware",
	"00:01:05": "Beckhoff",
	"00:02:a5": "HP",
	"3c:d9:2b": "HP",
	"00:25:b3": "HP",
	"00:1a:a0": "Dell",
	"f8:bc:12": "Dell",
	"00:0c:76": "Micro-Star",
	"00:1b:21": "Intel",
	"00:1e:67": "Intel",
	"f4:4d:30": "Intel",
	"00:e0:4c": "Realtek",
	"d8:cb:8a": "Micro-Star",
	"00:11:32": "Synology",
	"48:21:0b": "TP-Link",
	"e8:65:49": "Cisco",
	"00:1b:17": "Hikvision",
	"c0:56:e3": "Hikvision",
	"78:a5:04": "Texas Instruments",
	"b0:b2:1c": "Ubiquiti",
	"fc:ec:da": "Ubiquiti",
	"24:5a:4c": "Ubiquiti",
	"00:26:2d": "MikroTik",
	"e4:8d:8c": "MikroTik",
	"00:17:88": "Philips Hue",
	"30:b5:c2": "TP-Link",
	"14:cc:20": "TP-Link",
	"00:23:24": "Zebra",
	"ac:3f:a4": "Espressif (ESP32)",
	"24:6f:28": "Espressif (ESP32)",
}

// LookupVendor returns the vendor name for a MAC address based on OUI prefix.
// Returns empty string if vendor is unknown.
func LookupVendor(mac net.HardwareAddr) string {
	if len(mac) < 3 {
		return ""
	}
	prefix := strings.ToLower(mac[:3].String())
	return knownVendors[prefix]
}

// EnrichVendor sets the Vendor field on devices that don't already have one.
func EnrichVendor(devices []struct {
	MAC    net.HardwareAddr
	Vendor *string
}) {
	for i := range devices {
		if devices[i].Vendor != nil && *devices[i].Vendor != "" {
			continue
		}
		if v := LookupVendor(devices[i].MAC); v != "" {
			devices[i].Vendor = &v
		}
	}
}
