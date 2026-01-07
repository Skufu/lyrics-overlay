# Auth Package

> `internal/auth` — Spotify OAuth2 authentication

## Overview

Handles the complete OAuth2 flow for Spotify authentication, including:
- Authorization URL generation
- Callback server for token exchange
- Token persistence and refresh
- Client creation and management

---

## Types

### Service

```go
type Service struct {
    config        *config.Service
    authenticator *spotifyauth.Authenticator
    client        *spotify.Client
    server        *http.Server
    state         string
}
```

---

## Key Functions

### Constructor

```go
func New(configSvc *config.Service) (*Service, error)
```

Creates a new auth service. Returns error if Spotify credentials are missing.

**Scopes requested:**
- `ScopeUserReadCurrentlyPlaying` - Read current track
- `ScopeUserReadPlaybackState` - Read playback state

---

### Authentication Flow

```go
func (s *Service) StartOAuthFlow() error
```

1. Stops any existing callback server
2. Starts HTTP server on configured port (default: 8080)
3. Generates authorization URL with random state
4. Opens browser to Spotify login

```go
func (s *Service) handleCallback(w http.ResponseWriter, r *http.Request)
```

1. Verifies state parameter (CSRF protection)
2. Exchanges authorization code for tokens
3. Saves tokens to config
4. Creates Spotify client
5. Returns success HTML page

---

### Token Management

```go
func (s *Service) GetClient() *spotify.Client
```

Returns the authenticated client. Automatically refreshes token if expiring within 5 minutes.

```go
func (s *Service) refreshToken() error
```

Uses refresh token to get new access token. Updates stored tokens.

```go
func (s *Service) clearTokens()
```

Clears all stored authentication tokens.

---

### State Checks

```go
func (s *Service) IsAuthenticated() bool
```

Returns `true` if a valid Spotify client exists.

```go
func (s *Service) GetAuthURL() string
```

Returns the OAuth authorization URL for manual flow.

---

## OAuth Flow Diagram

```
┌──────────┐         ┌──────────┐         ┌─────────┐
│  SpotLy  │         │ Spotify  │         │ Browser │
└────┬─────┘         └────┬─────┘         └────┬────┘
     │                    │                    │
     │ StartOAuthFlow()   │                    │
     │────────────────────┼───────────────────>│
     │                    │                    │
     │                    │<───────────────────│
     │                    │  User logs in      │
     │                    │                    │
     │<───────────────────┼────────────────────│
     │  /callback?code=   │                    │
     │                    │                    │
     │ Exchange code      │                    │
     │───────────────────>│                    │
     │                    │                    │
     │<───────────────────│                    │
     │    Access Token    │                    │
     │                    │                    │
     │ Save & create      │                    │
     │ client             │                    │
     └────────────────────┴────────────────────┘
```

---

## Token Storage

Tokens are stored in config at `~/.spotly/config.json`:

```json
{
  "auth": {
    "access_token": "...",
    "refresh_token": "...",
    "token_type": "Bearer",
    "expires_at": 1704067200
  }
}
```

---

## See Also

- [Config Package](config.md) - Token persistence
- [Data Flow](../DATA_FLOW.md#authentication) - Authentication flow diagram
- [Troubleshooting](../TROUBLESHOOTING.md#oauth-issues) - OAuth errors
