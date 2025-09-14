package lyrics

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	nethtml "golang.org/x/net/html"

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

	// Add LRCLIB provider first (often returns synced lyrics)
	lrclibProvider := NewLRCLibProvider(service.client)
	service.AddProvider(lrclibProvider)

	// Add Genius provider next if token is available
	if geniusToken != "" {
		geniusProvider := NewGeniusProvider(geniusToken, service.client)
		service.AddProvider(geniusProvider)
	}

	// Add demo provider as a fallback
	demoProvider := NewDemoProvider()
	service.AddProvider(demoProvider)

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
		// Don't accept demo/info cache as final result
		if strings.EqualFold(lyrics.Source, "Info") || strings.EqualFold(lyrics.Source, "Demo") {
			log.Printf("Lyrics cache hit is Info/Demo for %s - %s, ignoring and refetching", artist, title)
		} else {
			return lyrics, nil
		}
	}

	// Normalize artist and title for cache lookup
	normalizedKey := normalizeForCache(artist, title)
	if lyrics := s.cache.GetByKey(normalizedKey); lyrics != nil {
		// Cache hit with normalized key, also cache by track ID
		if strings.EqualFold(lyrics.Source, "Info") || strings.EqualFold(lyrics.Source, "Demo") {
			log.Printf("Lyrics cache(key) is Info/Demo for %s - %s, ignoring and refetching", artist, title)
		} else {
			s.cache.SetByTrackID(trackID, lyrics)
			return lyrics, nil
		}
	}

	// No cache hit, fetch from providers
	for _, provider := range s.providers {
		log.Printf("Lyrics: trying provider %s for %s - %s", provider.GetName(), artist, title)
		lyrics, err := provider.SearchLyrics(artist, title)
		if err != nil {
			log.Printf("Lyrics: provider %s error: %v", provider.GetName(), err)
			continue // Try next provider
		}

		if lyrics != nil && len(lyrics.Lines) > 0 {
			// Cache the result (but skip caching demo/info fallback)
			lyrics.TrackID = trackID
			if !(strings.EqualFold(lyrics.Source, "Info") || strings.EqualFold(lyrics.Source, "Demo")) {
				s.cache.SetByTrackID(trackID, lyrics)
				s.cache.SetByKey(normalizedKey, lyrics)
			} else {
				log.Printf("Lyrics: not caching Info/Demo result for %s - %s", artist, title)
			}
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
		`\s*\(feat\..*?\)`,     // (feat. ...)
		`\s*\(ft\..*?\)`,       // (ft. ...)
		`\s*\(featuring.*?\)`,  // (featuring ...)
		`\s*\[.*?\]`,           // [anything]
		`\s*\(.*?remix.*?\)`,   // (remix)
		`\s*\(.*?version.*?\)`, // (version)
		`\s*\(.*?edit.*?\)`,    // (edit)
		`\s*-\s*remaster.*`,    // - remaster
		`\s*-\s*remix.*`,       // - remix
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
	token   string
	client  *http.Client
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

	// Fetch and parse actual lyrics from the Genius song page
	rawText, err := g.fetchLyricsFromPage(songInfo.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}

	cleaned := strings.TrimSpace(rawText)
	if cleaned == "" {
		return nil, fmt.Errorf("no lyrics found on Genius page")
	}

	lines := textToLyricsLines(cleaned)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no usable lyrics lines parsed")
	}

	lyrics := &overlay.LyricsData{
		Source:    "Genius",
		IsSynced:  false,
		FetchedAt: time.Now(),
		Lines:     lines,
	}

	return lyrics, nil
}

