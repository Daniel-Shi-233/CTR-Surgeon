package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xingshiyu/ctr-surgeon/cmd/fat32"
	"github.com/xingshiyu/ctr-surgeon/cmd/luma"
	"github.com/xingshiyu/ctr-surgeon/cmd/nds"
)

var (
	verbose bool
	sdRoot  string
)

var rootCmd = &cobra.Command{
	Use:   "ctr-surgeon",
	Short: "CTR-Surgeon — 3DS/NDS 急救箱",
	Long:  "跨平台 CLI 工具，自动化处理字库挂载、AP 补丁注入、烧录卡内核升级。",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "启用详细输出")
	rootCmd.PersistentFlags().StringVar(&sdRoot, "sd-root", "", "SD 卡根目录路径")

	rootCmd.AddCommand(nds.Cmd)
	rootCmd.AddCommand(luma.Cmd)
	rootCmd.AddCommand(fat32.Cmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
