package nds

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMinimalHeader() []byte {
	raw := make([]byte, HeaderSize)
	// Game title: "TESTGAME"
	copy(raw[0x000:], "TESTGAME\x00\x00\x00\x00")
	// Game code: "IPKE"
	copy(raw[0x00C:], "IPKE")
	// Maker code: "01"
	copy(raw[0x010:], "01")
	// Compute and store CRC
	crc := ComputeHeaderCRC(raw)
	binary.LittleEndian.PutUint16(raw[0x15E:], crc)
	return raw
}

func TestParseHeader(t *testing.T) {
	raw := makeMinimalHeader()

	h, rawOut, err := ParseHeader(bytes.NewReader(raw))
	require.NoError(t, err)
	assert.Equal(t, raw, rawOut)
	assert.Equal(t, "TESTGAME", h.Title())
	assert.Equal(t, "IPKE", h.Code())
	assert.Equal(t, "01", h.Maker())
}

func TestParseHeaderShortInput(t *testing.T) {
	_, _, err := ParseHeader(bytes.NewReader([]byte{1, 2, 3}))
	assert.Error(t, err)
}

func TestAPPatchID(t *testing.T) {
	raw := makeMinimalHeader()
	h, _, err := ParseHeader(bytes.NewReader(raw))
	require.NoError(t, err)

	id := h.APPatchID(raw)
	assert.Regexp(t, `^IPKE-[0-9A-F]{4}$`, id)
}

func TestValidateCRC(t *testing.T) {
	raw := makeMinimalHeader()
	h, _, err := ParseHeader(bytes.NewReader(raw))
	require.NoError(t, err)

	assert.True(t, h.ValidateCRC(raw))

	// Corrupt a byte and re-parse.
	raw[0x10] = 0xFF
	h2, _, err := ParseHeader(bytes.NewReader(raw))
	require.NoError(t, err)
	assert.False(t, h2.ValidateCRC(raw))
}
