package lyrics

import (
	"testing"
)

func TestParseSyncedLyrics(t *testing.T) {
	raw := `[00:12.34]First line
[00:15.67]Second line
[00:20.00]Third line`

	lines := ParseSyncedLyrics(raw)

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	tests := []struct {
		idx      int
		wantMs   int64
		wantText string
	}{
		{0, 12340, "First line"},
		{1, 15670, "Second line"},
		{2, 20000, "Third line"},
	}

	for _, tc := range tests {
		if lines[tc.idx].Timestamp != tc.wantMs {
			t.Errorf("Line %d time = %d; want %d", tc.idx, lines[tc.idx].Timestamp, tc.wantMs)
		}
		if lines[tc.idx].Text != tc.wantText {
			t.Errorf("Line %d text = %q; want %q", tc.idx, lines[tc.idx].Text, tc.wantText)
		}
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Song (feat. Artist)", "song"},
		{"Track [Remastered 2024]", "track"},
		{"Title - Radio Edit", "title"},
		{"Song (Remix)", "song"},
		{"Track - Remaster", "track"},
	}

	for _, tc := range tests {
		got := NormalizeTitle(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeTitle(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseSyncedLyrics_WithMetadata(t *testing.T) {
	raw := `[ti:Test Song]
[ar:Test Artist]
[00:10.00]First line
[00:15.50]Second line
[by:Test Author]`

	lines := ParseSyncedLyrics(raw)

	// Metadata lines should be skipped
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines (metadata skipped), got %d", len(lines))
	}

	if lines[0].Text != "First line" {
		t.Errorf("Expected first line 'First line', got %q", lines[0].Text)
	}
	if lines[0].Timestamp != 10000 {
		t.Errorf("Expected first line timestamp 10000, got %d", lines[0].Timestamp)
	}
}

func TestParseSyncedLyrics_MultipleTimestamps(t *testing.T) {
	raw := `[00:10.00][00:12.00]Line with multiple timestamps`

	lines := ParseSyncedLyrics(raw)

	// Should create one line per timestamp
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines (one per timestamp), got %d", len(lines))
	}

	if lines[0].Timestamp != 10000 {
		t.Errorf("Expected first timestamp 10000, got %d", lines[0].Timestamp)
	}
	if lines[1].Timestamp != 12000 {
		t.Errorf("Expected second timestamp 12000, got %d", lines[1].Timestamp)
	}
	if lines[0].Text != "Line with multiple timestamps" {
		t.Errorf("Expected text 'Line with multiple timestamps', got %q", lines[0].Text)
	}
}

func TestParseSyncedLyrics_Sorted(t *testing.T) {
	raw := `[00:20.00]Third line
[00:10.00]First line
[00:15.00]Second line`

	lines := ParseSyncedLyrics(raw)

	// Should be sorted by timestamp
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0].Timestamp != 10000 {
		t.Errorf("Expected first line timestamp 10000, got %d", lines[0].Timestamp)
	}
	if lines[1].Timestamp != 15000 {
		t.Errorf("Expected second line timestamp 15000, got %d", lines[1].Timestamp)
	}
	if lines[2].Timestamp != 20000 {
		t.Errorf("Expected third line timestamp 20000, got %d", lines[2].Timestamp)
	}
}

func TestNormalizeTitle_Complex(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Song (feat. Artist) [Remastered]", "song"},
		{"Track - Radio Edit (2024 Remix)", "track"},
		{"Title (ft. Someone) - Extended Version", "title"},
		{"Normal Song", "normal song"},
		{"  Extra   Spaces  ", "extra spaces"},
		{"Song!!!", "song"},
		{"Track (featuring Artist)", "track"},
	}

	for _, tc := range tests {
		got := NormalizeTitle(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeTitle(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestLRCLibProvider_GetName(t *testing.T) {
	provider := NewLRCLibProvider(nil)
	if provider.GetName() != "LRCLIB" {
		t.Errorf("Expected provider name 'LRCLIB', got %q", provider.GetName())
	}
}

func TestDemoProvider_GetName(t *testing.T) {
	provider := NewDemoProvider()
	if provider.GetName() != "Demo" {
		t.Errorf("Expected provider name 'Demo', got %q", provider.GetName())
	}
}
