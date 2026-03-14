package fat32

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "fat32",
	Short: "FAT32 SD 卡工具",
	Long:  "SD 卡健康检查、文件系统验证。",
}
