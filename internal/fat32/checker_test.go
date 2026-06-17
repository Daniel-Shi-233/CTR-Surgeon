package fat32

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFakeSD(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Create critical directories.
	os.MkdirAll(filepath.Join(root, "Nintendo 3DS"), 0755)
	os.MkdirAll(filepath.Join(root, "luma"), 0755)
	os.MkdirAll(filepath.Join(root, "_nds"), 0755)

	// Create a critical file.
	os.WriteFile(filepath.Join(root, "boot.firm"), []byte("FIRM"), 0644)

	return root
}

func TestCheckSD(t *testing.T) {
	root := setupFakeSD(t)

	report, err := CheckSD(root)
	require.NoError(t, err)
	assert.Equal(t, root, report.Path)
	assert.NotEmpty(t, report.Checks)

	// At least some checks should pass.
	assert.Greater(t, report.PassCount(), 0)
}

func TestCheckSDFullyPopulated(t *testing.T) {
	root := setupFakeSD(t)
	os.MkdirAll(filepath.Join(root, "__rpg"), 0755)
	os.WriteFile(filepath.Join(root, "boot.3dsx"), []byte("3DSX"), 0644)

	report, err := CheckSD(root)
	require.NoError(t, err)

	// All directory and file checks should pass (6 items).
	dirFileChecks := 0
	passed := 0
	for _, c := range report.Checks {
		if c.Name != "Free space" {
			dirFileChecks++
			if c.OK {
				passed++
			}
		}
	}
	assert.Equal(t, 6, dirFileChecks)
	assert.Equal(t, 6, passed)
}

func TestCheckSDEmpty(t *testing.T) {
	root := t.TempDir()

	report, err := CheckSD(root)
	require.NoError(t, err)

	// All directory/file checks should fail on empty dir.
	for _, c := range report.Checks {
		if c.Name != "Free space" {
			assert.False(t, c.OK, "expected %s to fail", c.Name)
		}
	}
}

func TestCheckSDNotExist(t *testing.T) {
	_, err := CheckSD("/nonexistent/path/nowhere")
	assert.Error(t, err)
}

func TestCheckSDNotDir(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	os.WriteFile(tmpFile, []byte("hi"), 0644)

	_, err := CheckSD(tmpFile)
	assert.Error(t, err)
}

func TestPassCount(t *testing.T) {
	report := &SDReport{
		Checks: []CheckResult{
			{OK: true},
			{OK: false},
			{OK: true},
		},
	}
	assert.Equal(t, 2, report.PassCount())
}

func TestDetectSDCards(t *testing.T) {
	// Just make sure it doesn't panic.
	cards := DetectSDCards()
	_ = cards
}
