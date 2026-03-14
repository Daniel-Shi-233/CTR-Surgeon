package luma

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidTitleID = errors.New("invalid title ID: must be 16 hex characters")
	ErrInvalidRegion  = errors.New("invalid region code")
	ErrInvalidLang    = errors.New("invalid language code")
)

// Valid region codes for Luma locale.
var ValidRegions = []string{"JPN", "USA", "EUR", "AUS", "CHN", "KOR", "TWN"}

// Valid language codes for Luma locale.
var ValidLanguages = []string{"JP", "EN", "FR", "DE", "IT", "ES", "ZH", "KO", "NL", "PT", "RU", "TW"}

// GenerateLocale creates a locale.txt file for a given title ID under the Luma
// locale directory on the SD card.
// Path: <sdRoot>/luma/titles/<titleID>/locale.txt
// Content: "<region> <language>\n"
func GenerateLocale(sdRoot, titleID, region, language string) error {
	titleID = strings.ToUpper(strings.TrimSpace(titleID))
	region = strings.ToUpper(strings.TrimSpace(region))
	language = strings.ToUpper(strings.TrimSpace(language))

	if len(titleID) != 16 {
		return ErrInvalidTitleID
	}
	for _, c := range titleID {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return ErrInvalidTitleID
		}
	}

	if !contains(ValidRegions, region) {
		return fmt.Errorf("%w: %q (valid: %s)", ErrInvalidRegion, region, strings.Join(ValidRegions, ", "))
	}
	if !contains(ValidLanguages, language) {
		return fmt.Errorf("%w: %q (valid: %s)", ErrInvalidLang, language, strings.Join(ValidLanguages, ", "))
	}

	dir := filepath.Join(sdRoot, "luma", "titles", titleID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating locale directory: %w", err)
	}

	content := fmt.Sprintf("%s %s\n", region, language)
	path := filepath.Join(dir, "locale.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing locale.txt: %w", err)
	}

	return nil
}

// ListLocales scans the luma/titles directory and returns all configured locales.
// Returns a map of titleID → locale content.
func ListLocales(sdRoot string) (map[string]string, error) {
	titlesDir := filepath.Join(sdRoot, "luma", "titles")
	entries, err := os.ReadDir(titlesDir)
	if err != nil {
		return nil, fmt.Errorf("reading titles directory: %w", err)
	}

	locales := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		localePath := filepath.Join(titlesDir, entry.Name(), "locale.txt")
		data, err := os.ReadFile(localePath)
		if err != nil {
			continue
		}
		locales[entry.Name()] = strings.TrimSpace(string(data))
	}

	return locales, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
