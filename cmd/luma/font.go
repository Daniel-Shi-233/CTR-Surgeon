package luma

import (
	"fmt"

	"github.com/spf13/cobra"
	lumalib "github.com/xingshiyu/ctr-surgeon/internal/luma"
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

var fontSD string

var fontCmd = &cobra.Command{
	Use:   "font",
	Short: "字库管理",
}

var fontInstallCmd = &cobra.Command{
	Use:   "install <bcfnt>",
	Short: "安装自定义字库到 SD 卡",
	Long:  "校验 .bcfnt 文件后复制到 Luma 字库目录。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if fontSD == "" {
			return fmt.Errorf("请指定 SD 卡路径: --sd <path>")
		}

		fontPath := args[0]
		if err := lumalib.InjectFont(fontSD, fontPath); err != nil {
			return err
		}

		fmt.Printf("%s 字库已安装到 %s/luma/font.bcfnt\n", ui.IconOK, fontSD)
		return nil
	},
}

var fontCheckCmd = &cobra.Command{
	Use:   "check <sd_root>",
	Short: "检查 SD 卡字库状态",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, path, _ := lumalib.CheckFont(args[0])
		if installed {
			fmt.Printf("%s 字库已安装: %s\n", ui.IconOK, path)
		} else {
			fmt.Printf("%s 未安装字库: %s\n", ui.IconWarn, path)
		}
		return nil
	},
}

func init() {
	fontInstallCmd.Flags().StringVar(&fontSD, "sd", "", "SD 卡根目录路径")

	fontCmd.AddCommand(fontInstallCmd)
	fontCmd.AddCommand(fontCheckCmd)
	Cmd.AddCommand(fontCmd)
}
