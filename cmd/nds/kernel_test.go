package nds

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeZip writes a zip at path containing the given name->content entries.
func makeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range entries {
		fw, err := w.Create(name)
		require.NoError(t, err)
		_, err = fw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
}

func TestExtractZipLimited_Normal(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "in.zip")
	dest := filepath.Join(tmp, "out")
	makeZip(t, zipPath, map[string]string{
		"_DSMENU.DAT":   "menu",
		"_nds/file.bin": "payload",
	})

	require.NoError(t, extractZipLimited(zipPath, dest, maxKernelFileSize))

	got, err := os.ReadFile(filepath.Join(dest, "_DSMENU.DAT"))
	require.NoError(t, err)
	assert.Equal(t, "menu", string(got))

	got, err = os.ReadFile(filepath.Join(dest, "_nds", "file.bin"))
	require.NoError(t, err)
	assert.Equal(t, "payload", string(got))
}

func TestExtractZipLimited_RejectsZipSlip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "evil.zip")
	dest := filepath.Join(tmp, "out")
	makeZip(t, zipPath, map[string]string{
		"../escaped.txt": "pwned",
		"safe.txt":       "ok",
	})

	require.NoError(t, extractZipLimited(zipPath, dest, maxKernelFileSize))

	// The traversal entry must NOT be written outside dest.
	_, err := os.Stat(filepath.Join(tmp, "escaped.txt"))
	assert.True(t, os.IsNotExist(err), "zip-slip entry escaped the destination dir")

	// The safe entry is still extracted.
	got, err := os.ReadFile(filepath.Join(dest, "safe.txt"))
	require.NoError(t, err)
	assert.Equal(t, "ok", string(got))
}

func TestExtractZipLimited_RejectsOversizeEntry(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "big.zip")
	dest := filepath.Join(tmp, "out")
	makeZip(t, zipPath, map[string]string{
		"big.bin": "0123456789", // 10 bytes
	})

	err := extractZipLimited(zipPath, dest, 4) // limit smaller than entry
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds size limit")

	// Partial output must be cleaned up, not left behind.
	_, statErr := os.Stat(filepath.Join(dest, "big.bin"))
	assert.True(t, os.IsNotExist(statErr), "oversize entry left a partial file behind")
}
