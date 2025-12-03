//go:build !windows

package main

import "fmt"

// GetActiveWindow returns the title of the currently active window (stub for non-Windows)
func (a *App) GetActiveWindow() (string, error) {
	return "", fmt.Errorf("GetActiveWindow not supported on this platform")
}

// IsOverlayFocused checks if the overlay window is currently focused (stub for non-Windows)
func (a *App) IsOverlayFocused() bool {
	return false
}

// resolveOverlayHWND is a no-op on non-Windows platforms
func (a *App) resolveOverlayHWND() {
	// No-op
}

// setOverlayClickThrough is a no-op on non-Windows platforms
func (a *App) setOverlayClickThrough(enable bool) {
	// No-op
}

// startClickThroughMonitor is a no-op on non-Windows platforms
func (a *App) startClickThroughMonitor() {
	// No-op on non-Windows platforms
}
