package overlay

import (
	"sync"
	"time"

	"lyrics-overlay/internal/config"
)

// Service manages the overlay window and lyrics display
type Service struct {
	config        *config.Service
	mu            sync.RWMutex
	currentTrack  *TrackInfo
	currentLyrics *LyricsData
	isVisible     bool
	lastUpdate    time.Time
}

// defaultSyncLeadMs is the default offset if not configured.
const defaultSyncLeadMs int64 = 350

// TrackInfo holds information about the currently playing track
type TrackInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Artists   []string  `json:"artists"`
	Album     string    `json:"album"`
	Duration  int64     `json:"duration_ms"`
	Progress  int64     `json:"progress_ms"`
	IsPlaying bool      `json:"is_playing"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LyricsData holds lyrics information
type LyricsData struct {
	TrackID   string       `json:"track_id"`
	Source    string       `json:"source"`
	Lines     []LyricsLine `json:"lines"`
	IsSynced  bool         `json:"is_synced"`
	FetchedAt time.Time    `json:"fetched_at"`
}

// LyricsLine represents a single line of lyrics
type LyricsLine struct {
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp_ms,omitempty"` // For synced lyrics
}

// New creates a new overlay service
func New(configSvc *config.Service) (*Service, error) {
	service := &Service{
		config:    configSvc,
		isVisible: configSvc.Get().Overlay.Visible,
	}

	return service, nil
}

// GetCurrentTrack returns the currently playing track information
func (s *Service) GetCurrentTrack() *TrackInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentTrack
}

// SetCurrentTrack updates the current track information
func (s *Service) SetCurrentTrack(track *TrackInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentTrack = track
	s.lastUpdate = time.Now()
}

// GetCurrentLyrics returns the current lyrics
func (s *Service) GetCurrentLyrics() *LyricsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLyrics
}

// SetCurrentLyrics updates the current lyrics
func (s *Service) SetCurrentLyrics(lyrics *LyricsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentLyrics = lyrics
}

