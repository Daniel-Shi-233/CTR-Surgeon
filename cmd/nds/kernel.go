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
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

const (
	woodR4Version = "1.62"
	woodR4URL     = "https://github.com/DS-Homebrew/nds-bootstrap/releases/download/v1.5.8/nds-bootstrap.zip"
	// Key files to look for after extraction.
	dsmenuFile = "_DSMENU.DAT"
	rpgDir     = "__rpg"
)

var kernelCmd = &cobra.Command{
	Use:   "kernel",
	Short: "烧录卡内核管理",
}

var kernelUpdateCmd = &cobra.Command{
	Use:   "update <sd_root>",
	Short: "更新烧录卡内核 (Wood R4)",
	Long: fmt.Sprintf(
		"下载 Wood R4 v%s 并安装到 SD 卡。\n自动备份旧内核文件。",
		woodR4Version,
	),
	Args: cobra.ExactArgs(1),
	RunE: runKernelUpdate,
}

var kernelCheckCmd = &cobra.Command{
	Use:   "check <sd_root>",
	Short: "检查 SD 卡内核文件状态",
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

	fmt.Println(ui.TitleStyle.Render("烧录卡内核检查"))
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
		fmt.Printf("  %s %s 不存在\n", ui.IconFail, name)
	}
}

func checkDir(sdRoot, name string) {
	path := filepath.Join(sdRoot, name)
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		fmt.Printf("  %s %s/\n", ui.IconOK, name)
	} else {
		fmt.Printf("  %s %s/ 不存在\n", ui.IconFail, name)
	}
}

func runKernelUpdate(cmd *cobra.Command, args []string) error {
	sdRoot := args[0]

	// Verify SD root exists.
	if _, err := os.Stat(sdRoot); err != nil {
		return fmt.Errorf("SD 卡路径不存在: %s", sdRoot)
	}

	fmt.Printf("%s 下载内核更新包...\n", ui.IconInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	zipData, err := downloadFile(ctx, woodR4URL)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	fmt.Printf("%s 下载完成 (%d bytes)\n", ui.IconOK, len(zipData))

	// Write to temp file for zip.Reader.
	tmpFile, err := os.CreateTemp("", "ctr-surgeon-kernel-*.zip")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(zipData); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	// Backup existing files.
	backupDir := filepath.Join(sdRoot, "_backup_kernel_"+time.Now().Format("20060102_150405"))
	backupExisting(sdRoot, backupDir)

	// Extract zip to SD root.
	fmt.Printf("%s 解压到 %s...\n", ui.IconInfo, sdRoot)
	if err := extractZip(tmpFile.Name(), sdRoot); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	fmt.Printf("%s 内核更新完成\n", ui.IconOK)
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

func backupExisting(sdRoot, backupDir string) {
	targets := []string{dsmenuFile, rpgDir}
	needsBackup := false

	for _, name := range targets {
		if _, err := os.Stat(filepath.Join(sdRoot, name)); err == nil {
			needsBackup = true
			break
		}
	}

	if !needsBackup {
		return
	}

	fmt.Printf("%s 备份旧内核文件到 %s\n", ui.IconInfo, filepath.Base(backupDir))
	os.MkdirAll(backupDir, 0755)

	for _, name := range targets {
		src := filepath.Join(sdRoot, name)
		dst := filepath.Join(backupDir, name)
		if _, err := os.Stat(src); err == nil {
			os.Rename(src, dst)
		}
	}
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Sanitize path to prevent zip slip.
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		outFile, err := os.Create(target)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
