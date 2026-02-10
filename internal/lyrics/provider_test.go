package lyrics

import (
	"testing"
)

// --- normalizeForCache ---

func TestNormalizeForCache(t *testing.T) {
	tests := []struct {
		artist, title string
		want          string
	}{
		{"Taylor Swift", "Love Story", "taylor swift|love story"},
		{"The Weeknd", "Blinding Lights (Radio Edit)", "the weeknd|blinding lights"},
		{"Ed Sheeran", "Shape of You (feat. Someone)", "ed sheeran|shape of you"},
	}

	for _, tc := range tests {
		got := normalizeForCache(tc.artist, tc.title)
		if got != tc.want {
			t.Errorf("normalizeForCache(%q, %q) = %q; want %q", tc.artist, tc.title, got, tc.want)
		}
	}
}

// --- normalizeString ---

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello world"},
		{"Song (feat. Artist)", "song"},
		{"Track [Remastered 2024]", "track"},
		{"Title - Radio Edit", "title"},
		{"Track (Remix)", "track"},
		{"Song - Remaster 2020", "song"},
		{"  Extra   Spaces  ", "extra spaces"},
		{"Normal Song!!!", "normal song"},
		{"Track (featuring Artist)", "track"},
		{"Song (ft. DJ)", "song"},
		{"Title - Extended Version", "title"},
	}

	for _, tc := range tests {
		got := normalizeString(tc.input)
		if got != tc.want {
			t.Errorf("normalizeString(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

// --- textToLyricsLines ---

func TestTextToLyricsLines_Basic(t *testing.T) {
	text := "First line\nSecond line\nThird line"
	lines := textToLyricsLines(text)

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0].Text != "First line" {
		t.Errorf("Expected 'First line', got %q", lines[0].Text)
	}
	if lines[2].Text != "Third line" {
		t.Errorf("Expected 'Third line', got %q", lines[2].Text)
	}
}

func TestTextToLyricsLines_FilterNoise(t *testing.T) {
	text := "Real lyrics here\nYou might also like\nMore real lyrics\n123Embed\nGenius annotation stuff\nSee the full lyrics\nFinal line"
	lines := textToLyricsLines(text)

	// Should filter out noise lines
	for _, line := range lines {
		if line.Text == "You might also like" {
			t.Error("Should have filtered 'You might also like'")
		}
		if line.Text == "123Embed" {
			t.Error("Should have filtered '123Embed'")
		}
	}

	// Should keep real lyrics
	found := false
	for _, line := range lines {
		if line.Text == "Real lyrics here" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Real lyrics here' to be kept")
	}
}

func TestTextToLyricsLines_DeduplicateEmptyLines(t *testing.T) {
	text := "Verse one\n\n\n\nVerse two"
	lines := textToLyricsLines(text)

	emptyCount := 0
	for _, line := range lines {
		if line.Text == "" {
			emptyCount++
		}
	}
	if emptyCount > 1 {
		t.Errorf("Expected at most 1 empty line (dedup), got %d", emptyCount)
	}
}

func TestTextToLyricsLines_TrimLeadingTrailingEmpty(t *testing.T) {
	text := "\n\nActual lyrics\n\n"
	lines := textToLyricsLines(text)

	if len(lines) == 0 {
		t.Fatal("Expected at least 1 line")
	}
	if lines[0].Text == "" {
		t.Error("Leading empty lines should be trimmed")
	}
	if lines[len(lines)-1].Text == "" {
		t.Error("Trailing empty lines should be trimmed")
	}
}

func TestTextToLyricsLines_FilterLanguageNames(t *testing.T) {
	text := "Real lyrics\nDeutsch\nFrançais\nMore lyrics\nEspañol"
	lines := textToLyricsLines(text)

	for _, line := range lines {
		lower := line.Text
		if lower == "Deutsch" || lower == "Français" || lower == "Español" {
			t.Errorf("Should have filtered language name %q", line.Text)
		}
	}
}

func TestTextToLyricsLines_FilterContributors(t *testing.T) {
	text := "Real lyrics\n5 contributors\ntranslation available\nMore lyrics"
	lines := textToLyricsLines(text)

	for _, line := range lines {
		if line.Text == "5 contributors" || line.Text == "translation available" {
			t.Errorf("Should have filtered %q", line.Text)
		}
	}
}

// --- atoiSafe ---

func TestAtoiSafe(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"123", 123},
		{"0", 0},
		{"42", 42},
		{"1a2b3", 123}, // non-digits skipped
		{"", 0},
		{"abc", 0},
	}

	for _, tc := range tests {
		got := atoiSafe(tc.input)
		if got != tc.want {
			t.Errorf("atoiSafe(%q) = %d; want %d", tc.input, got, tc.want)
		}
	}
}

