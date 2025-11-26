package spotify

import (
	"context"
	"net/http"
	"time"

	"github.com/zmb3/spotify/v2"

	"lyrics-overlay/internal/auth"
	"lyrics-overlay/internal/lyrics"
	"lyrics-overlay/internal/overlay"
)

// Service handles Spotify API interactions and polling
type Service struct {
	auth              *auth.Service
	overlay           *overlay.Service
	lyrics            *lyrics.Service
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
func New(authSvc *auth.Service, overlaySvc *overlay.Service, lyricsSvc *lyrics.Service) *Service {
	return &Service{
		auth:            authSvc,
		overlay:         overlaySvc,
		lyrics:          lyricsSvc,
		stopChan:        make(chan struct{}),
		baseInterval:    5 * time.Second,  // Faster polling when playing
		currentInterval: 5 * time.Second,  // Current polling interval
		backoffFactor:   1.5,              // Exponential backoff factor
		maxInterval:     30 * time.Second, // Maximum polling interval
	}
}

// Start begins the Spotify polling service
func (s *Service) Start() {
	if s.isPolling {
		return
	}
	s.isPolling = true
	go s.pollLoop()
}

// Stop stops the Spotify polling service
func (s *Service) Stop() {
	if !s.isPolling {
		return
	}
	s.isPolling = false
	close(s.stopChan)
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
		s.adjustInterval(false, true)
		s.overlay.SetCurrentTrack(nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	playerState, err := client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		s.handleError(err)
		return
	}

	if playerState == nil || playerState.Item == nil {
		s.handleNoPlayback()
		return
	}

	// Extract track information
	track := s.extractTrackInfo(playerState)

	// Check if track changed
	if track.ID != s.lastTrackID {
		s.lastTrackID = track.ID
		s.resetInterval()

		// Fetch lyrics on track change
		if s.lyrics != nil {
			go s.fetchAndSetLyrics(track)
		}
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

// fetchAndSetLyrics queries the lyrics service and updates the overlay
func (s *Service) fetchAndSetLyrics(track *overlay.TrackInfo) {
	artist := ""
	if len(track.Artists) > 0 {
		artist = track.Artists[0]
	}
	lyrics, err := s.lyrics.GetLyrics(track.ID, artist, track.Name)
	if err != nil || lyrics == nil {
		// Clear lyrics if not found to avoid stale display
		s.overlay.SetCurrentLyrics(nil)
		return
	}
	s.overlay.SetCurrentLyrics(lyrics)
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
	if httpErr, ok := err.(*spotify.Error); ok && httpErr.Status == http.StatusTooManyRequests {
		s.handleRateLimit(httpErr)
		return
	}

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
	s.currentInterval = s.maxInterval
}

// handleNoPlayback handles when there's no currently playing content
func (s *Service) handleNoPlayback() {
	s.overlay.SetCurrentTrack(nil)
	s.adjustInterval(false, true)
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
