package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"path/filepath"
	stdruntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailswindows "github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"

	"lyrics-overlay/internal/auth"
	"lyrics-overlay/internal/cache"
	"lyrics-overlay/internal/config"
	"lyrics-overlay/internal/lyrics"
	"lyrics-overlay/internal/overlay"
	"lyrics-overlay/internal/spotify"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx     context.Context
	config  *config.Service
	cache   *cache.Service
	auth    *auth.Service
	overlay *overlay.Service
	spotify *spotify.Service
	lyrics  *lyrics.Service

	// Windows-specific: manage click-through state for overlay during games
	overlayHWND      uintptr
	clickThrough     bool
	stopClickMonitor chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the app starts up
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Initialize config service
	configSvc, err := config.New()
	if err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}
	a.config = configSvc

	// Initialize cache service
	cacheSvc := cache.New(100) // 100 entry cache
	a.cache = cacheSvc

	// Initialize overlay service
	overlaySvc, err := overlay.New(configSvc)
	if err != nil {
		fmt.Printf("Failed to initialize overlay: %v\n", err)
		os.Exit(1)
	}
	a.overlay = overlaySvc

	// Initialize auth service
	authSvc, err := auth.New(configSvc)
	if err != nil {
		fmt.Printf("Failed to initialize auth: %v\n", err)
		// Don't exit, we can still show the UI for authentication
	}
	a.auth = authSvc

	// Initialize lyrics service
	lyricsConfig := configSvc.Get()
	lyricsSvc := lyrics.New(cacheSvc, lyricsConfig.GeniusToken)
	a.lyrics = lyricsSvc

	// Initialize Spotify service
	if authSvc != nil {
		spotifySvc := spotify.New(authSvc, overlaySvc, lyricsSvc)
		a.spotify = spotifySvc

		// Start polling if authenticated
		if authSvc.IsAuthenticated() {
			spotifySvc.Start()
		}
	}

	// Start background monitor to toggle click-through during games (e.g., VALORANT)
	a.startClickThroughMonitor()
}

// OnShutdown is called when the app is shutting down
func (a *App) OnShutdown(ctx context.Context) {
	// Stop click-through monitor if running
	if a.stopClickMonitor != nil {
		select {
		case <-a.stopClickMonitor:
			// already closed
		default:
			close(a.stopClickMonitor)
		}
	}

	if a.spotify != nil {
		a.spotify.Stop()
	}
	if a.auth != nil {
		a.auth.Logout()
	}
	if a.overlay != nil {
		a.overlay.Shutdown()
	}
	if a.config != nil {
		a.config.Save()
	}
}

// IsAuthenticated checks if user is authenticated with Spotify
func (a *App) IsAuthenticated() bool {
	if a.auth == nil {
		return false
	}
	return a.auth.IsAuthenticated()
}

// StartOAuthFlow starts the Spotify OAuth flow
func (a *App) StartOAuthFlow() error {
	if a.auth == nil {
		return fmt.Errorf("auth service not initialized - check that Spotify credentials are configured in ~/.spotly/config.json")
	}

	err := a.auth.StartOAuthFlow()
	if err != nil {
		return fmt.Errorf("failed to start OAuth flow: %w", err)
	}

	return nil
}

// StartSpotifyPolling manually starts Spotify polling (for use after auth)
func (a *App) StartSpotifyPolling() bool {
	if a.spotify != nil && a.auth != nil && a.auth.IsAuthenticated() {
		if !a.spotify.IsPolling() {
			a.spotify.Start()
			return true
		}
	}
	return false
}

// GetAuthURL returns the OAuth URL for manual authentication
func (a *App) GetAuthURL() (string, error) {
	if a.auth == nil {
		return "", fmt.Errorf("auth service not initialized - check that Spotify credentials are configured")
	}
	return a.auth.GetAuthURL(), nil
}

// GetDisplayInfo returns current lyrics display information
func (a *App) GetDisplayInfo() *overlay.DisplayInfo {
	if a.overlay == nil {
		return &overlay.DisplayInfo{
			CurrentLine: "Service not available",
			NextLine:    "",
			IsPlaying:   false,
		}
	}

	info := a.overlay.GetDisplayInfo()

	// Add debugging info if no track is playing
	if info.CurrentLine == "No track playing" && a.auth != nil && a.auth.IsAuthenticated() {
		if a.spotify != nil && a.spotify.IsPolling() {
			info.CurrentLine = "🎧 Ready and waiting"
			info.NextLine = "Start playing music in Spotify"
		} else {
			info.CurrentLine = "⚠️ Spotify connected but polling stopped"
			info.NextLine = "Try restarting the app"
		}
	}

	return info
}

