package nds

import "github.com/sigurn/crc16"

var crcTable = crc16.MakeTable(crc16.CRC16_MODBUS)

// ComputeHeaderCRC calculates the CRC-16/MODBUS over the first 350 bytes
// (offsets 0x000–0x15D) of a raw NDS ROM header.
func ComputeHeaderCRC(rawHeader []byte) uint16 {
	if len(rawHeader) < 0x15E {
		return 0
	}
	return crc16.Checksum(rawHeader[:0x15E], crcTable)
}
