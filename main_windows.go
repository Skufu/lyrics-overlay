//go:build windows

package main

import (
	"fmt"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows constants for extended window styles
const (
	_GWL_EXSTYLE       int32 = -20
	_WS_EX_TRANSPARENT int32 = 0x00000020
	_WS_EX_LAYERED     int32 = 0x00080000
)

// GetActiveWindow returns the title of the currently active window
func (a *App) GetActiveWindow() (string, error) {
	// Windows API calls to get the active window
	var (
		user32                  = windows.NewLazyDLL("user32.dll")
		procGetWindowText       = user32.NewProc("GetWindowTextW")
		procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	)

	// Get the handle to the foreground window
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", fmt.Errorf("no foreground window found")
	}

	// Get window title
	titleBuf := make([]uint16, 256)
	ret, _, _ := procGetWindowText.Call(
		hwnd,
		uintptr(unsafe.Pointer(&titleBuf[0])),
		uintptr(len(titleBuf)),
	)

	if ret == 0 {
		return "", fmt.Errorf("failed to get window title")
	}

	return windows.UTF16ToString(titleBuf), nil
}

// IsOverlayFocused checks if the overlay window is currently focused
func (a *App) IsOverlayFocused() bool {
	activeWindow, err := a.GetActiveWindow()
	if err != nil {
		return false
	}

	// Check if the active window is our overlay (title contains "SpotLy")
	return activeWindow == "SpotLy Overlay" || activeWindow == "SpotLy"
}

// resolveOverlayHWND finds and caches the HWND of the overlay window by its title
func (a *App) resolveOverlayHWND() {
	if a.overlayHWND != 0 {
		return
	}

	user32 := windows.NewLazyDLL("user32.dll")
	procFindWindowW := user32.NewProc("FindWindowW")

	title, _ := windows.UTF16PtrFromString("SpotLy Overlay")
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd != 0 {
		a.overlayHWND = hwnd
	}
}

// setOverlayClickThrough toggles WS_EX_TRANSPARENT so mouse events pass through the window
func (a *App) setOverlayClickThrough(enable bool) {
	a.resolveOverlayHWND()
	if a.overlayHWND == 0 {
		return
	}

	user32 := windows.NewLazyDLL("user32.dll")
	procGetWindowLongW := user32.NewProc("GetWindowLongW")
	procSetWindowLongW := user32.NewProc("SetWindowLongW")

	idx := _GWL_EXSTYLE
	exStyle, _, _ := procGetWindowLongW.Call(a.overlayHWND, uintptr(idx))
	cur := int32(exStyle)
	newStyle := cur | _WS_EX_LAYERED
	if enable {
		newStyle = newStyle | _WS_EX_TRANSPARENT
	} else {
		newStyle = newStyle &^ _WS_EX_TRANSPARENT
	}

	procSetWindowLongW.Call(a.overlayHWND, uintptr(idx), uintptr(newStyle))
	a.clickThrough = enable
}

func (a *App) startClickThroughMonitor() {
	if a.stopClickMonitor != nil {
		return // already running
	}

	a.stopClickMonitor = make(chan struct{})

	// List of games that require click-through (lowercase)
	gamesRequiringClickThrough := []string{
		"valorant",
		"league of legends",
		"cs2",
		"counter-strike",
		"dota 2",
		"overwatch",
		"apex legends",
	}

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				active, err := a.GetActiveWindow()
				if err != nil {
					continue
				}

				lower := strings.ToLower(active)
				isInGame := false

				// Check if any game in the list is the active window
				for _, game := range gamesRequiringClickThrough {
					if strings.Contains(lower, game) {
						isInGame = true
						break
					}
				}

				// Enable click-through (make unclickable) when in game
				// Disable click-through (make clickable) when not in game
				if isInGame && !a.clickThrough {
					a.setOverlayClickThrough(true) // Make unclickable
				} else if !isInGame && a.clickThrough {
					a.setOverlayClickThrough(false) // Make clickable
				}

			case <-a.stopClickMonitor:
				// Ensure click-through is disabled on shutdown so overlay is clickable
				if a.clickThrough {
					a.setOverlayClickThrough(false)
				}
				return
			}
		}
	}()
}
