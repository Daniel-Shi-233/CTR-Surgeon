package fat32

import (
	"fmt"

	"github.com/spf13/cobra"
	fat32lib "github.com/xingshiyu/ctr-surgeon/internal/fat32"
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

var (
	checkAuto bool
	checkJSON bool
)

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "SD 卡健康检查",
	Long:  "检查 SD 卡的关键目录和文件是否完整。",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&checkAuto, "auto", false, "自动检测 SD 卡")
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "以 JSON 格式输出")
	Cmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	var paths []string

	if len(args) > 0 {
		paths = append(paths, args[0])
	} else if checkAuto {
		paths = fat32lib.DetectSDCards()
		if len(paths) == 0 {
			fmt.Printf("%s 未检测到 SD 卡\n", ui.IconWarn)
			return nil
		}
	} else {
		return fmt.Errorf("请指定 SD 卡路径或使用 --auto 自动检测")
	}

	for _, path := range paths {
		report, err := fat32lib.CheckSD(path)
		if err != nil {
			fmt.Printf("%s %s: %v\n", ui.IconFail, path, err)
			continue
		}

		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("SD 卡检查: %s", report.Path)))
		fmt.Println()

		for _, c := range report.Checks {
			icon := ui.IconOK
			if !c.OK {
				icon = ui.IconFail
			}
			fmt.Printf("  %s %s — %s\n", icon, ui.LabelStyle.Render(c.Name), c.Message)
		}

		total := len(report.Checks)
		passed := report.PassCount()
		fmt.Println()
		if passed == total {
			fmt.Printf("  %s 全部通过 (%d/%d)\n", ui.IconOK, passed, total)
		} else {
			fmt.Printf("  %s 通过 %d/%d 项检查\n", ui.IconWarn, passed, total)
		}
		fmt.Println()
	}

	return nil
}
