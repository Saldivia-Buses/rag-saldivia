package fingerprint

import (
	"testing"

	"github.com/Camionerou/rag-saldivia/services/app/internal/ops/bigbrother/scanner"
)

func TestClassifyDevice(t *testing.T) {
	tests := []struct {
		name      string
		ports     []scanner.PortInfo
		vendor    string
		os        string
		snmpDescr string
		want      DeviceType
	}{
		{
			name:  "modbus PLC",
			ports: []scanner.PortInfo{{Port: 502, Protocol: "tcp", State: "open"}},
			want:  TypePLC,
		},
		{
			name:  "opcua PLC",
			ports: []scanner.PortInfo{{Port: 4840, Protocol: "tcp", State: "open"}},
			want:  TypePLC,
		},
		{
			name:   "siemens vendor",
			ports:  nil,
			vendor: "Siemens AG",
			want:   TypePLC,
		},
		{
			name:   "wago vendor",
			ports:  nil,
			vendor: "WAGO",
			want:   TypePLC,
		},
		{
			name:  "printer IPP port",
			ports: []scanner.PortInfo{{Port: 631, Protocol: "tcp", State: "open"}},
			want:  TypePrinter,
		},
		{
			name:  "printer raw port 9100",
			ports: []scanner.PortInfo{{Port: 9100, Protocol: "tcp", State: "open"}},
			want:  TypePrinter,
		},
		{
			name:      "printer SNMP description",
			ports:     nil,
			snmpDescr: "HP LaserJet Pro MFP",
			want:      TypePrinter,
		},
		{
			name:  "camera RTSP",
			ports: []scanner.PortInfo{{Port: 554, Protocol: "tcp", State: "open"}},
			want:  TypeCamera,
		},
		{
			name:   "hikvision camera",
			ports:  nil,
			vendor: "Hikvision",
			want:   TypeCamera,
		},
		{
			name:      "cisco switch",
			vendor:    "Cisco",
			snmpDescr: "Cisco IOS Software, C2960",
			want:      TypeSwitch,
		},
		{
			name:      "mikrotik router",
			vendor:    "MikroTik",
			snmpDescr: "RouterOS 7.1",
			want:      TypeSwitch,
		},
		{
			name:   "ubiquiti AP",
			vendor: "Ubiquiti",
			want:   TypeAP,
		},
		{
			name: "linux server with many ports",
			ports: []scanner.PortInfo{
				{Port: 22, Protocol: "tcp", State: "open"},
				{Port: 80, Protocol: "tcp", State: "open"},
				{Port: 443, Protocol: "tcp", State: "open"},
				{Port: 5432, Protocol: "tcp", State: "open"},
			},
			want: TypeServer,
		},
		{
			name: "windows server OS",
			os:   "Windows Server 2022",
			want: TypeServer,
		},
		{
			name: "windows 10 workstation",
			os:   "Windows 10 Pro",
			want: TypeWorkstation,
		},
		{
			name:  "ssh only — workstation",
			ports: []scanner.PortInfo{{Port: 22, Protocol: "tcp", State: "open"}},
			want:  TypeWorkstation,
		},
		{
			name:   "esp32 IoT",
			vendor: "Espressif (ESP32)",
			want:   TypeIoT,
		},
		{
			name:   "raspberry pi IoT",
			vendor: "Raspberry Pi",
			want:   TypeIoT,
		},
		{
			name:   "apple phone",
			vendor: "Apple",
			want:   TypePhone,
		},
		{
			name: "unknown device",
			want: TypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyDevice(tt.ports, tt.vendor, tt.os, tt.snmpDescr)
			if got != tt.want {
				t.Errorf("ClassifyDevice() = %q, want %q", got, tt.want)
			}
		})
	}
}
