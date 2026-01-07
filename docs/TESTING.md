# Testing

Test coverage and running tests for SpotLy.

---

## Running Tests

```bash
# Run all tests
go test -v ./internal/...

# Run with race detector
go test -v -race ./internal/...

# Run with coverage
go test -v -coverprofile=coverage.out ./internal/...

# View coverage report
go tool cover -html=coverage.out
```

---

## Test Files

| Package | Test File | Coverage |
|---------|-----------|----------|
| `internal/cache` | `cache_test.go` | Good |
| `internal/config` | `config_test.go` | Good |
| `internal/lyrics` | `lrclib_test.go` | Partial |

---

## Cache Tests

Location: `internal/cache/cache_test.go`

| Test | Description |
|------|-------------|
| `TestService_SetAndGet` | Basic set/get by track ID |
| `TestService_Eviction` | LRU eviction when full |
| `TestService_UpdateExisting` | Update existing entry |
| `TestService_GetByKey` | Get by normalized key |
| `TestService_Size` | Size tracking |
| `TestService_Clear` | Clear all entries |
| `TestService_Expiration` | TTL expiration (basic check) |
| `TestService_Stats` | Statistics reporting |

---

## Config Tests

Location: `internal/config/config_test.go`

| Test | Description |
|------|-------------|
| `TestLoadConfig_Default` | Load default config |
| `TestConfig_Save` | Save to file |
| `TestConfig_Load` | Load from file |
| `TestConfig_UpdateOverlay` | Update overlay settings |
| `TestConfig_UpdateAuth` | Update auth tokens |
| `TestGetDefaultConfig` | Default values |

---

## Lyrics Tests

Location: `internal/lyrics/lrclib_test.go`

| Test | Description |
|------|-------------|
| `TestParseSyncedLyrics` | Basic LRC parsing |
| `TestNormalizeTitle` | Title normalization |
| `TestParseSyncedLyrics_WithMetadata` | Skip metadata tags |
| `TestParseSyncedLyrics_MultipleTimestamps` | Multiple timestamps per line |
| `TestParseSyncedLyrics_Sorted` | Auto-sort by timestamp |
| `TestNormalizeTitle_Complex` | Complex title patterns |
| `TestLRCLibProvider_GetName` | Provider name |
| `TestDemoProvider_GetName` | Demo provider name |

---

## Coverage Gaps

Not currently tested:

| Component | Reason |
|-----------|--------|
| `internal/auth` | Requires mocking HTTP/OAuth flow |
| `internal/overlay` | Mostly state management |
| `internal/spotify` | Requires mocking Spotify API |
| `main.go` | Integration/E2E testing needed |
| Frontend | Would need browser testing |

---

## CI Pipeline

Tests run automatically via GitHub Actions:

```yaml
# .github/workflows/ci.yml
- name: Run tests
  run: go test -v -race -coverprofile=coverage.out ./internal/...

- name: Upload coverage
  uses: codecov/codecov-action@v4
```

---

## Adding Tests

When adding new functionality:

1. Create `*_test.go` file in same package
2. Use `testing` package
3. Follow existing patterns
4. Run tests before committing

Example:

```go
func TestNewFeature(t *testing.T) {
    // Arrange
    input := "test"
    
    // Act
    result := NewFeature(input)
    
    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

---

## See Also

- [Development](DEVELOPMENT.md) - Build and run instructions
- [Backend Overview](backend/README.md) - Package structure
