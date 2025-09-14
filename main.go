package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailswindows "github.com/wailsapp/wails/v2/pkg/options/windows"
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
}

// OnShutdown is called when the app is shutting down
func (a *App) OnShutdown(ctx context.Context) {
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
			info.CurrentLine = "🔍 Spotify connected, polling for music..."
			info.NextLine = "Make sure Spotify is playing and not private session"
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

	return a.overlay.UpdateOverlayConfig(current)
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

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "SpotLy Overlay",
		Width:  400,
		Height: 200,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Frameless:        true,
		AlwaysOnTop:      true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0}, // Transparent
		Windows: &wailswindows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		OnStartup:        app.OnStartup,
		OnShutdown:       app.OnShutdown,
		WindowStartState: options.Minimised, // Start minimized
		Bind:             []interface{}{app},
	})

	if err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		os.Exit(1)
	}
}
