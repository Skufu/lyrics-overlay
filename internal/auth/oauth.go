package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"lyrics-overlay/internal/config"
)

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// Service handles Spotify OAuth2 authentication
type Service struct {
	config        *config.Service
	authenticator *spotifyauth.Authenticator
	client        *spotify.Client
	server        *http.Server
	state         string
}

// New creates a new auth service
func New(configSvc *config.Service) (*Service, error) {
	cfg := configSvc.Get()

	if cfg.SpotifyClientID == "" || cfg.SpotifyClientSecret == "" {
		return nil, fmt.Errorf("Spotify client ID and secret must be configured")
	}

	// Generate random state for OAuth security
	state, err := generateRandomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OAuth state: %w", err)
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(cfg.RedirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
		),
		spotifyauth.WithClientID(cfg.SpotifyClientID),
		spotifyauth.WithClientSecret(cfg.SpotifyClientSecret),
	)

	service := &Service{
		config:        configSvc,
		authenticator: auth,
		state:         state,
	}

	// If we have existing tokens, try to create a client
	if cfg.Auth.AccessToken != "" {
		service.createClientFromStoredTokens()
	}

	return service, nil
}

// generateRandomState generates a random state string for OAuth security
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// createClientFromStoredTokens creates a Spotify client from stored tokens
func (s *Service) createClientFromStoredTokens() {
	cfg := s.config.Get()

	token := &oauth2.Token{
		AccessToken:  cfg.Auth.AccessToken,
		RefreshToken: cfg.Auth.RefreshToken,
		TokenType:    cfg.Auth.TokenType,
		Expiry:       time.Unix(cfg.Auth.ExpiresAt, 0),
	}

	client := spotify.New(s.authenticator.Client(context.Background(), token))
	s.client = client

	// Test if token is still valid
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := client.CurrentUser(ctx); err != nil {
		// Token might be expired, try to refresh
		if s.refreshToken() != nil {
			// Refresh failed, clear stored tokens
			s.clearTokens()
		}
	}
}

// IsAuthenticated checks if the user is authenticated
func (s *Service) IsAuthenticated() bool {
	return s.client != nil
}

// GetClient returns the authenticated Spotify client
func (s *Service) GetClient() *spotify.Client {
	if s.client == nil {
		return nil
	}

	// Check if token needs refresh
	cfg := s.config.Get()
	if time.Now().Unix() >= cfg.Auth.ExpiresAt-300 { // Refresh 5 minutes before expiry
		if err := s.refreshToken(); err != nil {
			s.clearTokens()
			return nil
		}
	}

	return s.client
}

// StartOAuthFlow starts the OAuth2 authentication flow
func (s *Service) StartOAuthFlow() error {
	cfg := s.config.Get()

	// Stop any existing callback server first to prevent duplicates
	s.stopCallbackServer()

	// Start the callback server
	if err := s.startCallbackServer(cfg.Port); err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}

	// Generate the authorization URL
	authURL := s.authenticator.AuthURL(s.state)

	// Open the browser automatically
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Please visit this URL to authenticate:\n%s\n", authURL)
	}

	return nil
}

// startCallbackServer starts the HTTP server to handle OAuth callbacks
func (s *Service) startCallbackServer(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Callback server error: %v\n", err)
		}
	}()

	return nil
}

// handleCallback handles the OAuth callback
func (s *Service) handleCallback(w http.ResponseWriter, r *http.Request) {
	defer s.stopCallbackServer()

	// Check for errors
	if err := r.URL.Query().Get("error"); err != "" {
		http.Error(w, fmt.Sprintf("OAuth error: %s", err), http.StatusBadRequest)
		return
	}

	// Verify state
	state := r.URL.Query().Get("state")
	if state != s.state {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for tokens
	code := r.URL.Query().Get("code")
	token, err := s.authenticator.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Save tokens
	if err := s.saveTokens(token); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save tokens: %v", err), http.StatusInternalServerError)
		return
	}

	// Create Spotify client
	s.client = spotify.New(s.authenticator.Client(context.Background(), token))

	// Send success response
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>SpotLy - Authentication Successful</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap');
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Inter', -apple-system, sans-serif;
            background: #0a0a0f;
            color: #fff;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
        }
        .card {
            text-align: center;
            padding: 48px;
            background: linear-gradient(135deg, rgba(29, 185, 84, 0.08) 0%%, rgba(29, 185, 84, 0.02) 100%%);
            border: 1px solid rgba(29, 185, 84, 0.2);
            border-radius: 20px;
            max-width: 400px;
            animation: slideUp 0.5s ease;
        }
        @keyframes slideUp {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .check {
            width: 64px; height: 64px;
            background: linear-gradient(135deg, #1db954, #1ed760);
            border-radius: 50%%;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
            margin-bottom: 20px;
            box-shadow: 0 8px 24px rgba(29, 185, 84, 0.3);
            animation: pop 0.4s ease 0.2s both;
        }
        @keyframes pop {
            from { transform: scale(0); }
            to { transform: scale(1); }
        }
        h1 { font-size: 24px; font-weight: 700; margin-bottom: 8px; }
        p { color: rgba(255,255,255,0.6); font-size: 14px; line-height: 1.6; }
        .countdown { margin-top: 20px; font-size: 12px; color: rgba(255,255,255,0.3); }
    </style>
</head>
<body>
    <div class="card">
        <div class="check">✓</div>
        <h1>Connected!</h1>
        <p>SpotLy is now linked to your Spotify account.<br>You can close this window and return to the app.</p>
        <div class="countdown" id="cd">Closing in 3s...</div>
    </div>
    <script>
        let t=3;
        const el=document.getElementById('cd');
        setInterval(()=>{t--;if(t<=0){window.close();}else{el.textContent='Closing in '+t+'s...';}},1000);
    </script>
</body>
</html>`)
}

// stopCallbackServer stops the callback server
func (s *Service) stopCallbackServer() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(ctx)
		s.server = nil
	}
}

// saveTokens saves OAuth tokens to configuration
func (s *Service) saveTokens(token *oauth2.Token) error {
	cfg := s.config.Get()
	cfg.Auth = config.AuthConfig{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry.Unix(),
	}

	return s.config.UpdateAuth(cfg.Auth)
}

// refreshToken refreshes the OAuth token
func (s *Service) refreshToken() error {
	if s.client == nil {
		return fmt.Errorf("no client available")
	}

	cfg := s.config.Get()
	if cfg.Auth.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	token := &oauth2.Token{
		AccessToken:  cfg.Auth.AccessToken,
		RefreshToken: cfg.Auth.RefreshToken,
		TokenType:    cfg.Auth.TokenType,
		Expiry:       time.Unix(cfg.Auth.ExpiresAt, 0),
	}

	// Use the authenticator to refresh the token
	newToken, err := s.authenticator.RefreshToken(context.Background(), token)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save the new token
	if err := s.saveTokens(newToken); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	// Update the client
	s.client = spotify.New(s.authenticator.Client(context.Background(), newToken))

	return nil
}

// clearTokens clears stored authentication tokens
func (s *Service) clearTokens() {
	cfg := s.config.Get()
	cfg.Auth = config.AuthConfig{}
	_ = s.config.UpdateAuth(cfg.Auth)
	s.client = nil
}

// Logout clears authentication and logs out the user
func (s *Service) Logout() {
	s.clearTokens()
	s.stopCallbackServer()
}

// GetAuthURL returns the OAuth authorization URL
func (s *Service) GetAuthURL() string {
	return s.authenticator.AuthURL(s.state)
}
