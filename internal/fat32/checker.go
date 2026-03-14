package fat32

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// CheckResult represents the result of a single check item.
type CheckResult struct {
	Name    string
	OK      bool
	Message string
}

// SDReport is the full health check report for an SD card.
type SDReport struct {
	Path    string
	Checks  []CheckResult
	IsSD    bool
	FSType  string
}

// Critical directories that should exist on a healthy 3DS SD card.
var criticalDirs = []struct {
	Path string
	Desc string
}{
	{"Nintendo 3DS", "3DS 系统数据目录"},
	{"luma", "Luma3DS 配置目录"},
	{"_nds", "NDS 工具链目录"},
	{"__rpg", "烧录卡内核目录"},
}

// Critical files that should exist.
var criticalFiles = []struct {
	Path string
	Desc string
}{
	{"boot.firm", "Luma3DS 引导文件"},
	{"boot.3dsx", "Homebrew Launcher"},
}

// CheckSD performs a health check on the given SD card path.
func CheckSD(sdRoot string) (*SDReport, error) {
	report := &SDReport{
		Path: sdRoot,
	}

	// Check if path exists and is a directory.
	info, err := os.Stat(sdRoot)
	if err != nil {
		return nil, fmt.Errorf("无法访问路径: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("路径不是目录: %s", sdRoot)
	}

	// Check critical directories.
	for _, d := range criticalDirs {
		path := filepath.Join(sdRoot, d.Path)
		if dirInfo, err := os.Stat(path); err == nil && dirInfo.IsDir() {
			report.Checks = append(report.Checks, CheckResult{
				Name:    d.Desc,
				OK:      true,
				Message: d.Path + "/ 存在",
			})
		} else {
			report.Checks = append(report.Checks, CheckResult{
				Name:    d.Desc,
				OK:      false,
				Message: d.Path + "/ 不存在",
			})
		}
	}

	// Check critical files.
	for _, f := range criticalFiles {
		path := filepath.Join(sdRoot, f.Path)
		if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
			report.Checks = append(report.Checks, CheckResult{
				Name:    f.Desc,
				OK:      true,
				Message: fmt.Sprintf("%s (%d bytes)", f.Path, fileInfo.Size()),
			})
		} else {
			report.Checks = append(report.Checks, CheckResult{
				Name:    f.Desc,
				OK:      false,
				Message: f.Path + " 不存在",
			})
		}
	}

	// Check free space (best-effort).
	if freeBytes, err := getFreeSpace(sdRoot); err == nil {
		freeMB := freeBytes / (1024 * 1024)
		ok := freeMB > 100
		msg := fmt.Sprintf("%d MB 可用", freeMB)
		if !ok {
			msg += " (建议 > 100 MB)"
		}
		report.Checks = append(report.Checks, CheckResult{
			Name:    "可用空间",
			OK:      ok,
			Message: msg,
		})
	}

	return report, nil
}

// DetectSDCards attempts to find mounted SD cards.
func DetectSDCards() []string {
	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		// macOS: check /Volumes/
		entries, err := os.ReadDir("/Volumes")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() && e.Name() != "Macintosh HD" {
					candidates = append(candidates, filepath.Join("/Volumes", e.Name()))
				}
			}
		}
	case "windows":
		// Windows: check drive letters D: through Z:
		for c := 'D'; c <= 'Z'; c++ {
			path := string(c) + ":\\"
			if _, err := os.Stat(path); err == nil {
				candidates = append(candidates, path)
			}
		}
	case "linux":
		// Linux: check /media/$USER/ and /mnt/
		if user := os.Getenv("USER"); user != "" {
			mediaDir := filepath.Join("/media", user)
			entries, err := os.ReadDir(mediaDir)
			if err == nil {
				for _, e := range entries {
					if e.IsDir() {
						candidates = append(candidates, filepath.Join(mediaDir, e.Name()))
					}
				}
			}
		}
		entries, err := os.ReadDir("/mnt")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					candidates = append(candidates, filepath.Join("/mnt", e.Name()))
				}
			}
		}
	}

	return candidates
}

// PassCount returns the number of passed checks.
func (r *SDReport) PassCount() int {
	count := 0
	for _, c := range r.Checks {
		if c.OK {
			count++
		}
	}
	return count
}
