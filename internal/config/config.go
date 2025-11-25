package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all application configuration
type Config struct {
	// Spotify OAuth settings
	SpotifyClientID     string `json:"spotify_client_id"`
	SpotifyClientSecret string `json:"spotify_client_secret"`
	RedirectURI         string `json:"redirect_uri"`
	Port                int    `json:"port"`

	// Genius API settings
	GeniusToken string `json:"genius_token"`

	// Overlay settings
	Overlay OverlayConfig `json:"overlay"`

	// Auth tokens (persisted locally)
	Auth AuthConfig `json:"auth"`
}

// OverlayConfig holds overlay window settings
type OverlayConfig struct {
	X            int     `json:"x"`
	Y            int     `json:"y"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Opacity      float64 `json:"opacity"`
	FontSize     int     `json:"font_size"`
	Visible      bool    `json:"visible"`
	Locked       bool    `json:"locked"`
	Position     string  `json:"position"` // "top-left", "top-right", "bottom-left", "bottom-right"
	ResizeLocked bool    `json:"resize_locked"`
	SyncOffset   int64   `json:"sync_offset"` // Lyrics timing offset in ms (positive = earlier)
}

// AuthConfig holds OAuth tokens
type AuthConfig struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Service manages configuration persistence
type Service struct {
	config   *Config
	filePath string
}

// New creates a new config service
func New() (*Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".spotly")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	service := &Service{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	// Load existing config if it exists, otherwise create a default config file
	if _, err := os.Stat(configPath); err == nil {
		if err := service.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		if err := service.Save(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	return service, nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	return &Config{
		RedirectURI: "http://127.0.0.1:8080/callback",
		Port:        8080,
		Overlay: OverlayConfig{
			X:            100,
			Y:            100,
			Width:        600,
			Height:       120,
			Opacity:      0.9,
			FontSize:     16,
			Visible:      true,
			Locked:       false,
			Position:     "bottom-left",
			ResizeLocked: false,
			SyncOffset:   350,
		},
	}
}

// Get returns the current configuration
func (s *Service) Get() *Config {
	return s.config
}

// Set updates the configuration
func (s *Service) Set(config *Config) {
	s.config = config
}

// Load loads configuration from file
func (s *Service) Load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.config)
}

// Save saves configuration to file
func (s *Service) Save() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// Path returns the full path to the configuration file
func (s *Service) Path() string {
    return s.filePath
}

// UpdateOverlay updates overlay configuration
func (s *Service) UpdateOverlay(overlay OverlayConfig) error {
	s.config.Overlay = overlay
	return s.Save()
}

// UpdateAuth updates auth configuration
func (s *Service) UpdateAuth(auth AuthConfig) error {
	s.config.Auth = auth
	return s.Save()
}
