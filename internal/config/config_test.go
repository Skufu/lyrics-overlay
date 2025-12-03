package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a service with the temp path
	service := &Service{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	// Save default config
	if err := service.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	if err := service.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := service.Get()
	if cfg.Port != 8080 {
		t.Errorf("Default port = %d; want 8080", cfg.Port)
	}

	if cfg.RedirectURI != "http://127.0.0.1:8080/callback" {
		t.Errorf("Unexpected redirect URI: %s", cfg.RedirectURI)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	service := &Service{
		filePath: configPath,
		config: &Config{
			SpotifyClientID: "test-id",
			Port:            9000,
		},
	}

	err := service.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify we can load it back
	if err := service.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := service.Get()
	if cfg.SpotifyClientID != "test-id" {
		t.Errorf("Expected SpotifyClientID 'test-id', got %s", cfg.SpotifyClientID)
	}
	if cfg.Port != 9000 {
		t.Errorf("Expected Port 9000, got %d", cfg.Port)
	}
}

func TestConfig_Load(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a config file manually
	cfg := &Config{
		SpotifyClientID: "loaded-id",
		Port:            9090,
		RedirectURI:     "http://127.0.0.1:9090/callback",
	}

	service := &Service{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	// Save the config
	service.Set(cfg)
	if err := service.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create a new service and load
	service2 := &Service{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	if err := service2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	loaded := service2.Get()
	if loaded.SpotifyClientID != "loaded-id" {
		t.Errorf("Expected SpotifyClientID 'loaded-id', got %s", loaded.SpotifyClientID)
	}
	if loaded.Port != 9090 {
		t.Errorf("Expected Port 9090, got %d", loaded.Port)
	}
}

func TestConfig_UpdateOverlay(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	service := &Service{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	overlayCfg := OverlayConfig{
		X:        200,
		Y:        300,
		Width:    800,
		Height:   200,
		Opacity:  0.8,
		FontSize: 18,
		Visible:  true,
		Locked:   false,
	}

	if err := service.UpdateOverlay(overlayCfg); err != nil {
		t.Fatalf("UpdateOverlay failed: %v", err)
	}

	cfg := service.Get()
	if cfg.Overlay.X != 200 {
		t.Errorf("Expected X 200, got %d", cfg.Overlay.X)
	}
	if cfg.Overlay.FontSize != 18 {
		t.Errorf("Expected FontSize 18, got %d", cfg.Overlay.FontSize)
	}
}

func TestConfig_UpdateAuth(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	service := &Service{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	authCfg := AuthConfig{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    1234567890,
	}

	if err := service.UpdateAuth(authCfg); err != nil {
		t.Fatalf("UpdateAuth failed: %v", err)
	}

	cfg := service.Get()
	if cfg.Auth.AccessToken != "test-token" {
		t.Errorf("Expected AccessToken 'test-token', got %s", cfg.Auth.AccessToken)
	}
	if cfg.Auth.TokenType != "Bearer" {
		t.Errorf("Expected TokenType 'Bearer', got %s", cfg.Auth.TokenType)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := getDefaultConfig()

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}

	if cfg.RedirectURI != "http://127.0.0.1:8080/callback" {
		t.Errorf("Expected default redirect URI, got %s", cfg.RedirectURI)
	}

	if cfg.Overlay.X != 100 {
		t.Errorf("Expected default overlay X 100, got %d", cfg.Overlay.X)
	}

	if cfg.Overlay.FontSize != 16 {
		t.Errorf("Expected default font size 16, got %d", cfg.Overlay.FontSize)
	}
}

