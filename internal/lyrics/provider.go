package lyrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"lyrics-overlay/internal/cache"
	"lyrics-overlay/internal/overlay"
)

// LyricsProvider defines the interface for lyrics sources
type LyricsProvider interface {
	SearchLyrics(artist, title string) (*overlay.LyricsData, error)
	GetName() string
}

// Service manages lyrics fetching and caching
type Service struct {
	providers []LyricsProvider
	cache     *cache.Service
	client    *http.Client
}

// New creates a new lyrics service
func New(cacheSvc *cache.Service, geniusToken string) *Service {
	service := &Service{
		providers: make([]LyricsProvider, 0),
		cache:     cacheSvc,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Add Genius provider if token is available
	if geniusToken != "" {
		geniusProvider := NewGeniusProvider(geniusToken, service.client)
		service.AddProvider(geniusProvider)
	}

	return service
}

// AddProvider adds a lyrics provider
func (s *Service) AddProvider(provider LyricsProvider) {
	s.providers = append(s.providers, provider)
}

// GetLyrics fetches lyrics for a track, checking cache first
func (s *Service) GetLyrics(trackID, artist, title string) (*overlay.LyricsData, error) {
	// Check cache first by track ID
	if lyrics := s.cache.GetByTrackID(trackID); lyrics != nil {
		return lyrics, nil
	}

	// Normalize artist and title for cache lookup
	normalizedKey := normalizeForCache(artist, title)
	if lyrics := s.cache.GetByKey(normalizedKey); lyrics != nil {
		// Cache hit with normalized key, also cache by track ID
		s.cache.SetByTrackID(trackID, lyrics)
		return lyrics, nil
	}

	// No cache hit, fetch from providers
	for _, provider := range s.providers {
		lyrics, err := provider.SearchLyrics(artist, title)
		if err != nil {
			continue // Try next provider
		}

		if lyrics != nil && len(lyrics.Lines) > 0 {
			// Cache the result
			lyrics.TrackID = trackID
			s.cache.SetByTrackID(trackID, lyrics)
			s.cache.SetByKey(normalizedKey, lyrics)
			return lyrics, nil
		}
	}

	return nil, fmt.Errorf("no lyrics found for %s - %s", artist, title)
}

// normalizeForCache creates a normalized cache key from artist and title
func normalizeForCache(artist, title string) string {
	normalizedArtist := normalizeString(artist)
	normalizedTitle := normalizeString(title)
	return fmt.Sprintf("%s|%s", normalizedArtist, normalizedTitle)
}

// normalizeString normalizes text for lyrics matching
func normalizeString(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Remove common patterns
	patterns := []string{
		`\s*\(feat\..*?\)`,           // (feat. ...)
		`\s*\(ft\..*?\)`,             // (ft. ...)
		`\s*\(featuring.*?\)`,        // (featuring ...)
		`\s*\[.*?\]`,                 // [anything]
		`\s*\(.*?remix.*?\)`,         // (remix)
		`\s*\(.*?version.*?\)`,       // (version)
		`\s*\(.*?edit.*?\)`,          // (edit)
		`\s*-\s*remaster.*`,          // - remaster
		`\s*-\s*remix.*`,             // - remix
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	// Remove extra whitespace and special characters
	re := regexp.MustCompile(`[^\w\s]`)
	text = re.ReplaceAllString(text, "")
	
	// Normalize whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

// GeniusProvider implements lyrics fetching from Genius
type GeniusProvider struct {
	token  string
	client *http.Client
	baseURL string
}

// NewGeniusProvider creates a new Genius provider
func NewGeniusProvider(token string, client *http.Client) *GeniusProvider {
	return &GeniusProvider{
		token:   token,
		client:  client,
		baseURL: "https://api.genius.com",
	}
}

// GetName returns the provider name
func (g *GeniusProvider) GetName() string {
	return "Genius"
}

// SearchLyrics searches for lyrics on Genius
func (g *GeniusProvider) SearchLyrics(artist, title string) (*overlay.LyricsData, error) {
	// Search for the song
	searchQuery := fmt.Sprintf("%s %s", artist, title)
	songInfo, err := g.searchSong(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("genius search failed: %w", err)
	}

	if songInfo == nil {
		return nil, fmt.Errorf("song not found on Genius")
	}

	// For now, return a placeholder since full lyrics scraping requires
	// parsing the Genius song page HTML (which is more complex)
	// This is a starting point that can be extended
	lyrics := &overlay.LyricsData{
		Source:    "Genius",
		IsSynced:  false,
		FetchedAt: time.Now(),
		Lines: []overlay.LyricsLine{
			{Text: fmt.Sprintf("🎵 %s", songInfo.Title)},
			{Text: fmt.Sprintf("by %s", songInfo.Artist)},
			{Text: ""},
			{Text: "Lyrics available on Genius.com"},
			{Text: "Visit: " + songInfo.URL},
		},
	}

	return lyrics, nil
}

// GeniusSearchResponse represents the Genius API search response
type GeniusSearchResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Hits []struct {
			Result struct {
				ID             int    `json:"id"`
				Title          string `json:"title"`
				URL            string `json:"url"`
				PrimaryArtist struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

// SongInfo holds basic song information from Genius
type SongInfo struct {
	ID     int
	Title  string
	Artist string
	URL    string
}

// searchSong searches for a song on Genius
func (g *GeniusProvider) searchSong(query string) (*SongInfo, error) {
	// Prepare search URL
	searchURL := fmt.Sprintf("%s/search?q=%s", g.baseURL, url.QueryEscape(query))

	// Create request
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("User-Agent", "SpotLy/1.0")

	// Make request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genius API returned status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResp GeniusSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}

	// Find best match
	if len(searchResp.Response.Hits) == 0 {
		return nil, nil
	}

	// Return the first hit for now (could implement better matching logic)
	hit := searchResp.Response.Hits[0].Result
	return &SongInfo{
		ID:     hit.ID,
		Title:  hit.Title,
		Artist: hit.PrimaryArtist.Name,
		URL:    hit.URL,
	}, nil
}
