package patch

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidHexPatch = errors.New("invalid hex patch format")
	ErrVerifyFailed    = errors.New("hex patch verification failed: original bytes do not match")
)

// HexPatch represents a single offset→data replacement from DS-Scene OpenPatch format.
type HexPatch struct {
	Offset   uint32
	Original []byte // bytes expected at offset before patching (for verification)
	Patched  []byte // bytes to write at offset
}

// ParseOpenPatch parses DS-Scene GameList.txt format.
// Each game entry has lines like:
//
//	[game title]
//	save_type = ...
//	00012FA8 = 64636FF3 -> 64646FF3
//	00012FA8 = 64636FF3 → 64646FF3  (unicode arrow also supported)
//
// Lines with "=" and arrow separators are hex patches.
func ParseOpenPatch(text string) ([]HexPatch, error) {
	var patches []HexPatch
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip non-patch lines (section headers, save_type, etc.)
		if !strings.Contains(line, "=") {
			continue
		}

		// Normalize arrow variants.
		normalized := strings.ReplaceAll(line, "→", "->")
		if !strings.Contains(normalized, "->") {
			continue
		}

		p, err := parseHexPatchLine(normalized)
		if err != nil {
			continue // Skip unparseable lines silently.
		}
		patches = append(patches, p)
	}

	return patches, scanner.Err()
}

func parseHexPatchLine(line string) (HexPatch, error) {
	// Format: "OFFSET = ORIGINAL -> PATCHED"
	eqParts := strings.SplitN(line, "=", 2)
	if len(eqParts) != 2 {
		return HexPatch{}, ErrInvalidHexPatch
	}

	offsetStr := strings.TrimSpace(eqParts[0])
	var offset uint32
	if _, err := fmt.Sscanf(offsetStr, "%X", &offset); err != nil {
		return HexPatch{}, fmt.Errorf("%w: bad offset %q", ErrInvalidHexPatch, offsetStr)
	}

	arrowParts := strings.SplitN(strings.TrimSpace(eqParts[1]), "->", 2)
	if len(arrowParts) != 2 {
		return HexPatch{}, ErrInvalidHexPatch
	}

	origHex := strings.ReplaceAll(strings.TrimSpace(arrowParts[0]), " ", "")
	patchHex := strings.ReplaceAll(strings.TrimSpace(arrowParts[1]), " ", "")

	original, err := hex.DecodeString(origHex)
	if err != nil {
		return HexPatch{}, fmt.Errorf("%w: bad original hex %q", ErrInvalidHexPatch, origHex)
	}

	patched, err := hex.DecodeString(patchHex)
	if err != nil {
		return HexPatch{}, fmt.Errorf("%w: bad patched hex %q", ErrInvalidHexPatch, patchHex)
	}

	if len(original) != len(patched) {
		return HexPatch{}, fmt.Errorf("%w: original and patched length mismatch", ErrInvalidHexPatch)
	}

	return HexPatch{
		Offset:   offset,
		Original: original,
		Patched:  patched,
	}, nil
}

// ApplyHexPatches applies a list of hex patches to a ROM byte slice in-place.
// If verify is true, it checks that the original bytes match before patching.
func ApplyHexPatches(rom []byte, patches []HexPatch, verify bool) error {
	for i, p := range patches {
		end := int(p.Offset) + len(p.Patched)
		if end > len(rom) {
			return fmt.Errorf("patch %d: offset 0x%X + size %d exceeds ROM size %d",
				i, p.Offset, len(p.Patched), len(rom))
		}

		if verify {
			actual := rom[p.Offset:end]
			if !bytesEqual(actual, p.Original) {
				return fmt.Errorf("patch %d at 0x%X: %w (expected %X, got %X)",
					i, p.Offset, ErrVerifyFailed, p.Original, actual)
			}
		}

		copy(rom[p.Offset:end], p.Patched)
	}
	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