// fetchLyricsFromPage downloads and extracts lyrics text from a Genius song page URL
func (g *GeniusProvider) fetchLyricsFromPage(pageURL string) (string, error) {
	// Create request
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", err
	}

	// Set headers to look like a regular browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("genius page returned status %d", resp.StatusCode)
	}

	// Parse HTML
	root, err := nethtml.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse genius page: %w", err)
	}

	// Find containers with data-lyrics-container="true"
	containers := make([]*nethtml.Node, 0, 8)
	findNodes(root, func(n *nethtml.Node) bool {
		if n.Type == nethtml.ElementNode {
			var hasContainer bool
			lang := ""
			for _, a := range n.Attr {
				if a.Key == "data-lyrics-container" && (a.Val == "true" || a.Val == "") {
					hasContainer = true
				}
				if a.Key == "data-lyrics-language" {
					lang = strings.ToLower(a.Val)
				}
			}
			if hasContainer {
				// Prefer English or unspecified language, skip obvious translation blocks
				if lang == "" || strings.HasPrefix(lang, "en") {
					return true
				}
			}
		}
		return false
	}, &containers)

	// Fallback: look for class containing "Lyrics__Container" (filter translations/footers)
	if len(containers) == 0 {
		findNodes(root, func(n *nethtml.Node) bool {
			if n.Type == nethtml.ElementNode {
				var cls string
				for _, a := range n.Attr {
					if a.Key == "class" {
						cls = a.Val
					}
				}
				if strings.Contains(cls, "Lyrics__Container") &&
					!strings.Contains(strings.ToLower(cls), "translation") &&
					!strings.Contains(strings.ToLower(cls), "contributor") &&
					!strings.Contains(strings.ToLower(cls), "footer") {
					return true
				}
			}
			return false
		}, &containers)
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("could not locate lyrics container")
	}

	var sb strings.Builder
	for i, c := range containers {
		if i > 0 {
			sb.WriteString("\n")
		}
		writeNodeText(c, &sb)
	}

	text := sb.String()
	text = html.UnescapeString(text)

	// Normalize line endings and collapse excessive blank lines
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	reNL := regexp.MustCompile(`\n{3,}`)
	text = reNL.ReplaceAllString(text, "\n\n")

	return text, nil
}

