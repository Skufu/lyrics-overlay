package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

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
		spotifySvc := spotify.New(authSvc, overlaySvc)
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

// Frontend API methods (these will be exposed to the frontend)

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
		return fmt.Errorf("auth service not initialized")
	}

	err := a.auth.StartOAuthFlow()
	if err != nil {
		return err
	}

	// Start Spotify polling after successful auth
	if a.spotify != nil && a.auth.IsAuthenticated() {
		a.spotify.Start()
	}

	return nil
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
	return a.overlay.GetDisplayInfo()
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
		OnStartup:        app.OnStartup,
		OnShutdown:       app.OnShutdown,
		WindowStartState: options.Minimised, // Start minimized

		// Bind App struct methods to frontend
		BackendInstance: app,
	})

	if err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		os.Exit(1)
	}
}
