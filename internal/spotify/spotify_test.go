package spotify

import (
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	svc := New(nil, nil, nil)

	if svc.baseInterval != 2*time.Second {
		t.Errorf("Expected base interval 2s, got %v", svc.baseInterval)
	}
	if svc.currentInterval != 2*time.Second {
		t.Errorf("Expected current interval 2s, got %v", svc.currentInterval)
	}
	if svc.maxInterval != 30*time.Second {
		t.Errorf("Expected max interval 30s, got %v", svc.maxInterval)
	}
	if svc.backoffFactor != 1.5 {
		t.Errorf("Expected backoff factor 1.5, got %v", svc.backoffFactor)
	}
	if svc.isPolling {
		t.Error("Expected isPolling to be false initially")
	}
	if svc.consecutiveErrors != 0 {
		t.Error("Expected 0 consecutive errors initially")
	}
}

func TestStartStop(t *testing.T) {
	svc := New(nil, nil, nil)

	if svc.IsPolling() {
		t.Error("Expected not polling before Start")
	}

	// Start sets isPolling — note: pollLoop will panic without auth,
	// so we only test the flag, not actual polling
	svc.isPolling = true
	if !svc.IsPolling() {
		t.Error("Expected polling after setting flag")
	}

	// Simulate stop
	svc.isPolling = false
	if svc.IsPolling() {
		t.Error("Expected not polling after stop")
	}
}

func TestAdjustInterval_Playing(t *testing.T) {
	svc := New(nil, nil, nil)

	// When playing, should use base interval
	svc.currentInterval = 15 * time.Second
	svc.adjustInterval(true, false)

	if svc.currentInterval != svc.baseInterval {
		t.Errorf("Expected base interval %v when playing, got %v", svc.baseInterval, svc.currentInterval)
	}
}

func TestAdjustInterval_Paused(t *testing.T) {
	svc := New(nil, nil, nil)

	// When paused (not playing, no error), should use 3x base interval
	svc.adjustInterval(false, false)

	expected := svc.baseInterval * 3
	if svc.currentInterval != expected {
		t.Errorf("Expected %v when paused, got %v", expected, svc.currentInterval)
	}
}

func TestAdjustInterval_ErrorBackoff(t *testing.T) {
	svc := New(nil, nil, nil)
	initial := svc.currentInterval

	// Simulate error backoff
	svc.adjustInterval(false, true)
	after1 := svc.currentInterval

	if after1 <= initial {
		t.Errorf("Expected interval to increase on error: %v -> %v", initial, after1)
	}

	// Apply again — should keep increasing
	svc.adjustInterval(false, true)
	after2 := svc.currentInterval

	if after2 <= after1 {
		t.Errorf("Expected interval to keep increasing: %v -> %v", after1, after2)
	}
}

func TestAdjustInterval_MaxCapped(t *testing.T) {
	svc := New(nil, nil, nil)

	// Many errors should cap at max
	for i := 0; i < 20; i++ {
		svc.adjustInterval(false, true)
	}

	if svc.currentInterval > svc.maxInterval {
		t.Errorf("Expected interval capped at %v, got %v", svc.maxInterval, svc.currentInterval)
	}
	if svc.currentInterval != svc.maxInterval {
		t.Errorf("Expected interval to reach max %v, got %v", svc.maxInterval, svc.currentInterval)
	}
}

func TestResetInterval(t *testing.T) {
	svc := New(nil, nil, nil)

	// Set non-default values
	svc.currentInterval = 25 * time.Second
	svc.consecutiveErrors = 5

	svc.resetInterval()

	if svc.currentInterval != svc.baseInterval {
		t.Errorf("Expected interval reset to %v, got %v", svc.baseInterval, svc.currentInterval)
	}
	if svc.consecutiveErrors != 0 {
		t.Errorf("Expected errors reset to 0, got %d", svc.consecutiveErrors)
	}
}

func TestHandleRateLimit(t *testing.T) {
	svc := New(nil, nil, nil)

	// Rate limit should set interval to max
	svc.handleRateLimit(nil)

	if svc.currentInterval != svc.maxInterval {
		t.Errorf("Expected interval set to max %v on rate limit, got %v", svc.maxInterval, svc.currentInterval)
	}
}

func TestHandleError_IncreasesCount(t *testing.T) {
	svc := New(nil, nil, nil)

	if svc.consecutiveErrors != 0 {
		t.Error("Expected 0 errors initially")
	}

	// handleError increments the counter
	svc.handleError(nil)
	if svc.consecutiveErrors != 1 {
		t.Errorf("Expected 1 error after handleError, got %d", svc.consecutiveErrors)
	}

	svc.handleError(nil)
	if svc.consecutiveErrors != 2 {
		t.Errorf("Expected 2 errors, got %d", svc.consecutiveErrors)
	}
}

func TestHandleError_BackoffAfterThreshold(t *testing.T) {
	svc := New(nil, nil, nil)
	initial := svc.currentInterval

	// First two errors shouldn't trigger backoff
	svc.handleError(nil)
	svc.handleError(nil)

	if svc.currentInterval != initial {
		t.Errorf("Expected no backoff before threshold, interval changed: %v -> %v", initial, svc.currentInterval)
	}

	// Third error should trigger backoff
	svc.handleError(nil)
	if svc.currentInterval <= initial {
		t.Errorf("Expected backoff after 3 errors, interval: %v", svc.currentInterval)
	}
}
