package fingerprint

import (
	"strings"

	"github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/scanner"
)

// DeviceType is the classified type of a network device.
type DeviceType string

const (
	TypePLC         DeviceType = "plc"
	TypeWorkstation DeviceType = "workstation"
	TypeServer      DeviceType = "server"
	TypeSwitch      DeviceType = "switch"
	TypePrinter     DeviceType = "printer"
	TypeCamera      DeviceType = "camera"
	TypeAP          DeviceType = "ap"
	TypePhone       DeviceType = "phone"
	TypeIoT         DeviceType = "iot"
	TypeUnknown     DeviceType = "unknown"
)

// ClassifyDevice determines the device type based on open ports, vendor, OS, and SNMP data.
// Uses deterministic rules — no ML or heuristics.
func ClassifyDevice(ports []scanner.PortInfo, vendor string, os string, snmpDescr string) DeviceType {
	vendorLower := strings.ToLower(vendor)
	osLower := strings.ToLower(os)
	descrLower := strings.ToLower(snmpDescr)

	// PLC detection: Modbus (502) or OPC-UA (4840) or PLC vendor
	if scanner.DetectModbus(ports) || scanner.DetectOPCUA(ports) {
		return TypePLC
	}
	if isPLCVendor(vendorLower) {
		return TypePLC
	}

	// Printer detection: IPP (631) or printer vendor or SNMP description
	if scanner.HasOpenPort(ports, 631) || scanner.HasOpenPort(ports, 9100) {
		return TypePrinter
	}
	if containsAny(vendorLower, "zebra", "brother", "epson", "canon", "lexmark", "ricoh") {
		return TypePrinter
	}
	if containsAny(descrLower, "printer", "laserjet", "deskjet") {
		return TypePrinter
	}

	// Camera detection: RTSP (554) or camera vendor
	if scanner.HasOpenPort(ports, 554) {
		return TypeCamera
	}
	if containsAny(vendorLower, "hikvision", "dahua", "axis", "reolink", "amcrest") {
		return TypeCamera
	}

	// Network equipment: SNMP + switch/router keywords
	if containsAny(descrLower, "switch", "cisco ios", "routeros", "junos") {
		return TypeSwitch
	}
	if containsAny(vendorLower, "cisco", "mikrotik", "juniper", "arista", "netgear") {
		// Could be switch or AP — check further
		if containsAny(descrLower, "access point", "wireless", "wifi") {
			return TypeAP
		}
		return TypeSwitch
	}

	// AP detection
	if containsAny(vendorLower, "ubiquiti", "unifi", "aruba", "ruckus") {
		return TypeAP
	}

	// Server detection: common server ports or server OS
	if isServerLikely(ports, osLower, descrLower) {
		return TypeServer
	}

	// Workstation: Windows or desktop Linux with SSH
	if containsAny(osLower, "windows 10", "windows 11", "windows 7", "ubuntu desktop", "fedora") {
		return TypeWorkstation
	}
	if scanner.DetectSSH(ports) && !isServerLikely(ports, osLower, descrLower) {
		// SSH but no server indicators — likely a workstation
		return TypeWorkstation
	}

	// IoT: ESP32, Raspberry Pi, or small embedded vendors
	if containsAny(vendorLower, "espressif", "esp32", "raspberry") {
		return TypeIoT
	}

	// Phone detection
	if containsAny(vendorLower, "apple", "samsung", "huawei", "xiaomi", "oneplus") {
		return TypePhone
	}

	return TypeUnknown
}

func isPLCVendor(vendor string) bool {
	return containsAny(vendor,
		"siemens", "wago", "schneider", "beckhoff", "allen-bradley",
		"rockwell", "omron", "mitsubishi", "abb", "honeywell",
		"phoenix contact", "pilz", "festo",
	)
}

func isServerLikely(ports []scanner.PortInfo, os, descr string) bool {
	serverPorts := 0
	for _, p := range ports {
		if p.State != "open" {
			continue
		}
		switch p.Port {
		case 22, 80, 443, 3306, 5432, 6379, 8080, 8443, 9090, 27017:
			serverPorts++
		}
	}
	if serverPorts >= 3 {
		return true
	}
	if containsAny(os, "ubuntu server", "centos", "rhel", "debian", "windows server") {
		return true
	}
	if containsAny(descr, "server", "linux", "freebsd") {
		return true
	}
	return false
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
