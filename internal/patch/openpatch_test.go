package patch

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOpenPatch(t *testing.T) {
	data, err := os.ReadFile("../../testdata/gamelist_sample.txt")
	require.NoError(t, err)

	patches, err := ParseOpenPatch(string(data))
	require.NoError(t, err)
	require.Len(t, patches, 3)

	// First patch: ASCII arrow
	assert.Equal(t, uint32(0x00012FA8), patches[0].Offset)
	assert.Equal(t, []byte{0x64, 0x63, 0x6F, 0xF3}, patches[0].Original)
	assert.Equal(t, []byte{0x64, 0x64, 0x6F, 0xF3}, patches[0].Patched)

	// Second patch: Unicode arrow
	assert.Equal(t, uint32(0x000C3E20), patches[1].Offset)
	assert.Equal(t, []byte{0xAA, 0xBB, 0xCC, 0xDD}, patches[1].Original)
	assert.Equal(t, []byte{0xEE, 0xFF, 0x00, 0x11}, patches[1].Patched)

	// Third patch: single byte
	assert.Equal(t, uint32(0x00001000), patches[2].Offset)
	assert.Equal(t, []byte{0xFF}, patches[2].Original)
	assert.Equal(t, []byte{0x00}, patches[2].Patched)
}

func TestParseOpenPatchEmpty(t *testing.T) {
	patches, err := ParseOpenPatch("")
	require.NoError(t, err)
	assert.Empty(t, patches)
}

func TestParseOpenPatchCommentsOnly(t *testing.T) {
	text := "; comment\n# another comment\n\n"
	patches, err := ParseOpenPatch(text)
	require.NoError(t, err)
	assert.Empty(t, patches)
}

func TestApplyHexPatches(t *testing.T) {
	rom := make([]byte, 0x200)
	rom[0x100] = 0xAA
	rom[0x101] = 0xBB

	patches := []HexPatch{
		{Offset: 0x100, Original: []byte{0xAA, 0xBB}, Patched: []byte{0xCC, 0xDD}},
	}

	err := ApplyHexPatches(rom, patches, true)
	require.NoError(t, err)
	assert.Equal(t, byte(0xCC), rom[0x100])
	assert.Equal(t, byte(0xDD), rom[0x101])
}

func TestApplyHexPatchesVerifyFail(t *testing.T) {
	rom := make([]byte, 0x200)
	rom[0x100] = 0xFF // Wrong original byte.

	patches := []HexPatch{
		{Offset: 0x100, Original: []byte{0xAA}, Patched: []byte{0xBB}},
	}

	err := ApplyHexPatches(rom, patches, true)
	assert.ErrorIs(t, err, ErrVerifyFailed)
}

func TestApplyHexPatchesNoVerify(t *testing.T) {
	rom := make([]byte, 0x200)
	rom[0x100] = 0xFF // "Wrong" original, but verify=false.

	patches := []HexPatch{
		{Offset: 0x100, Original: []byte{0xAA}, Patched: []byte{0xBB}},
	}

	err := ApplyHexPatches(rom, patches, false)
	require.NoError(t, err)
	assert.Equal(t, byte(0xBB), rom[0x100])
}

func TestApplyHexPatchesOutOfBounds(t *testing.T) {
	rom := make([]byte, 0x10)
	patches := []HexPatch{
		{Offset: 0x0F, Original: []byte{0x00, 0x00}, Patched: []byte{0xAA, 0xBB}},
	}
	err := ApplyHexPatches(rom, patches, false)
	assert.Error(t, err)
}
