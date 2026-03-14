package nds

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "nds",
	Short: "NDS ROM 工具集",
	Long:  "NDS ROM 信息查看、AP 补丁、数据库管理、烧录卡内核升级。",
}
