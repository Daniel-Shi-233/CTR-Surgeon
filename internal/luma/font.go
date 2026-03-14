package luma

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrInvalidFont = errors.New("invalid font file: missing CFNT magic")
	ErrFontTooLarge = errors.New("font file exceeds 1.5MB size limit")
)

const (
	cfntMagic   = "CFNT"
	maxFontSize = 1536 * 1024 // 1.5 MB
)

// ValidateFont checks that a file has the CFNT magic header and is within size limits.
func ValidateFont(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("checking font file: %w", err)
	}

	if info.Size() > maxFontSize {
		return fmt.Errorf("%w: %d bytes (max %d)", ErrFontTooLarge, info.Size(), maxFontSize)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening font file: %w", err)
	}
	defer f.Close()

	magic := make([]byte, 4)
	if _, err := io.ReadFull(f, magic); err != nil {
		return fmt.Errorf("%w: cannot read header", ErrInvalidFont)
	}

	if string(magic) != cfntMagic {
		return fmt.Errorf("%w: got %q", ErrInvalidFont, string(magic))
	}

	return nil
}

// InjectFont copies a validated .bcfnt file to the Luma font directory on the SD card.
// Path: <sdRoot>/luma/font.bcfnt
func InjectFont(sdRoot, fontPath string) error {
	if err := ValidateFont(fontPath); err != nil {
		return err
	}

	lumaDir := filepath.Join(sdRoot, "luma")
	if err := os.MkdirAll(lumaDir, 0755); err != nil {
		return fmt.Errorf("creating luma directory: %w", err)
	}

	src, err := os.Open(fontPath)
	if err != nil {
		return fmt.Errorf("opening source font: %w", err)
	}
	defer src.Close()

	dstPath := filepath.Join(lumaDir, "font.bcfnt")
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("creating destination font: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copying font data: %w", err)
	}

	return nil
}

// CheckFont checks if a font is installed on the SD card.
func CheckFont(sdRoot string) (bool, string, error) {
	fontPath := filepath.Join(sdRoot, "luma", "font.bcfnt")
	info, err := os.Stat(fontPath)
	if os.IsNotExist(err) {
		return false, fontPath, nil
	}
	if err != nil {
		return false, fontPath, err
	}
	return true, fontPath, fmt.Errorf("font installed: %d bytes", info.Size())
}
