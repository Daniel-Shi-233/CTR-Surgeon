package luma

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestFont(t *testing.T, dir string, size int) string {
	t.Helper()
	path := filepath.Join(dir, "test.bcfnt")
	data := make([]byte, size)
	copy(data, "CFNT") // Magic header
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

func TestValidateFont(t *testing.T) {
	tmpDir := t.TempDir()
	fontPath := createTestFont(t, tmpDir, 1024)

	err := ValidateFont(fontPath)
	assert.NoError(t, err)
}

func TestValidateFontInvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.bcfnt")
	require.NoError(t, os.WriteFile(path, []byte("NOPE1234"), 0644))

	err := ValidateFont(path)
	assert.ErrorIs(t, err, ErrInvalidFont)
}

func TestValidateFontTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	fontPath := createTestFont(t, tmpDir, maxFontSize+1)

	err := ValidateFont(fontPath)
	assert.ErrorIs(t, err, ErrFontTooLarge)
}

func TestValidateFontNotExist(t *testing.T) {
	err := ValidateFont("/nonexistent/font.bcfnt")
	assert.Error(t, err)
}

func TestInjectFont(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := t.TempDir()
	fontPath := createTestFont(t, srcDir, 512)

	err := InjectFont(tmpDir, fontPath)
	require.NoError(t, err)

	dstPath := filepath.Join(tmpDir, "luma", "font.bcfnt")
	data, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, 512, len(data))
	assert.Equal(t, "CFNT", string(data[:4]))
}

func TestInjectFontInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	badPath := filepath.Join(t.TempDir(), "bad.bcfnt")
	require.NoError(t, os.WriteFile(badPath, []byte("NOPE"), 0644))

	err := InjectFont(tmpDir, badPath)
	assert.ErrorIs(t, err, ErrInvalidFont)
}

func TestCheckFontNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	installed, _, err := CheckFont(tmpDir)
	assert.False(t, installed)
	assert.NoError(t, err)
}

func TestCheckFontInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := t.TempDir()
	fontPath := createTestFont(t, srcDir, 256)
	require.NoError(t, InjectFont(tmpDir, fontPath))

	installed, path, _ := CheckFont(tmpDir)
	assert.True(t, installed)
	assert.Contains(t, path, "font.bcfnt")
}
