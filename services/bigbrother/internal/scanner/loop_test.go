package scanner

import (
	"context"
	"testing"
	"time"
)

// mockScanner returns empty results for every scan.
type mockScanner struct {
	scanCount int
}

func (m *mockScanner) Scan(_ context.Context) ([]Device, error) {
	m.scanCount++
	return nil, nil
}

func TestNewLoop_SetsMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode ScanMode
	}{
		{name: "passive", mode: ModePassive},
		{name: "active", mode: ModeActive},
		{name: "full", mode: ModeFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := NewLoop(&mockScanner{}, tt.mode, nil)
			if l.Mode() != tt.mode {
				t.Fatalf("expected mode %q, got %q", tt.mode, l.Mode())
			}
		})
	}
}

func TestMode_ReturnsCurrentMode(t *testing.T) {
	t.Parallel()
	l := NewLoop(&mockScanner{}, ModeActive, nil)

	if l.Mode() != ModeActive {
		t.Fatalf("expected %q, got %q", ModeActive, l.Mode())
	}
}

func TestSetMode_ChangesMode(t *testing.T) {
	t.Parallel()
	l := NewLoop(&mockScanner{}, ModePassive, nil)

	l.SetMode(ModeFull)

	// SetMode uses a buffered channel — the mode is stored asynchronously
	// by the run loop. But we can read the channel ourselves and store it
	// to verify the value was sent. Since the loop isn't running, we drain
	// the channel and apply it manually.
	select {
	case newMode := <-l.modeCh:
		if newMode != ModeFull {
			t.Fatalf("expected %q on channel, got %q", ModeFull, newMode)
		}
	default:
		t.Fatal("expected mode change to be on modeCh")
	}
}

func TestTrigger_NonBlocking(t *testing.T) {
	t.Parallel()
	l := NewLoop(&mockScanner{}, ModePassive, nil)

	// Call Trigger twice — the channel has capacity 1, so the second
	// should silently drop without blocking.
	done := make(chan struct{})
	go func() {
		l.Trigger()
		l.Trigger()
		close(done)
	}()

	select {
	case <-done:
		// success — didn't block
	case <-time.After(time.Second):
		t.Fatal("Trigger() blocked — expected non-blocking send")
	}
}

func TestIsAlive_FalseBeforeStart(t *testing.T) {
	t.Parallel()
	l := NewLoop(&mockScanner{}, ModePassive, nil)

	if l.IsAlive() {
		t.Fatal("expected IsAlive=false before Start")
	}
}

func TestIsAlive_TrueAfterStart(t *testing.T) {
	t.Parallel()
	l := NewLoop(&mockScanner{}, ModePassive, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l.Start(ctx)

	// Give the goroutine a moment to set alive=true and run the initial scan.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if l.IsAlive() {
			return // success
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected IsAlive=true after Start within 1s")
}

func TestStop_MakesIsAliveFalse(t *testing.T) {
	t.Parallel()
	l := NewLoop(&mockScanner{}, ModePassive, nil)
	ctx := context.Background()

	l.Start(ctx)

	// Wait for alive
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if l.IsAlive() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !l.IsAlive() {
		t.Fatal("loop never became alive")
	}

	l.Stop()

	// Wait for not alive
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if !l.IsAlive() {
			return // success
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected IsAlive=false after Stop within 1s")
}

func TestScanModeConstants(t *testing.T) {
	t.Parallel()
	// Verify mode string values match what the API expects.
	if ModePassive != "passive" {
		t.Fatalf("expected %q, got %q", "passive", ModePassive)
	}
	if ModeActive != "active" {
		t.Fatalf("expected %q, got %q", "active", ModeActive)
	}
	if ModeFull != "full" {
		t.Fatalf("expected %q, got %q", "full", ModeFull)
	}
}
