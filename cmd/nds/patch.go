package nds

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xingshiyu/ctr-surgeon/internal/apdb"
	ndslib "github.com/xingshiyu/ctr-surgeon/internal/nds"
	"github.com/xingshiyu/ctr-surgeon/internal/patch"
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

var (
	patchOutput string
	patchBackup bool
	patchDryRun bool
	patchIPS    string
)

var patchCmd = &cobra.Command{
	Use:   "patch <rom>",
	Short: "自动 AP 补丁",
	Long:  "自动识别 NDS ROM 并应用对应的 AP 补丁。支持手动指定 IPS 文件。",
	Args:  cobra.ExactArgs(1),
	RunE:  runPatch,
}

func init() {
	patchCmd.Flags().StringVarP(&patchOutput, "output", "o", "", "输出文件路径 (默认原地修改)")
	patchCmd.Flags().BoolVar(&patchBackup, "backup", false, "修改前创建 .bak 备份")
	patchCmd.Flags().BoolVar(&patchDryRun, "dry-run", false, "仅显示补丁信息，不实际修改")
	patchCmd.Flags().StringVar(&patchIPS, "ips", "", "手动指定 IPS 补丁文件")
	Cmd.AddCommand(patchCmd)
}

func runPatch(cmd *cobra.Command, args []string) error {
	romPath := args[0]

	// Parse ROM header.
	f, err := os.Open(romPath)
	if err != nil {
		return fmt.Errorf("打开 ROM 失败: %w", err)
	}
	defer f.Close()

	h, raw, err := ndslib.ParseHeader(f)
	if err != nil {
		return err
	}

	apID := h.APPatchID(raw)
	fmt.Printf("%s 游戏: %s (%s)\n", ui.IconInfo, h.Title(), apID)

	var patchData []byte

	if patchIPS != "" {
		// Manual IPS file.
		patchData, err = patch.ReadIPSFile(patchIPS)
		if err != nil {
			return fmt.Errorf("读取 IPS 文件失败: %w", err)
		}
		fmt.Printf("%s 使用手动指定的 IPS: %s\n", ui.IconInfo, patchIPS)
	} else {
		// Auto-lookup from database.
		db, err := apdb.NewDatabase()
		if err != nil {
			return fmt.Errorf("加载 AP 数据库失败: %w", err)
		}

		crc := fmt.Sprintf("%04X", ndslib.ComputeHeaderCRC(raw))
		entry, err := db.Lookup(h.Code(), crc)
		if err != nil {
			return fmt.Errorf("%s 未找到 AP 补丁: %s\n建议使用 --ips 手动指定补丁文件", ui.IconFail, apID)
		}

		fmt.Printf("%s 找到补丁: %s (%s)\n", ui.IconOK, entry.Title, entry.File)

		if patchDryRun {
			fmt.Printf("\n%s Dry run 模式 — 不会修改文件\n", ui.IconWarn)
			fmt.Printf("  补丁文件: %s\n", entry.File)
			fmt.Printf("  缓存路径: %s/%s\n", db.CacheDir(), entry.File)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		patchData, err = db.FetchPatch(ctx, entry)
		if err != nil {
			return fmt.Errorf("下载补丁失败: %w", err)
		}
		fmt.Printf("%s 补丁已下载 (%d bytes)\n", ui.IconOK, len(patchData))
	}

	// Apply IPS patch.
	patcher := patch.NewIPSPatcher(patchData)
	opts := patch.PatchOptions{
		DryRun:     patchDryRun,
		Backup:     patchBackup,
		OutputPath: patchOutput,
	}

	if err := patcher.Apply(romPath, opts); err != nil {
		return fmt.Errorf("应用补丁失败: %w", err)
	}

	if !patchDryRun {
		target := romPath
		if patchOutput != "" {
			target = patchOutput
		}
		fmt.Printf("%s 补丁应用成功: %s\n", ui.IconOK, target)
	}

	return nil
}