// GetDisplayInfo returns the current lyrics lines to display
func (s *Service) GetDisplayInfo() *DisplayInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentTrack == nil || s.currentLyrics == nil {
		return &DisplayInfo{
			CurrentLine: "No track playing",
			NextLine:    "",
			IsPlaying:   false,
		}
	}

	// For synced lyrics, find current line based on progress
	if s.currentLyrics.IsSynced && len(s.currentLyrics.Lines) > 0 {
		// Derive effective progress using last known Spotify progress + elapsed time
		progress := s.currentTrack.Progress
		if s.currentTrack.IsPlaying {
			elapsed := time.Since(s.currentTrack.UpdatedAt).Milliseconds()
			if elapsed > 0 {
				progress += elapsed
			}
		}
		// Apply configurable sync offset (or default)
		syncOffset := s.config.Get().Overlay.SyncOffset
		if syncOffset == 0 {
			syncOffset = defaultSyncLeadMs
		}
		progress += syncOffset
		currentIdx := -1

		// Find the current lyrics line based on playback progress
		for i, line := range s.currentLyrics.Lines {
			if line.Timestamp <= progress {
				currentIdx = i
			} else {
				break
			}
		}

		if currentIdx >= 0 && currentIdx < len(s.currentLyrics.Lines) {
			currentLine := s.currentLyrics.Lines[currentIdx].Text
			lineStartTime := s.currentLyrics.Lines[currentIdx].Timestamp
			nextLine := ""
			nextLineTime := int64(0)

			// Find next non-empty line for preview and timing
			for j := currentIdx + 1; j < len(s.currentLyrics.Lines); j++ {
				if s.currentLyrics.Lines[j].Text != "" {
					nextLine = s.currentLyrics.Lines[j].Text
					nextLineTime = s.currentLyrics.Lines[j].Timestamp
					break
				} else if nextLineTime == 0 {
					// Use empty line's timestamp for duration calc
					nextLineTime = s.currentLyrics.Lines[j].Timestamp
				}
			}

			// Skip empty lines for current line too
			if currentLine == "" && currentIdx+1 < len(s.currentLyrics.Lines) {
				for j := currentIdx + 1; j < len(s.currentLyrics.Lines); j++ {
					if s.currentLyrics.Lines[j].Text != "" {
						currentLine = s.currentLyrics.Lines[j].Text
						lineStartTime = s.currentLyrics.Lines[j].Timestamp
						// Update next line
						for k := j + 1; k < len(s.currentLyrics.Lines); k++ {
							if s.currentLyrics.Lines[k].Text != "" {
								nextLine = s.currentLyrics.Lines[k].Text
								nextLineTime = s.currentLyrics.Lines[k].Timestamp
								break
							}
						}
						break
					}
				}
			}

			// Calculate line duration and progress
			lineDuration := int64(3000) // Default 3 seconds
			if nextLineTime > lineStartTime {
				lineDuration = nextLineTime - lineStartTime
			}
			lineProgress := progress - lineStartTime
			if lineProgress < 0 {
				lineProgress = 0
			}
			if lineProgress > lineDuration {
				lineProgress = lineDuration
			}

			return &DisplayInfo{
				CurrentLine:   currentLine,
				NextLine:      nextLine,
				IsPlaying:     s.currentTrack.IsPlaying,
				LineDuration:  lineDuration,
				LineProgress:  lineProgress,
				LineStartTime: lineStartTime,
			}
		}
	}

	// For non-synced lyrics, show first few lines
	if len(s.currentLyrics.Lines) > 0 {
		currentLine := s.currentLyrics.Lines[0].Text
		nextLine := ""
		if len(s.currentLyrics.Lines) > 1 {
			nextLine = s.currentLyrics.Lines[1].Text
		}

		return &DisplayInfo{
			CurrentLine: currentLine,
			NextLine:    nextLine,
			IsPlaying:   s.currentTrack.IsPlaying,
		}
	}

	return &DisplayInfo{
		CurrentLine: "No lyrics available",
		NextLine:    "Enjoying the instrumental vibes 🎸",
		IsPlaying:   s.currentTrack.IsPlaying,
	}
}

// DisplayInfo holds the information to display in the overlay
type DisplayInfo struct {
	CurrentLine   string `json:"current_line"`
	NextLine      string `json:"next_line"`
	IsPlaying     bool   `json:"is_playing"`
	LineDuration  int64  `json:"line_duration_ms"`   // Duration of current line in ms
	LineProgress  int64  `json:"line_progress_ms"`   // Progress into current line in ms
	LineStartTime int64  `json:"line_start_time_ms"` // Timestamp when current line started
}

// ToggleVisibility toggles the overlay visibility
func (s *Service) ToggleVisibility() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isVisible = !s.isVisible

	// Update config
	cfg := s.config.Get()
	cfg.Overlay.Visible = s.isVisible
	s.config.UpdateOverlay(cfg.Overlay)

	return s.isVisible
}

// IsVisible returns current visibility state
func (s *Service) IsVisible() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isVisible
}

// SetVisibility sets the overlay visibility
func (s *Service) SetVisibility(visible bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isVisible = visible

	// Update config
	cfg := s.config.Get()
	cfg.Overlay.Visible = visible
	s.config.UpdateOverlay(cfg.Overlay)
}

// GetOverlayConfig returns current overlay configuration
func (s *Service) GetOverlayConfig() config.OverlayConfig {
	return s.config.Get().Overlay
}

// UpdateOverlayConfig updates overlay configuration
func (s *Service) UpdateOverlayConfig(overlayConfig config.OverlayConfig) error {
	return s.config.UpdateOverlay(overlayConfig)
}

// Shutdown performs cleanup
func (s *Service) Shutdown() {
	// Save current state
	s.config.Save()
}
