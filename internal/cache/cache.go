package cache

import (
	"container/list"
	"sync"
	"time"

	"lyrics-overlay/internal/overlay"
)

// Service implements an LRU cache for lyrics
type Service struct {
	mu          sync.RWMutex
	maxSize     int
	trackCache  map[string]*cacheEntry      // Cache by Spotify track ID
	keyCache    map[string]*cacheEntry      // Cache by normalized "artist|title"
	lruList     *list.List                  // LRU list for eviction
	trackToElem map[string]*list.Element    // Map track ID to list element
	keyToElem   map[string]*list.Element    // Map cache key to list element
}

// cacheEntry holds cached lyrics data with metadata
type cacheEntry struct {
	lyrics    *overlay.LyricsData
	trackID   string
	cacheKey  string
	timestamp time.Time
}

// New creates a new cache service
func New(maxSize int) *Service {
	if maxSize <= 0 {
		maxSize = 100 // Default cache size
	}

	return &Service{
		maxSize:     maxSize,
		trackCache:  make(map[string]*cacheEntry),
		keyCache:    make(map[string]*cacheEntry),
		lruList:     list.New(),
		trackToElem: make(map[string]*list.Element),
		keyToElem:   make(map[string]*list.Element),
	}
}

// GetByTrackID retrieves lyrics by Spotify track ID
func (s *Service) GetByTrackID(trackID string) *overlay.LyricsData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.trackCache[trackID]
	if !exists {
		return nil
	}

	// Check if entry is still valid (24 hours)
	if time.Since(entry.timestamp) > 24*time.Hour {
		// Entry is stale, remove it
		s.removeEntryUnsafe(entry)
		return nil
	}

	// Move to front of LRU list
	if elem, exists := s.trackToElem[trackID]; exists {
		s.lruList.MoveToFront(elem)
	}

	return entry.lyrics
}

// GetByKey retrieves lyrics by normalized cache key
func (s *Service) GetByKey(cacheKey string) *overlay.LyricsData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.keyCache[cacheKey]
	if !exists {
		return nil
	}

	// Check if entry is still valid (24 hours)
	if time.Since(entry.timestamp) > 24*time.Hour {
		// Entry is stale, remove it
		s.removeEntryUnsafe(entry)
		return nil
	}

	// Move to front of LRU list
	if elem, exists := s.keyToElem[cacheKey]; exists {
		s.lruList.MoveToFront(elem)
	}

	return entry.lyrics
}

// SetByTrackID caches lyrics by Spotify track ID
func (s *Service) SetByTrackID(trackID string, lyrics *overlay.LyricsData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	if existingEntry, exists := s.trackCache[trackID]; exists {
		// Update existing entry
		existingEntry.lyrics = lyrics
		existingEntry.timestamp = time.Now()
		
		// Move to front
		if elem, exists := s.trackToElem[trackID]; exists {
			s.lruList.MoveToFront(elem)
		}
		return
	}

	// Create new entry
	entry := &cacheEntry{
		lyrics:    lyrics,
		trackID:   trackID,
		timestamp: time.Now(),
	}

	// Add to cache maps
	s.trackCache[trackID] = entry

	// Add to LRU list
	elem := s.lruList.PushFront(entry)
	s.trackToElem[trackID] = elem

	// Enforce size limit
	s.enforceMaxSize()
}

// SetByKey caches lyrics by normalized cache key
func (s *Service) SetByKey(cacheKey string, lyrics *overlay.LyricsData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	if existingEntry, exists := s.keyCache[cacheKey]; exists {
		// Update existing entry
		existingEntry.lyrics = lyrics
		existingEntry.timestamp = time.Now()
		
		// Move to front
		if elem, exists := s.keyToElem[cacheKey]; exists {
			s.lruList.MoveToFront(elem)
		}
		return
	}

	// Create new entry
	entry := &cacheEntry{
		lyrics:    lyrics,
		cacheKey:  cacheKey,
		timestamp: time.Now(),
	}

	// Add to cache maps
	s.keyCache[cacheKey] = entry

	// Add to LRU list
	elem := s.lruList.PushFront(entry)
	s.keyToElem[cacheKey] = elem

	// Enforce size limit
	s.enforceMaxSize()
}

// enforceMaxSize removes old entries if cache exceeds max size
func (s *Service) enforceMaxSize() {
	for s.lruList.Len() > s.maxSize {
		// Remove least recently used entry
		elem := s.lruList.Back()
		if elem != nil {
			entry := elem.Value.(*cacheEntry)
			s.removeEntryUnsafe(entry)
		}
	}
}

// removeEntryUnsafe removes an entry from all cache structures (must hold write lock)
func (s *Service) removeEntryUnsafe(entry *cacheEntry) {
	// Remove from track cache
	if entry.trackID != "" {
		delete(s.trackCache, entry.trackID)
		if elem, exists := s.trackToElem[entry.trackID]; exists {
			s.lruList.Remove(elem)
			delete(s.trackToElem, entry.trackID)
		}
	}

	// Remove from key cache
	if entry.cacheKey != "" {
		delete(s.keyCache, entry.cacheKey)
		if elem, exists := s.keyToElem[entry.cacheKey]; exists {
			s.lruList.Remove(elem)
			delete(s.keyToElem, entry.cacheKey)
		}
	}
}

// Clear removes all entries from the cache
func (s *Service) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.trackCache = make(map[string]*cacheEntry)
	s.keyCache = make(map[string]*cacheEntry)
	s.lruList = list.New()
	s.trackToElem = make(map[string]*list.Element)
	s.keyToElem = make(map[string]*list.Element)
}

// Size returns the current cache size
func (s *Service) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lruList.Len()
}

// Stats returns cache statistics
func (s *Service) Stats() CacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return CacheStats{
		Size:         s.lruList.Len(),
		MaxSize:      s.maxSize,
		TrackEntries: len(s.trackCache),
		KeyEntries:   len(s.keyCache),
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size         int `json:"size"`
	MaxSize      int `json:"max_size"`
	TrackEntries int `json:"track_entries"`
	KeyEntries   int `json:"key_entries"`
}