// GetSpotifyStatus returns debug info about Spotify connection
func (a *App) GetSpotifyStatus() map[string]interface{} {
	status := map[string]interface{}{
		"authenticated": false,
		"polling":       false,
		"has_client":    false,
		"current_track": nil,
	}

	if a.auth != nil {
		status["authenticated"] = a.auth.IsAuthenticated()
		status["has_client"] = a.auth.GetClient() != nil
	}

	if a.spotify != nil {
		status["polling"] = a.spotify.IsPolling()
	}

	if a.overlay != nil {
		currentTrack := a.overlay.GetCurrentTrack()
		if currentTrack != nil {
			status["current_track"] = map[string]interface{}{
				"name":    currentTrack.Name,
				"artists": currentTrack.Artists,
				"playing": currentTrack.IsPlaying,
				"id":      currentTrack.ID,
			}
		}
	}

	return status
}

// TestSpotifyConnection manually tests the Spotify API connection
func (a *App) TestSpotifyConnection() string {
	if a.auth == nil {
		return "❌ Auth service not available"
	}

	if !a.auth.IsAuthenticated() {
		return "❌ Not authenticated"
	}

	client := a.auth.GetClient()
	if client == nil {
		return "❌ No Spotify client"
	}

	// Test API call
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playerState, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return fmt.Sprintf("❌ API Error: %v", err)
	}

	if playerState == nil {
		return "⚠️ No active playback (start music in Spotify)"
	}

	if playerState.Item == nil {
		return "⚠️ No track item (ads or podcast?)"
	}

	return fmt.Sprintf("✅ Found: %s by %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
}

// RefreshNow forces an immediate Spotify poll and lyrics fetch
func (a *App) RefreshNow() string {
	if a.spotify == nil {
		return "❌ Spotify service not available"
	}

	if a.auth == nil || !a.auth.IsAuthenticated() {
		return "❌ Not authenticated"
	}

	// Force a poll
	client := a.auth.GetClient()
	if client == nil {
		return "❌ No Spotify client"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playerState, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return fmt.Sprintf("❌ API Error: %v", err)
	}

	if playerState == nil || playerState.Item == nil {
		a.overlay.SetCurrentTrack(nil)
		return "⚠️ No active playback"
	}

	// Extract and set track info
	track := &overlay.TrackInfo{
		ID:        playerState.Item.ID.String(),
		Name:      playerState.Item.Name,
		Artists:   []string{playerState.Item.Artists[0].Name},
		Album:     playerState.Item.Album.Name,
		Duration:  int64(playerState.Item.Duration),
		Progress:  int64(playerState.Progress),
		IsPlaying: playerState.Playing,
		UpdatedAt: time.Now(),
	}

	a.overlay.SetCurrentTrack(track)

	// Try to fetch lyrics if we have the lyrics service
	if a.lyrics != nil {
		go func() {
			lyrics, err := a.lyrics.GetLyrics(track.ID, track.Artists[0], track.Name)
			if err == nil && lyrics != nil {
				a.overlay.SetCurrentLyrics(lyrics)
			} else {
				// If lyrics failed, clear any old lyrics
				a.overlay.SetCurrentLyrics(nil)
			}
		}()
	}

	return fmt.Sprintf("✅ Refreshed: %s by %s", track.Name, track.Artists[0])
}

// ToggleVisibility toggles overlay visibility
func (a *App) ToggleVisibility() bool {
	if a.overlay == nil {
		return false
	}
	return a.overlay.ToggleVisibility()
}

// ResizeWindow resizes the overlay window with smooth transition
func (a *App) ResizeWindow(width, height int) error {
	if a.ctx == nil {
		return fmt.Errorf("context not available")
	}

	// Get current window position to maintain center point
	x, y := runtime.WindowGetPosition(a.ctx)

	// Calculate new position to keep window centered at same spot
	// (optional - comment out if you want it to grow from top-left)
	currentWidth, currentHeight := runtime.WindowGetSize(a.ctx)
	deltaWidth := (currentWidth - width) / 2
	deltaHeight := (currentHeight - height) / 2
	newX := x + deltaWidth
	newY := y + deltaHeight

	// Set new size
	runtime.WindowSetSize(a.ctx, width, height)

	// Maintain center position (optional)
	runtime.WindowSetPosition(a.ctx, newX, newY)

	return nil
}

