package apdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDB(t *testing.T) *Database {
	t.Helper()
	db, err := NewDatabase()
	require.NoError(t, err)
	return db
}

func TestNewDatabase(t *testing.T) {
	db := testDB(t)
	assert.NotEmpty(t, db.index.Entries)
	assert.Equal(t, 1, db.index.Version)
}

func TestNewDatabaseInvalidJSON(t *testing.T) {
	_, err := NewDatabaseFromData([]byte("not json"))
	assert.Error(t, err)
}

func TestLookupExact(t *testing.T) {
	db := testDB(t)
	entry, err := db.Lookup("IPKE", "FFFF")
	require.NoError(t, err)
	assert.Equal(t, "IPKE", entry.GameCode)
	assert.Contains(t, entry.Title, "Pokemon")
}

func TestLookupFallback(t *testing.T) {
	db := testDB(t)
	// Non-matching CRC should fallback to FFFF entry.
	entry, err := db.Lookup("IPKE", "1234")
	require.NoError(t, err)
	assert.Equal(t, "IPKE", entry.GameCode)
	assert.Equal(t, "FFFF", entry.HeaderCRC)
}

func TestLookupNotFound(t *testing.T) {
	db := testDB(t)
	_, err := db.Lookup("ZZZZ", "0000")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestLookupCaseInsensitive(t *testing.T) {
	db := testDB(t)
	entry, err := db.Lookup("ipke", "ffff")
	require.NoError(t, err)
	assert.Equal(t, "IPKE", entry.GameCode)
}

func TestSearch(t *testing.T) {
	db := testDB(t)

	results := db.Search("pokemon")
	assert.Len(t, results, 2) // SoulSilver + HeartGold

	results = db.Search("mario")
	assert.Len(t, results, 2) // Bowser's Inside Story + Partners in Time

	results = db.Search("nonexistent")
	assert.Empty(t, results)
}

func TestSearchByCode(t *testing.T) {
	db := testDB(t)
	results := db.Search("B6RE")
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "Mega Man")
}

func TestEntries(t *testing.T) {
	db := testDB(t)
	assert.Len(t, db.Entries(), 5)
}
