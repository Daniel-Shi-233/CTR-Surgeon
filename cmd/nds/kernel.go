package nds

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

const (
	woodR4Version = "1.62"
	woodR4URL     = "https://github.com/DS-Homebrew/nds-bootstrap/releases/download/v1.5.8/nds-bootstrap.zip"
	// Key files to look for after extraction.
	dsmenuFile = "_DSMENU.DAT"
	rpgDir     = "__rpg"
	// maxKernelFileSize caps a single extracted entry, guarding against
	// decompression bombs in a tampered or MITM'd archive.
	maxKernelFileSize = 64 << 20 // 64 MiB
)

var kernelCmd = &cobra.Command{
	Use:   "kernel",
	Short: "flashcart kernel management",
}

var kernelUpdateCmd = &cobra.Command{
	Use:   "update <sd_root>",
	Short: "update the flashcart kernel (Wood R4)",
	Long: fmt.Sprintf(
		"Download Wood R4 v%s and install it to the SD card.\nExisting kernel files are backed up automatically.",
		woodR4Version,
	),
	Args: cobra.ExactArgs(1),
	RunE: runKernelUpdate,
}

var kernelCheckCmd = &cobra.Command{
	Use:   "check <sd_root>",
	Short: "check the SD card kernel file status",
	Args:  cobra.ExactArgs(1),
	RunE:  runKernelCheck,
}

func init() {
	kernelCmd.AddCommand(kernelUpdateCmd)
	kernelCmd.AddCommand(kernelCheckCmd)
	Cmd.AddCommand(kernelCmd)
}

func runKernelCheck(cmd *cobra.Command, args []string) error {
	sdRoot := args[0]

	fmt.Println(ui.TitleStyle.Render("Flashcart kernel check"))
	fmt.Println()

	checkFile(sdRoot, dsmenuFile)
	checkDir(sdRoot, rpgDir)
	checkDir(sdRoot, "_nds")

	return nil
}

func checkFile(sdRoot, name string) {
	path := filepath.Join(sdRoot, name)
	if info, err := os.Stat(path); err == nil {
		fmt.Printf("  %s %s (%d bytes)\n", ui.IconOK, name, info.Size())
	} else {
		fmt.Printf("  %s %s missing\n", ui.IconFail, name)
	}
}

func checkDir(sdRoot, name string) {
	path := filepath.Join(sdRoot, name)
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		fmt.Printf("  %s %s/\n", ui.IconOK, name)
	} else {
		fmt.Printf("  %s %s/ missing\n", ui.IconFail, name)
	}
}

func runKernelUpdate(cmd *cobra.Command, args []string) error {
	sdRoot := args[0]

	// Verify SD root exists.
	if _, err := os.Stat(sdRoot); err != nil {
		return fmt.Errorf("SD card path does not exist: %s", sdRoot)
	}

	fmt.Printf("%s downloading kernel update package...\n", ui.IconInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	zipData, err := downloadFile(ctx, woodR4URL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Printf("%s download complete (%d bytes)\n", ui.IconOK, len(zipData))

	// Write to temp file for zip.Reader.
	tmpFile, err := os.CreateTemp("", "ctr-surgeon-kernel-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(zipData); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Backup existing files. Abort before extracting if the backup fails, so
	// we never clobber the user's existing kernel without a copy.
	backupDir := filepath.Join(sdRoot, "_backup_kernel_"+time.Now().Format("20060102_150405"))
	if err := backupExisting(sdRoot, backupDir); err != nil {
		return fmt.Errorf("backup failed, aborting before extraction: %w", err)
	}

	// Extract zip to SD root.
	fmt.Printf("%s extracting to %s...\n", ui.IconInfo, sdRoot)
	if err := extractZip(tmpFile.Name(), sdRoot); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Printf("%s kernel update complete\n", ui.IconOK)
	return nil
}

func downloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func backupExisting(sdRoot, backupDir string) error {
	targets := []string{dsmenuFile, rpgDir}
	needsBackup := false

	for _, name := range targets {
		if _, err := os.Stat(filepath.Join(sdRoot, name)); err == nil {
			needsBackup = true
			break
		}
	}

	if !needsBackup {
		return nil
	}

	fmt.Printf("%s backing up old kernel files to %s\n", ui.IconInfo, filepath.Base(backupDir))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	for _, name := range targets {
		src := filepath.Join(sdRoot, name)
		dst := filepath.Join(backupDir, name)
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to back up %s: %w", name, err)
			}
		}
	}

	return nil
}

func extractZip(zipPath, destDir string) error {
	return extractZipLimited(zipPath, destDir, maxKernelFileSize)
}

func extractZipLimited(zipPath, destDir string, maxFileSize int64) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	cleanDest := filepath.Clean(destDir)
	for _, f := range r.File {
		// Sanitize path to prevent zip slip. filepath.Join cleans the result,
		// so an entry escaping destDir (e.g. "../x") fails the prefix check.
		target := filepath.Join(destDir, f.Name)
		if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			fmt.Printf("%s skipping unsafe path in archive: %q\n", ui.IconWarn, f.Name)
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		if err := extractFile(f, target, maxFileSize); err != nil {
			return err
		}
	}

	return nil
}

// extractFile writes a single zip entry to target, refusing to copy more than
// maxFileSize bytes to guard against decompression bombs.
func extractFile(f *zip.File, target string, maxFileSize int64) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.Create(target)
	if err != nil {
		return err
	}

	// LimitReader caps the copy at maxFileSize+1 so we can detect overflow.
	// Close before any os.Remove so cleanup works on Windows too.
	n, copyErr := io.Copy(outFile, io.LimitReader(rc, maxFileSize+1))
	closeErr := outFile.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if n > maxFileSize {
		os.Remove(target)
		return fmt.Errorf("archive entry %q exceeds size limit of %d bytes", f.Name, maxFileSize)
	}

	return nil
}
