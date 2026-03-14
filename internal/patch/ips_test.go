package patch

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIPS(t *testing.T) {
	data, err := os.ReadFile("../../testdata/test.ips")
	require.NoError(t, err)

	patch, err := ParseIPS(data)
	require.NoError(t, err)
	require.Len(t, patch.Records, 2)

	// Normal record.
	assert.Equal(t, uint32(0x100), patch.Records[0].Offset)
	assert.Equal(t, []byte{0xAA, 0xBB, 0xCC}, patch.Records[0].Data)
	assert.False(t, patch.Records[0].IsRLE)

	// RLE record.
	assert.Equal(t, uint32(0x200), patch.Records[1].Offset)
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF}, patch.Records[1].Data)
	assert.True(t, patch.Records[1].IsRLE)
}

func TestParseIPSInvalid(t *testing.T) {
	_, err := ParseIPS([]byte("NOPE"))
	assert.ErrorIs(t, err, ErrInvalidIPS)

	_, err = ParseIPS([]byte("PATC"))
	assert.ErrorIs(t, err, ErrInvalidIPS)
}

func TestParseIPSTruncated(t *testing.T) {
	// PATCH header + partial record.
	data := []byte("PATCH\x00\x01\x00")
	_, err := ParseIPS(data)
	assert.ErrorIs(t, err, ErrTruncatedIPS)
}

func TestIPSApply(t *testing.T) {
	data, err := os.ReadFile("../../testdata/test.ips")
	require.NoError(t, err)

	patch, err := ParseIPS(data)
	require.NoError(t, err)

	// Create a ROM large enough.
	rom := make([]byte, 0x300)

	err = patch.Apply(rom)
	require.NoError(t, err)

	// Verify normal record.
	assert.Equal(t, byte(0xAA), rom[0x100])
	assert.Equal(t, byte(0xBB), rom[0x101])
	assert.Equal(t, byte(0xCC), rom[0x102])

	// Verify RLE record.
	for i := 0; i < 4; i++ {
		assert.Equal(t, byte(0xFF), rom[0x200+i])
	}

	// Bytes outside patches should be zero.
	assert.Equal(t, byte(0x00), rom[0x103])
	assert.Equal(t, byte(0x00), rom[0x204])
}

func TestIPSApplyOutOfBounds(t *testing.T) {
	data, err := os.ReadFile("../../testdata/test.ips")
	require.NoError(t, err)

	patch, err := ParseIPS(data)
	require.NoError(t, err)

	// ROM too small for the second record (offset 0x200).
	rom := make([]byte, 0x100)
	err = patch.Apply(rom)
	assert.ErrorIs(t, err, ErrOffsetTooLarge)
}

func TestParseIPSEmptyPatch(t *testing.T) {
	// Just PATCH + EOF, no records.
	data := []byte("PATCHEOF")
	patch, err := ParseIPS(data)
	require.NoError(t, err)
	assert.Empty(t, patch.Records)
}
