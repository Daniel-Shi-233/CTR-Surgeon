package nds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeHeaderCRC(t *testing.T) {
	t.Run("short input returns zero", func(t *testing.T) {
		assert.Equal(t, uint16(0), ComputeHeaderCRC([]byte{1, 2, 3}))
	})

	t.Run("all zeros", func(t *testing.T) {
		raw := make([]byte, HeaderSize)
		crc := ComputeHeaderCRC(raw)
		// CRC-16/MODBUS of 350 zero bytes is a known constant.
		assert.NotEqual(t, uint16(0), crc)
	})

	t.Run("deterministic", func(t *testing.T) {
		raw := make([]byte, HeaderSize)
		for i := range raw[:0x15E] {
			raw[i] = byte(i & 0xFF)
		}
		crc1 := ComputeHeaderCRC(raw)
		crc2 := ComputeHeaderCRC(raw)
		assert.Equal(t, crc1, crc2)
	})
}
