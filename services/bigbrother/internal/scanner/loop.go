package scanner

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// ScanMode determines how aggressively the scanner operates.
type ScanMode string

const (
	// ModePassive monitors ARP table only — no packets sent. Zero IDS risk.
	ModePassive ScanMode = "passive"
	// ModeActive does ARP sweep every 30 min + fingerprint new devices. Low IDS risk.
	ModeActive ScanMode = "active"
	// ModeFull does ARP sweep every 5 min + full nmap + SNMP walk. High IDS risk.
	ModeFull ScanMode = "full"
)

// Loop orchestrates periodic network discovery scans.
type Loop struct {
	scanner   NetworkScanner
	mode      atomic.Value // ScanMode
	onResult  func(ctx context.Context, devices []Device) // callback to persist results
	modeCh    chan ScanMode                                // runtime mode changes
	alive     atomic.Bool
	stopCh    chan struct{}
}

// NewLoop creates a discovery loop with the given scanner and result callback.
func NewLoop(scanner NetworkScanner, mode ScanMode, onResult func(ctx context.Context, devices []Device)) *Loop {
	l := &Loop{
		scanner:  scanner,
		onResult: onResult,
		modeCh:   make(chan ScanMode, 1),
		stopCh:   make(chan struct{}),
	}
	l.mode.Store(mode)
	return l
}

// Start begins the discovery loop in a goroutine. Returns immediately.
func (l *Loop) Start(ctx context.Context) {
	l.alive.Store(true)
	go l.run(ctx)
}

// Stop signals the loop to stop.
func (l *Loop) Stop() {
	close(l.stopCh)
}

// IsAlive returns true if the scanner goroutine is running.
// Used by health checks to detect dead goroutines.
func (l *Loop) IsAlive() bool {
	return l.alive.Load()
}

// SetMode changes the scan mode at runtime (without restart).
func (l *Loop) SetMode(mode ScanMode) {
	select {
	case l.modeCh <- mode:
	default:
		// Channel full — previous mode change not yet processed
	}
}

// Mode returns the current scan mode.
func (l *Loop) Mode() ScanMode {
	return l.mode.Load().(ScanMode)
}

func (l *Loop) run(ctx context.Context) {
	defer l.alive.Store(false)

	// Run initial scan immediately
	l.doScan(ctx)

	for {
		interval := l.interval()

		select {
		case <-ctx.Done():
			slog.Info("scanner loop stopping: context cancelled")
			return
		case <-l.stopCh:
			slog.Info("scanner loop stopping: stop signal")
			return
		case newMode := <-l.modeCh:
			slog.Info("scanner mode changed", "from", l.Mode(), "to", newMode)
			l.mode.Store(newMode)
			// Re-scan immediately with new mode
			l.doScan(ctx)
		case <-time.After(interval):
			l.doScan(ctx)
		}
	}
}

func (l *Loop) doScan(ctx context.Context) {
	mode := l.Mode()
	if mode == ModePassive {
		// Passive mode: no active scanning, just log
		slog.Debug("scanner in passive mode, skipping active scan")
		return
	}

	start := time.Now()
	slog.Info("scan starting", "mode", mode)

	devices, err := l.scanner.Scan(ctx)
	if err != nil {
		slog.Error("scan failed", "error", err, "mode", mode)
		return
	}

	duration := time.Since(start)
	slog.Info("scan completed", "mode", mode, "found", len(devices), "duration", duration)

	if l.onResult != nil {
		l.onResult(ctx, devices)
	}
}

func (l *Loop) interval() time.Duration {
	switch l.Mode() {
	case ModeFull:
		return 5 * time.Minute
	case ModeActive:
		return 30 * time.Minute
	default: // passive
		return 5 * time.Minute // check for mode changes
	}
}
