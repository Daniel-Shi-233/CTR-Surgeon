package luma

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "luma",
	Short: "Luma3DS 工具集",
	Long:  "Luma locale 设置、字库安装与检查。",
}