// UpdateOverlayConfig updates overlay configuration
func (a *App) UpdateOverlayConfig(config map[string]interface{}) error {
	if a.overlay == nil {
		return fmt.Errorf("overlay service not available")
	}

	current := a.overlay.GetOverlayConfig()

	// Update fields if provided
	if opacity, ok := config["opacity"].(float64); ok {
		current.Opacity = opacity
	}
	if fontSize, ok := config["font_size"].(float64); ok {
		current.FontSize = int(fontSize)
	}
	if visible, ok := config["visible"].(bool); ok {
		current.Visible = visible
	}
	if locked, ok := config["locked"].(bool); ok {
		current.Locked = locked
	}
	if position, ok := config["position"].(string); ok {
		current.Position = position
	}
	if resizeLocked, ok := config["resize_locked"].(bool); ok {
		current.ResizeLocked = resizeLocked
	}
	if syncOffset, ok := config["sync_offset"].(float64); ok {
		current.SyncOffset = int64(syncOffset)
	}

	return a.overlay.UpdateOverlayConfig(current)
}

// GetOverlayConfig returns current overlay configuration
func (a *App) GetOverlayConfig() config.OverlayConfig {
	if a.overlay == nil {
		return config.OverlayConfig{}
	}
	return a.overlay.GetOverlayConfig()
}

// Quit closes the application
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// GetConfigPath returns the full path to the user's config file
func (a *App) GetConfigPath() string {
	if a.config == nil {
		return ""
	}
	return a.config.Path()
}

// OpenConfig opens the user's config file location in Explorer (Windows) and returns the path
func (a *App) OpenConfig() (string, error) {
	if a.config == nil {
		return "", fmt.Errorf("config service not available")
	}
	path := a.config.Path()
	// Best-effort: ensure the file exists on disk
	_ = a.config.Save()
	// Windows: open Explorer highlighting the config file
	_ = exec.Command("explorer.exe", "/select,", path).Start()
	return path, nil
}

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

// Windows constants for extended window styles
const (
	_GWL_EXSTYLE       int32 = -20
	_WS_EX_TRANSPARENT int32 = 0x00000020
	_WS_EX_LAYERED     int32 = 0x00080000
)

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

// OpenConfigDirectory opens the config folder in file explorer
func (a *App) OpenConfigDirectory() error {
	configDir := filepath.Dir(a.config.Path())
	var cmd *exec.Cmd

	switch stdruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", configDir)
	case "darwin":
		cmd = exec.Command("open", configDir)
	case "linux":
		cmd = exec.Command("xdg-open", configDir)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// SaveSpotifyCredentials saves credentials from the UI
func (a *App) SaveSpotifyCredentials(clientID, clientSecret string) error {
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	cfg := a.config.Get()
	cfg.SpotifyClientID = clientID
	cfg.SpotifyClientSecret = clientSecret
	cfg.RedirectURI = "http://127.0.0.1:8080/callback"
	cfg.Port = 8080

	if err := a.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Reinitialize auth service with new credentials
	authSvc, err := auth.New(a.config)
	if err != nil {
		return fmt.Errorf("failed to initialize auth: %w", err)
	}
	a.auth = authSvc

	return nil
}

// ValidateCredentials tests if the provided credentials work
func (a *App) ValidateCredentials(clientID, clientSecret string) error {
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("credentials cannot be empty")
	}

	// Basic validation - check format
	if len(clientID) < 32 {
		return fmt.Errorf("client ID appears invalid (too short)")
	}

	if len(clientSecret) < 32 {
		return fmt.Errorf("client secret appears invalid (too short)")
	}

	return nil
}

// HasCredentials checks if Spotify credentials are configured
func (a *App) HasCredentials() bool {
	cfg := a.config.Get()
	return cfg.SpotifyClientID != "" && cfg.SpotifyClientSecret != ""
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Preload config to determine startup options (e.g., disable resize)
	preConfig, _ := config.New()
	disableResizeAtStartup := true // Default to disabled resize
	if preConfig != nil {
		cfg := preConfig.Get()
		disableResizeAtStartup = cfg.Overlay.ResizeLocked
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "SpotLy Overlay",
		Width:  600,
		Height: 500, // Start with auth screen size (will resize to 120 after auth)
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Frameless:        true,
		AlwaysOnTop:      true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0}, // Transparent
		DisableResize:    disableResizeAtStartup,
		Windows: &wailswindows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableFramelessWindowDecorations: true,
		},
		OnStartup:        app.OnStartup,
		OnShutdown:       app.OnShutdown,
		WindowStartState: options.Normal,
		Bind:             []interface{}{app},
	})

	if err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		os.Exit(1)
	}
}
