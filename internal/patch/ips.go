package patch

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrInvalidIPS    = errors.New("invalid IPS patch: missing PATCH header")
	ErrTruncatedIPS  = errors.New("truncated IPS patch data")
	ErrOffsetTooLarge = errors.New("IPS patch offset exceeds ROM size")
)

// IPSRecord represents a single IPS patch record.
type IPSRecord struct {
	Offset uint32
	Data   []byte // For RLE: expanded data
	IsRLE  bool
}

// IPSPatch holds all records parsed from an IPS file.
type IPSPatch struct {
	Records []IPSRecord
}

// ParseIPS parses an IPS format patch from raw bytes.
// IPS format: "PATCH" header, then records, then "EOF" footer.
// Record: [3B offset BE][2B size BE][data]
// RLE:    [3B offset BE][0x0000][2B rle_count BE][1B value]
// Edge case: offset 0x454F46 == "EOF" ASCII — must check if more data follows.
func ParseIPS(data []byte) (*IPSPatch, error) {
	if len(data) < 5 || string(data[:5]) != "PATCH" {
		return nil, ErrInvalidIPS
	}

	patch := &IPSPatch{}
	pos := 5

	for {
		if pos+3 > len(data) {
			return nil, ErrTruncatedIPS
		}

		// Check for EOF marker.
		if string(data[pos:pos+3]) == "EOF" {
			// If exactly at the end, we're done.
			if pos+3 >= len(data) {
				break
			}
			// More data after "EOF" could be a truncation offset (IPS32)
			// or it could be a record at offset 0x454F46.
			// Check if remaining data forms a valid record.
			if pos+5 <= len(data) {
				// Try parsing as a record at offset 0x454F46.
				size := binary.BigEndian.Uint16(data[pos+3 : pos+5])
				if size == 0 {
					// Could be RLE at this offset.
					if pos+8 <= len(data) {
						rec, n, err := parseRecord(data[pos:])
						if err == nil {
							patch.Records = append(patch.Records, rec)
							pos += n
							continue
						}
					}
				} else if pos+5+int(size) <= len(data) {
					rec, n, err := parseRecord(data[pos:])
					if err == nil {
						patch.Records = append(patch.Records, rec)
						pos += n
						continue
					}
				}
			}
			break
		}

		rec, n, err := parseRecord(data[pos:])
		if err != nil {
			return nil, err
		}
		patch.Records = append(patch.Records, rec)
		pos += n
	}

	return patch, nil
}

func parseRecord(data []byte) (IPSRecord, int, error) {
	if len(data) < 5 {
		return IPSRecord{}, 0, ErrTruncatedIPS
	}

	offset := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	size := binary.BigEndian.Uint16(data[3:5])

	if size == 0 {
		// RLE record.
		if len(data) < 8 {
			return IPSRecord{}, 0, ErrTruncatedIPS
		}
		rleCount := binary.BigEndian.Uint16(data[5:7])
		value := data[7]
		expanded := make([]byte, rleCount)
		for i := range expanded {
			expanded[i] = value
		}
		return IPSRecord{
			Offset: offset,
			Data:   expanded,
			IsRLE:  true,
		}, 8, nil
	}

	// Normal record.
	end := 5 + int(size)
	if len(data) < end {
		return IPSRecord{}, 0, ErrTruncatedIPS
	}
	recData := make([]byte, size)
	copy(recData, data[5:end])
	return IPSRecord{
		Offset: offset,
		Data:   recData,
		IsRLE:  false,
	}, end, nil
}

// Apply applies the IPS patch records to a ROM byte slice in-place.
func (p *IPSPatch) Apply(rom []byte) error {
	for i, rec := range p.Records {
		end := int(rec.Offset) + len(rec.Data)
		if end > len(rom) {
			return fmt.Errorf("record %d: %w (offset=0x%X, size=%d, rom_size=%d)",
				i, ErrOffsetTooLarge, rec.Offset, len(rec.Data), len(rom))
		}
		copy(rom[rec.Offset:end], rec.Data)
	}
	return nil
}

// IPSPatcher implements Patcher for IPS format patches.
type IPSPatcher struct {
	PatchData []byte
}

func NewIPSPatcher(patchData []byte) *IPSPatcher {
	return &IPSPatcher{PatchData: patchData}
}

func (p *IPSPatcher) Apply(romPath string, opts PatchOptions) error {
	ips, err := ParseIPS(p.PatchData)
	if err != nil {
		return fmt.Errorf("parsing IPS: %w", err)
	}

	rom, err := os.ReadFile(romPath)
	if err != nil {
		return fmt.Errorf("reading ROM: %w", err)
	}

	if opts.DryRun {
		fmt.Printf("Dry run: %d records would be applied to %s\n", len(ips.Records), romPath)
		for i, rec := range ips.Records {
			fmt.Printf("  Record %d: offset=0x%06X size=%d\n", i, rec.Offset, len(rec.Data))
		}
		return nil
	}

	if opts.Backup {
		if err := os.WriteFile(romPath+".bak", rom, 0644); err != nil {
			return fmt.Errorf("creating backup: %w", err)
		}
	}

	if err := ips.Apply(rom); err != nil {
		return err
	}

	outPath := romPath
	if opts.OutputPath != "" {
		outPath = opts.OutputPath
	}

	if err := os.WriteFile(outPath, rom, 0644); err != nil {
		return fmt.Errorf("writing patched ROM: %w", err)
	}

	return nil
}

// ReadIPSFile reads an IPS patch from a file path.
func ReadIPSFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