// findNodes traverses the HTML node tree and collects nodes matching the predicate
func findNodes(n *nethtml.Node, pred func(*nethtml.Node) bool, out *[]*nethtml.Node) {
	if n == nil {
		return
	}
	if pred(n) {
		*out = append(*out, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findNodes(c, pred, out)
	}
}

// writeNodeText writes the visible text content of an HTML node to the builder
func writeNodeText(n *nethtml.Node, sb *strings.Builder) {
	if n == nil {
		return
	}
	switch n.Type {
	case nethtml.TextNode:
		sb.WriteString(n.Data)
	case nethtml.ElementNode:
		// Insert newlines for line/paragraph breaks
		switch strings.ToLower(n.Data) {
		case "br":
			sb.WriteString("\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNodeText(c, sb)
		}
		switch strings.ToLower(n.Data) {
		case "p", "div":
			sb.WriteString("\n")
		}
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNodeText(c, sb)
		}
	}
}

// textToLyricsLines converts raw lyrics text into overlay lines, filtering noise
func textToLyricsLines(text string) []overlay.LyricsLine {
	// Split lines, trim, and filter common non-lyrics artifacts
	rawLines := strings.Split(text, "\n")
	lines := make([]overlay.LyricsLine, 0, len(rawLines))

	// Helpers
	isSkippable := func(s string) bool {
		t := strings.TrimSpace(strings.ToLower(s))
		if t == "" {
			return false // keep empties for spacing (dedup below)
		}
		if strings.Contains(t, "you might also like") {
			return true
		}
		if strings.Contains(t, "genius annotation") {
			return true
		}
		if strings.HasPrefix(t, "see ") {
			return true
		}
		// e.g., "123Embed"
		re := regexp.MustCompile(`^\d+\s*embed$`)
		if re.MatchString(t) {
			return true
		}

		// Skip contributor/translation UI strings from Genius
		if strings.Contains(t, "contributors") {
			return true
		}
		if strings.Contains(t, "translation") || strings.Contains(t, "translations") {
			return true
		}

		// Skip standalone language names often listed under translations
		langWords := map[string]struct{}{
			"cesky": {}, "česky": {}, "čeština": {}, "deutsch": {}, "français": {}, "francais": {},
			"español": {}, "espanol": {}, "português": {}, "portugues": {}, "italiano": {}, "polski": {},
			"nederlands": {}, "svenska": {}, "suomi": {}, "dansk": {}, "norsk": {}, "русский": {},
			"русский язык": {}, "bahasa": {}, "bahasa indonesia": {}, "tiếng": {}, "tiếng việt": {}, "tieng viet": {},
			"türkçe": {}, "turkce": {}, "العربية": {}, "hebrew": {}, "עברית": {},
			"日本語": {}, "한국어": {}, "中文": {}, "简体中文": {}, "繁體中文": {}, "ไทย": {},
		}
		ws := regexp.MustCompile(`\s+`)
		norm := ws.ReplaceAllString(t, " ")
		tokens := strings.Fields(norm)
		if len(tokens) > 0 && len(tokens) <= 3 {
			allLang := true
			for _, tok := range tokens {
				if _, ok := langWords[tok]; !ok {
					allLang = false
					break
				}
			}
			if allLang {
				return true
			}
		}

		return false
	}

	lastWasEmpty := false
	for _, l := range rawLines {
		t := strings.TrimSpace(l)
		if isSkippable(t) {
			continue
		}
		if t == "" {
			if lastWasEmpty {
				continue
			}
			lines = append(lines, overlay.LyricsLine{Text: ""})
			lastWasEmpty = true
			continue
		}
		lines = append(lines, overlay.LyricsLine{Text: t})
		lastWasEmpty = false
	}

	// Trim leading/trailing empty lines
	for len(lines) > 0 && lines[0].Text == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1].Text == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// GeniusSearchResponse represents the Genius API search response
type GeniusSearchResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Hits []struct {
			Result struct {
				ID            int    `json:"id"`
				Title         string `json:"title"`
				URL           string `json:"url"`
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

// LRCLibProvider implements lyrics fetching from LRCLIB
type LRCLibProvider struct {
	client  *http.Client
	baseURL string
}

// NewLRCLibProvider creates a new LRCLIB provider
func NewLRCLibProvider(client *http.Client) *LRCLibProvider {
	return &LRCLibProvider{
		client:  client,
		baseURL: "https://lrclib.net/api",
	}
}

// GetName returns the provider name
func (l *LRCLibProvider) GetName() string {
	return "LRCLIB"
}

// lrcLibTrack is the structure returned by LRCLIB
type lrcLibTrack struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"` // seconds
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// SearchLyrics queries LRCLIB for lyrics
func (l *LRCLibProvider) SearchLyrics(artist, title string) (*overlay.LyricsData, error) {
	// First, try direct get endpoint for an exact match
	if track := l.tryGet(artist, title); track != nil {
		if data := l.trackToLyricsData(track); data != nil {
			return data, nil
		}
	}

	// Fallback to search endpoint
	results, err := l.search(artist, title)
	if err != nil {
		return nil, err
	}

	// If empty, try query fallback
	if len(results) == 0 {
		q := strings.TrimSpace(fmt.Sprintf("%s %s", title, artist))
		if q != "" {
			results, err = l.searchByQuery(q)
			if err != nil {
				return nil, err
			}
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("no lrclib results")
		}
	}

	// Score and pick best match
	best := pickBestLRCLibMatch(results, artist, title)
	if best == nil {
		best = &results[0]
	}

	// Important: LRCLIB search results may not include lyrics; fetch by ID
	full, err := l.getByID(best.ID)
	if err == nil && full != nil {
		if data := l.trackToLyricsData(full); data != nil {
			return data, nil
		}
	}

	// Fallback to whatever search returned (if it had lyrics fields)
	data := l.trackToLyricsData(best)
	if data == nil {
		return nil, fmt.Errorf("lrclib returned empty lyrics")
	}
	return data, nil
}

func (l *LRCLibProvider) tryGet(artist, title string) *lrcLibTrack {
	endpoint := fmt.Sprintf("%s/get?track_name=%s&artist_name=%s", l.baseURL, url.QueryEscape(title), url.QueryEscape(artist))
	// Note: duration/album params can be added if available from caller
	// e.g., &album_name=...&duration=...
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var track lrcLibTrack
	if err := json.Unmarshal(body, &track); err != nil {
		return nil
	}
	if track.PlainLyrics == "" && track.SyncedLyrics == "" {
		return nil
	}
	return &track
}

func (l *LRCLibProvider) search(artist, title string) ([]lrcLibTrack, error) {
	endpoint := fmt.Sprintf("%s/search?track_name=%s&artist_name=%s", l.baseURL, url.QueryEscape(title), url.QueryEscape(artist))
	// Note: duration/album params can be added if available from caller
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib search status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var results []lrcLibTrack
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (l *LRCLibProvider) searchByQuery(query string) ([]lrcLibTrack, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s", l.baseURL, url.QueryEscape(query))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SpotLy/1.0")
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib search status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var results []lrcLibTrack
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func pickBestLRCLibMatch(results []lrcLibTrack, artist, title string) *lrcLibTrack {
	nArtist := normalizeString(artist)
	nTitle := normalizeString(title)

	bestIdx := -1
	bestScore := -1
	for i, r := range results {
		artistMatch := normalizeString(r.ArtistName) == nArtist
		titleMatch := normalizeString(r.TrackName) == nTitle
		score := 0
		if artistMatch {
			score += 3
		}
		if titleMatch {
			score += 3
		}
		if r.SyncedLyrics != "" {
			score += 2
		}
		if r.PlainLyrics != "" {
			score += 1
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		return &results[bestIdx]
	}
	return nil
}

func (l *LRCLibProvider) trackToLyricsData(track *lrcLibTrack) *overlay.LyricsData {
	if track == nil {
		return nil
	}
	if track.SyncedLyrics != "" {
		lines := parseLRCToLines(track.SyncedLyrics)
		if len(lines) > 0 {
			return &overlay.LyricsData{
				Source:    "LRCLIB",
				IsSynced:  true,
				FetchedAt: time.Now(),
				Lines:     lines,
			}
		}
	}
	if track.PlainLyrics != "" {
		lines := textToLyricsLines(track.PlainLyrics)
		if len(lines) > 0 {
			return &overlay.LyricsData{
				Source:    "LRCLIB",
				IsSynced:  false,
				FetchedAt: time.Now(),
				Lines:     lines,
			}
		}
	}
	return nil
}

// parseLRCToLines parses LRC formatted lyrics into timestamped lines
func parseLRCToLines(lrc string) []overlay.LyricsLine {
	lines := make([]overlay.LyricsLine, 0)
	// Timestamp pattern: [mm:ss.xx] or [mm:ss.xxx]
	re := regexp.MustCompile(`\[(\d{1,2}):(\d{1,2})(?:\.(\d{1,3}))?\]`)
	for _, raw := range strings.Split(lrc, "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		// Skip metadata tags like [ti:], [ar:], [by:], [offset:]
		if strings.HasPrefix(raw, "[ti:") || strings.HasPrefix(raw, "[ar:") || strings.HasPrefix(raw, "[al:") || strings.HasPrefix(raw, "[by:") || strings.HasPrefix(raw, "[offset:") {
			continue
		}
		matches := re.FindAllStringSubmatchIndex(raw, -1)
		if len(matches) == 0 {
			continue
		}
		// Extract text after last timestamp tag
		last := matches[len(matches)-1]
		text := strings.TrimSpace(raw[last[1]:])
		if text == "" {
			continue
		}
		for _, m := range matches {
			mm := raw[m[0]:m[1]]
			parts := re.FindStringSubmatch(mm)
			if len(parts) >= 3 {
				min := atoiSafe(parts[1])
				sec := atoiSafe(parts[2])
				ms := 0
				if len(parts) >= 4 && parts[3] != "" {
					p := parts[3]
					if len(p) == 2 { // .xx -> .xx0
						p = p + "0"
					}
					if len(p) == 1 { // .x -> .x00
						p = p + "00"
					}
					ms = atoiSafe(p)
				}
				timestamp := int64(min*60*1000 + sec*1000 + ms)
				lines = append(lines, overlay.LyricsLine{Text: text, Timestamp: timestamp})
			}
		}
	}
	// Sort by timestamp
	sort.Slice(lines, func(i, j int) bool { return lines[i].Timestamp < lines[j].Timestamp })
	return lines
}

func atoiSafe(s string) int {
	res := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		res = res*10 + int(c-'0')
	}
	return res
}

// getByID fetches a single track with lyrics by LRCLIB ID
func (l *LRCLibProvider) getByID(id int) (*lrcLibTrack, error) {
	// Try REST style first: /get/{id}
	endpoint := fmt.Sprintf("%s/get/%d", l.baseURL, id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SpotLy/1.0")
	resp, err := l.client.Do(req)
	if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var track lrcLibTrack
		if err := json.Unmarshal(body, &track); err == nil {
			return &track, nil
		}
	}
	// Fallback to query param style: /get?id=123
	endpoint = fmt.Sprintf("%s/get?id=%d", l.baseURL, id)
	req, err = http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SpotLy/1.0")
	resp, err = l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib get status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var track lrcLibTrack
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, err
	}
	return &track, nil
}

// DemoProvider provides demo lyrics for any track
type DemoProvider struct{}

// NewDemoProvider creates a new demo provider
func NewDemoProvider() *DemoProvider {
	return &DemoProvider{}
}

// GetName returns the provider name
func (d *DemoProvider) GetName() string {
	return "Demo"
}

// SearchLyrics provides fallback when no other provider works
func (d *DemoProvider) SearchLyrics(artist, title string) (*overlay.LyricsData, error) {
	// Only provide basic track info, not full lyrics
	lyrics := &overlay.LyricsData{
		Source:    "Info",
		IsSynced:  false,
		FetchedAt: time.Now(),
		Lines: []overlay.LyricsLine{
			{Text: fmt.Sprintf("🎵 %s", title), Timestamp: 0},
			{Text: fmt.Sprintf("by %s", artist), Timestamp: 2000},
			{Text: "", Timestamp: 4000},
			{Text: "♪ Playing on Spotify ♪", Timestamp: 6000},
		},
	}

	return lyrics, nil
}
