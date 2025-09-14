package spotify

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/zmb3/spotify/v2"

	"lyrics-overlay/internal/auth"
	"lyrics-overlay/internal/overlay"
)

// Service handles Spotify API interactions and polling
type Service struct {
	auth              *auth.Service
	overlay           *overlay.Service
	stopChan          chan struct{}
	isPolling         bool
	baseInterval      time.Duration
	currentInterval   time.Duration
	backoffFactor     float64
	maxInterval       time.Duration
	lastTrackID       string
	consecutiveErrors int
}

// New creates a new Spotify service
func New(authSvc *auth.Service, overlaySvc *overlay.Service) *Service {
	return &Service{
		auth:            authSvc,
		overlay:         overlaySvc,
		stopChan:        make(chan struct{}),
		baseInterval:    4 * time.Second,  // Base polling interval
		currentInterval: 4 * time.Second,  // Current polling interval
		backoffFactor:   1.5,              // Exponential backoff factor
		maxInterval:     60 * time.Second, // Maximum polling interval
	}
}

// Start begins the Spotify polling service
func (s *Service) Start() {
	if s.isPolling {
		return
	}

	s.isPolling = true
	go s.pollLoop()
	log.Println("Spotify polling service started")
}

// Stop stops the Spotify polling service
func (s *Service) Stop() {
	if !s.isPolling {
		return
	}

	s.isPolling = false
	close(s.stopChan)
	log.Println("Spotify polling service stopped")
}

// pollLoop is the main polling loop
func (s *Service) pollLoop() {
	ticker := time.NewTicker(s.currentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.pollCurrentlyPlaying()

			// Update ticker with current interval
			ticker.Reset(s.currentInterval)
		}
	}
}

// pollCurrentlyPlaying polls the Spotify currently playing endpoint
func (s *Service) pollCurrentlyPlaying() {
	client := s.auth.GetClient()
	if client == nil {
		// Not authenticated, slow down polling
		s.adjustInterval(false, true)
		s.overlay.SetCurrentTrack(nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add jitter to prevent thundering herd
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(jitter)

	playerState, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		s.handleError(err)
		return
	}

	// Handle different response scenarios
	if playerState == nil || playerState.Item == nil {
		// No currently playing track or player not active
		s.handleNoPlayback()
		return
	}

	// Check if it's an ad or podcast (not music)
	if playerState.CurrentlyPlayingType != "track" {
		s.handleNonMusicContent()
		return
	}

	// Extract track information
	track := s.extractTrackInfo(playerState)

	// Check if track changed
	if track.ID != s.lastTrackID {
		log.Printf("Track changed: %s - %s", track.Artists[0], track.Name)
		s.lastTrackID = track.ID

		// Reset polling interval on track change
		s.resetInterval()
	}

	// Update overlay with current track
	s.overlay.SetCurrentTrack(track)

	// Adjust polling based on playback state
	if track.IsPlaying {
		s.adjustInterval(true, false)
	} else {
		s.adjustInterval(false, false)
	}

	// Reset error count on successful poll
	s.consecutiveErrors = 0
}

// extractTrackInfo extracts track information from Spotify API response
func (s *Service) extractTrackInfo(playerState *spotify.CurrentlyPlaying) *overlay.TrackInfo {
	track := playerState.Item

	artists := make([]string, len(track.Artists))
	for i, artist := range track.Artists {
		artists[i] = artist.Name
	}

	return &overlay.TrackInfo{
		ID:        track.ID.String(),
		Name:      track.Name,
		Artists:   artists,
		Album:     track.Album.Name,
		Duration:  int64(track.Duration),
		Progress:  int64(playerState.Progress),
		IsPlaying: playerState.Playing,
		UpdatedAt: time.Now(),
	}
}

// handleError handles API errors with appropriate backoff
func (s *Service) handleError(err error) {
	s.consecutiveErrors++

	// Check for rate limiting (429)
	if httpErr, ok := err.(*spotify.Error); ok {
		if httpErr.Status == http.StatusTooManyRequests {
			s.handleRateLimit(httpErr)
			return
		}
	}

	log.Printf("Spotify API error (attempt %d): %v", s.consecutiveErrors, err)

	// Exponential backoff for general errors
	if s.consecutiveErrors >= 3 {
		s.adjustInterval(false, true)
	}

	// Clear current track on persistent errors
	if s.consecutiveErrors >= 5 {
		s.overlay.SetCurrentTrack(nil)
	}
}

// handleRateLimit handles 429 rate limit responses
func (s *Service) handleRateLimit(err *spotify.Error) {
	log.Printf("Rate limited by Spotify API")

	// Check for Retry-After header
	retryAfter := 60 // Default to 60 seconds
	if retryAfterStr := err.Response.Header.Get("Retry-After"); retryAfterStr != "" {
		if ra, parseErr := strconv.Atoi(retryAfterStr); parseErr == nil {
			retryAfter = ra
		}
	}

	// Set interval to retry-after + some buffer
	s.currentInterval = time.Duration(retryAfter+10) * time.Second
	if s.currentInterval > s.maxInterval {
		s.currentInterval = s.maxInterval
	}

	log.Printf("Backing off for %v due to rate limit", s.currentInterval)
}

// handleNoPlayback handles when there's no currently playing content
func (s *Service) handleNoPlayback() {
	log.Println("No currently playing content")

	// Clear current track
	s.overlay.SetCurrentTrack(nil)

	// Slow down polling when nothing is playing
	s.adjustInterval(false, true)
}

// handleNonMusicContent handles ads, podcasts, etc.
func (s *Service) handleNonMusicContent() {
	log.Println("Non-music content playing (ad/podcast)")

	// Clear current track
	s.overlay.SetCurrentTrack(nil)

	// Slow down polling for non-music content
	s.adjustInterval(false, false)
}

// adjustInterval adjusts the polling interval based on current state
func (s *Service) adjustInterval(isPlaying, hasError bool) {
	if hasError {
		// Exponential backoff on errors
		s.currentInterval = time.Duration(float64(s.currentInterval) * s.backoffFactor)
		if s.currentInterval > s.maxInterval {
			s.currentInterval = s.maxInterval
		}
	} else if isPlaying {
		// Faster polling when music is playing
		s.currentInterval = s.baseInterval
	} else {
		// Slower polling when paused or no content
		s.currentInterval = s.baseInterval * 3
	}
}

// resetInterval resets the polling interval to base value
func (s *Service) resetInterval() {
	s.currentInterval = s.baseInterval
	s.consecutiveErrors = 0
}

// GetCurrentTrack returns the currently playing track
func (s *Service) GetCurrentTrack() *overlay.TrackInfo {
	return s.overlay.GetCurrentTrack()
}

// IsPolling returns whether the service is currently polling
func (s *Service) IsPolling() bool {
	return s.isPolling
}
