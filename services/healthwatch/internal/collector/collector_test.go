package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKnownService(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    bool
	}{
		{"valid service", "auth", true},
		{"valid service ws", "ws", true},
		{"valid service healthwatch", "healthwatch", true},
		{"unknown service", "evil-service", false},
		{"empty string", "", false},
		{"sql injection attempt", "auth'; DROP TABLE users;--", false},
		{"promql injection", `auth"}[5m]) or vector(1)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsKnownService(tt.service))
		})
	}
}

func TestServicePortMap_AllKnownServicesHavePorts(t *testing.T) {
	for _, svc := range KnownServices {
		port, ok := ServicePortMap[svc]
		assert.True(t, ok, "service %s missing from port map", svc)
		assert.NotEmpty(t, port, "service %s has empty port", svc)
	}
}
