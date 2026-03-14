package apdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotFound     = errors.New("AP patch not found in database")
	ErrFetchFailed  = errors.New("failed to fetch AP patch")
)

// APEntry represents one entry in the AP patch index.
type APEntry struct {
	GameCode  string `json:"game_code"`
	HeaderCRC string `json:"header_crc"`
	File      string `json:"file"`
	Title     string `json:"title"`
	Notes     string `json:"notes"`
}

// Index is the top-level structure of the embedded index.json.
type Index struct {
	Version int       `json:"version"`
	BaseURL string    `json:"base_url"`
	Entries []APEntry `json:"entries"`
}

// Database provides AP patch lookup and download.
type Database struct {
	index    Index
	cacheDir string
}

// NewDatabase loads the embedded index and sets up the cache directory.
func NewDatabase() (*Database, error) {
	return NewDatabaseFromData(embeddedIndex)
}

// NewDatabaseFromData creates a Database from raw JSON index data.
func NewDatabaseFromData(indexData []byte) (*Database, error) {
	var idx Index
	if err := json.Unmarshal(indexData, &idx); err != nil {
		return nil, fmt.Errorf("parsing AP index: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	cacheDir := filepath.Join(homeDir, ".ctr-surgeon", "patches")

	return &Database{
		index:    idx,
		cacheDir: cacheDir,
	}, nil
}

// Lookup finds an AP entry by game code and header CRC.
// Strategy: exact match first (XXXX-YYYY), then fallback (XXXX-FFFF).
func (db *Database) Lookup(gameCode, headerCRC string) (*APEntry, error) {
	gameCode = strings.ToUpper(gameCode)
	headerCRC = strings.ToUpper(headerCRC)

	// Exact match.
	for i := range db.index.Entries {
		e := &db.index.Entries[i]
		if strings.ToUpper(e.GameCode) == gameCode && strings.ToUpper(e.HeaderCRC) == headerCRC {
			return e, nil
		}
	}

	// Fallback: FFFF wildcard.
	for i := range db.index.Entries {
		e := &db.index.Entries[i]
		if strings.ToUpper(e.GameCode) == gameCode && strings.ToUpper(e.HeaderCRC) == "FFFF" {
			return e, nil
		}
	}

	return nil, fmt.Errorf("%w: %s-%s", ErrNotFound, gameCode, headerCRC)
}

// Search returns all entries whose title contains the query (case-insensitive).
func (db *Database) Search(query string) []APEntry {
	query = strings.ToLower(query)
	var results []APEntry
	for _, e := range db.index.Entries {
		if strings.Contains(strings.ToLower(e.Title), query) ||
			strings.Contains(strings.ToLower(e.GameCode), query) {
			results = append(results, e)
		}
	}
	return results
}

// Entries returns all entries in the database.
func (db *Database) Entries() []APEntry {
	return db.index.Entries
}

// FetchPatch downloads or returns cached IPS patch data for an entry.
func (db *Database) FetchPatch(ctx context.Context, entry *APEntry) ([]byte, error) {
	// Check cache first.
	cachePath := filepath.Join(db.cacheDir, entry.File)
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	// Download from base URL.
	url := db.index.BaseURL + entry.File
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d for %s", ErrFetchFailed, resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: reading response: %v", ErrFetchFailed, err)
	}

	// Cache for next time.
	if err := os.MkdirAll(db.cacheDir, 0755); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}

	return data, nil
}

// CacheDir returns the current cache directory path.
func (db *Database) CacheDir() string {
	return db.cacheDir
}

// SetCacheDir overrides the default cache directory.
func (db *Database) SetCacheDir(dir string) {
	db.cacheDir = dir
}
