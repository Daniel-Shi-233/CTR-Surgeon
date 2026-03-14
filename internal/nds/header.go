package nds

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const HeaderSize = 0x200 // 512 bytes

// NDSHeader represents the full 512-byte NDS ROM header.
// Field offsets follow the official NDS header specification.
type NDSHeader struct {
	GameTitle          [12]byte  // 0x000
	GameCode           [4]byte   // 0x00C
	MakerCode          [2]byte   // 0x010
	UnitCode           byte      // 0x012
	EncryptionSeed     byte      // 0x013
	DeviceCapacity     byte      // 0x014
	Reserved1          [7]byte   // 0x015
	TWLFlags           byte      // 0x01C
	NDSRegion          byte      // 0x01D
	ROMVersion         byte      // 0x01E
	AutoStart          byte      // 0x01F
	ARM9ROMOffset      uint32    // 0x020
	ARM9EntryAddress   uint32    // 0x024
	ARM9RAMAddress     uint32    // 0x028
	ARM9Size           uint32    // 0x02C
	ARM7ROMOffset      uint32    // 0x030
	ARM7EntryAddress   uint32    // 0x034
	ARM7RAMAddress     uint32    // 0x038
	ARM7Size           uint32    // 0x03C
	FNTOffset          uint32    // 0x040
	FNTSize            uint32    // 0x044
	FATOffset          uint32    // 0x048
	FATSize            uint32    // 0x04C
	ARM9OverlayOffset  uint32    // 0x050
	ARM9OverlaySize    uint32    // 0x054
	ARM7OverlayOffset  uint32    // 0x058
	ARM7OverlaySize    uint32    // 0x05C
	NormalCardControl  uint32    // 0x060
	SecureCardControl  uint32    // 0x064
	IconBannerOffset   uint32    // 0x068
	SecureAreaCRC      uint16    // 0x06C
	SecureTransferTime uint16    // 0x06E
	ARM9AutoLoad       uint32    // 0x070
	ARM7AutoLoad       uint32    // 0x074
	SecureDisable      uint64    // 0x078
	TotalUsedROMSize   uint32    // 0x080
	ROMHeaderSize      uint32    // 0x084
	Reserved2          [0x38]byte // 0x088
	NintendoLogo       [0x9C]byte // 0x0C0
	NintendoLogoCRC    uint16    // 0x15C
	HeaderCRC          uint16    // 0x15E
	DebugROMOffset     uint32    // 0x160
	DebugSize          uint32    // 0x164
	DebugRAMAddress    uint32    // 0x168
	Reserved3          [4]byte   // 0x16C
	Reserved4          [0x90]byte // 0x170
}

// ParseHeader reads a 512-byte NDS ROM header from r.
// Returns the parsed header, the raw 512-byte slice, and any error.
func ParseHeader(r io.Reader) (*NDSHeader, []byte, error) {
	raw := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, raw); err != nil {
		return nil, nil, fmt.Errorf("reading NDS header: %w", err)
	}

	var h NDSHeader
	if err := binary.Read(
		io.NewSectionReader(newBytesReaderAt(raw), 0, int64(HeaderSize)),
		binary.LittleEndian,
		&h,
	); err != nil {
		return nil, nil, fmt.Errorf("decoding NDS header: %w", err)
	}

	return &h, raw, nil
}

// GameTitle returns the trimmed game title string.
func (h *NDSHeader) Title() string {
	return strings.TrimRight(string(h.GameTitle[:]), "\x00 ")
}

// Code returns the 4-character game code.
func (h *NDSHeader) Code() string {
	return string(h.GameCode[:])
}

// Maker returns the 2-character maker code.
func (h *NDSHeader) Maker() string {
	return string(h.MakerCode[:])
}

// APPatchID returns the AP patch identifier in "XXXX-YYYY" format,
// where XXXX is the game code and YYYY is the header CRC-16 in uppercase hex.
func (h *NDSHeader) APPatchID(rawHeader []byte) string {
	crc := ComputeHeaderCRC(rawHeader)
	return fmt.Sprintf("%s-%04X", h.Code(), crc)
}

// ValidateCRC checks whether the stored header CRC matches the computed CRC.
func (h *NDSHeader) ValidateCRC(rawHeader []byte) bool {
	return h.HeaderCRC == ComputeHeaderCRC(rawHeader)
}

// bytesReaderAt wraps a byte slice to implement io.ReaderAt.
type bytesReaderAt struct {
	data []byte
}

func newBytesReaderAt(data []byte) *bytesReaderAt {
	return &bytesReaderAt{data: data}
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n := copy(p, b.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
