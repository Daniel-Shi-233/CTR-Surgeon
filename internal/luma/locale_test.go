package luma

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateLocale(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateLocale(tmpDir, "0004000000055D00", "JPN", "JP")
	require.NoError(t, err)

	path := filepath.Join(tmpDir, "luma", "titles", "0004000000055D00", "locale.txt")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "JPN JP\n", string(data))
}

func TestGenerateLocaleInvalidTitleID(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateLocale(tmpDir, "short", "JPN", "JP")
	assert.ErrorIs(t, err, ErrInvalidTitleID)

	err = GenerateLocale(tmpDir, "000400000055D00G", "JPN", "JP")
	assert.ErrorIs(t, err, ErrInvalidTitleID)
}

func TestGenerateLocaleInvalidRegion(t *testing.T) {
	tmpDir := t.TempDir()
	err := GenerateLocale(tmpDir, "0004000000055D00", "XXX", "JP")
	assert.ErrorIs(t, err, ErrInvalidRegion)
}

func TestGenerateLocaleInvalidLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	err := GenerateLocale(tmpDir, "0004000000055D00", "JPN", "XX")
	assert.ErrorIs(t, err, ErrInvalidLang)
}

func TestGenerateLocaleCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	err := GenerateLocale(tmpDir, "0004000000055d00", "jpn", "jp")
	require.NoError(t, err)

	path := filepath.Join(tmpDir, "luma", "titles", "0004000000055D00", "locale.txt")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "JPN JP\n", string(data))
}

func TestListLocales(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, GenerateLocale(tmpDir, "0004000000055D00", "JPN", "JP"))
	require.NoError(t, GenerateLocale(tmpDir, "00040000001B5100", "USA", "EN"))

	locales, err := ListLocales(tmpDir)
	require.NoError(t, err)
	assert.Len(t, locales, 2)
	assert.Equal(t, "JPN JP", locales["0004000000055D00"])
	assert.Equal(t, "USA EN", locales["00040000001B5100"])
}

func TestListLocalesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "luma", "titles"), 0755)

	locales, err := ListLocales(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, locales)
}
