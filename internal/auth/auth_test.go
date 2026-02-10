package auth

import (
	"testing"

	"lyrics-overlay/internal/config"
)

// getTestConfig creates a config service for testing.
// Uses the standard New() which accesses ~/.spotly/config.json.
func getTestConfig(t *testing.T) *config.Service {
	t.Helper()
	svc, err := config.New()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	return svc
}

func TestNew_WithoutCredentials(t *testing.T) {
	cfgSvc := getTestConfig(t)
	// Clear credentials
	cfg := cfgSvc.Get()
	cfg.SpotifyClientID = ""
	cfg.SpotifyClientSecret = ""

	_, err := New(cfgSvc)
	if err == nil {
		t.Error("Expected error when creating auth service without credentials")
	}
}

func TestNew_WithCredentials(t *testing.T) {
	cfgSvc := getTestConfig(t)
	cfg := cfgSvc.Get()
	cfg.SpotifyClientID = "abcdefghijklmnopqrstuvwxyz123456"
	cfg.SpotifyClientSecret = "abcdefghijklmnopqrstuvwxyz789012"

	svc, err := New(cfgSvc)
	if err != nil {
		t.Fatalf("Expected no error with valid credentials, got: %v", err)
	}
	if svc == nil {
		t.Fatal("Expected non-nil service")
	}
}

func TestIsAuthenticated_NoClient(t *testing.T) {
	cfgSvc := getTestConfig(t)
	cfg := cfgSvc.Get()
	cfg.SpotifyClientID = "abcdefghijklmnopqrstuvwxyz123456"
	cfg.SpotifyClientSecret = "abcdefghijklmnopqrstuvwxyz789012"
	// Clear any stored tokens so auth starts fresh
	cfg.Auth.AccessToken = ""
	cfg.Auth.RefreshToken = ""
	cfg.Auth.ExpiresAt = 0
	cfgSvc.UpdateAuth(cfg.Auth)

	svc, err := New(cfgSvc)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Without valid tokens, should not be authenticated
	if svc.IsAuthenticated() {
		t.Error("Expected IsAuthenticated() false without valid tokens")
	}
}

func TestGetClient_NoAuth(t *testing.T) {
	cfgSvc := getTestConfig(t)
	cfg := cfgSvc.Get()
	cfg.SpotifyClientID = "abcdefghijklmnopqrstuvwxyz123456"
	cfg.SpotifyClientSecret = "abcdefghijklmnopqrstuvwxyz789012"
	// Clear any stored tokens
	cfg.Auth.AccessToken = ""
	cfg.Auth.RefreshToken = ""
	cfg.Auth.ExpiresAt = 0
	cfgSvc.UpdateAuth(cfg.Auth)

	svc, err := New(cfgSvc)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if svc.GetClient() != nil {
		t.Error("Expected nil client when no tokens stored")
	}
}

func TestLogout_ClearsState(t *testing.T) {
	cfgSvc := getTestConfig(t)
	cfg := cfgSvc.Get()
	cfg.SpotifyClientID = "abcdefghijklmnopqrstuvwxyz123456"
	cfg.SpotifyClientSecret = "abcdefghijklmnopqrstuvwxyz789012"

	svc, err := New(cfgSvc)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Logout should not panic even with no active session
	svc.Logout()

	if svc.GetClient() != nil {
		t.Error("Expected nil client after logout")
	}
	if svc.IsAuthenticated() {
		t.Error("Expected not authenticated after logout")
	}
}

func TestGetAuthURL_NotEmpty(t *testing.T) {
	cfgSvc := getTestConfig(t)
	cfg := cfgSvc.Get()
	cfg.SpotifyClientID = "abcdefghijklmnopqrstuvwxyz123456"
	cfg.SpotifyClientSecret = "abcdefghijklmnopqrstuvwxyz789012"

	svc, err := New(cfgSvc)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	url := svc.GetAuthURL()
	if url == "" {
		t.Error("Expected non-empty auth URL")
	}
	if len(url) < 50 {
		t.Errorf("Auth URL seems too short: %s", url)
	}
}

func TestGenerateRandomState(t *testing.T) {
	s1, err1 := generateRandomState()
	if err1 != nil {
		t.Fatalf("Failed: %v", err1)
	}
	s2, err2 := generateRandomState()
	if err2 != nil {
		t.Fatalf("Failed: %v", err2)
	}

	if s1 == "" || s2 == "" {
		t.Error("Expected non-empty state strings")
	}
	if s1 == s2 {
		t.Error("Expected different random states")
	}
}