// --- parseLRCToLines ---

func TestParseLRCToLines_Basic(t *testing.T) {
	lrc := "[00:05.000]First line\n[00:10.500]Second line\n[00:15.200]Third line"
	lines := parseLRCToLines(lrc)

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0].Timestamp != 5000 {
		t.Errorf("Expected timestamp 5000, got %d", lines[0].Timestamp)
	}
	if lines[0].Text != "First line" {
		t.Errorf("Expected 'First line', got %q", lines[0].Text)
	}
	if lines[1].Timestamp != 10500 {
		t.Errorf("Expected timestamp 10500, got %d", lines[1].Timestamp)
	}
}

func TestParseLRCToLines_SkipsMetadata(t *testing.T) {
	lrc := "[ti:Test Song]\n[ar:Test Artist]\n[by:Author]\n[00:05.000]Actual lyrics"
	lines := parseLRCToLines(lrc)

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line (metadata skipped), got %d", len(lines))
	}
	if lines[0].Text != "Actual lyrics" {
		t.Errorf("Expected 'Actual lyrics', got %q", lines[0].Text)
	}
}

func TestParseLRCToLines_SortsByTimestamp(t *testing.T) {
	lrc := "[00:20.000]Third\n[00:05.000]First\n[00:12.000]Second"
	lines := parseLRCToLines(lrc)

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0].Timestamp != 5000 {
		t.Errorf("Expected first sorted timestamp 5000, got %d", lines[0].Timestamp)
	}
	if lines[2].Timestamp != 20000 {
		t.Errorf("Expected last sorted timestamp 20000, got %d", lines[2].Timestamp)
	}
}

func TestParseLRCToLines_MultipleTimestamps(t *testing.T) {
	lrc := "[00:05.000][00:30.000]Repeated line"
	lines := parseLRCToLines(lrc)

	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines (one per timestamp), got %d", len(lines))
	}
	if lines[0].Timestamp != 5000 || lines[1].Timestamp != 30000 {
		t.Errorf("Unexpected timestamps: %d, %d", lines[0].Timestamp, lines[1].Timestamp)
	}
}

func TestParseLRCToLines_TwoDigitMilliseconds(t *testing.T) {
	lrc := "[01:23.45]Line with 2-digit ms"
	lines := parseLRCToLines(lrc)

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	// 01:23.45 -> 1*60*1000 + 23*1000 + 450 = 83450
	if lines[0].Timestamp != 83450 {
		t.Errorf("Expected timestamp 83450, got %d", lines[0].Timestamp)
	}
}

func TestParseLRCToLines_SkipsEmptyText(t *testing.T) {
	lrc := "[00:05.000]\n[00:10.000]Real lyrics"
	lines := parseLRCToLines(lrc)

	// Empty text after timestamp should be skipped
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line (empty text skipped), got %d", len(lines))
	}
	if lines[0].Text != "Real lyrics" {
		t.Errorf("Expected 'Real lyrics', got %q", lines[0].Text)
	}
}
