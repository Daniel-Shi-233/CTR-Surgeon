package nds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xingshiyu/ctr-surgeon/internal/apdb"
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "AP 补丁数据库管理",
}

var dbUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新 AP 补丁数据库",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s 当前使用嵌入式数据库，无需手动更新\n", ui.IconInfo)
		fmt.Println("  升级 ctr-surgeon 版本即可获取最新数据库")
		return nil
	},
}

var dbSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "搜索 AP 补丁数据库",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := apdb.NewDatabase()
		if err != nil {
			return fmt.Errorf("加载数据库失败: %w", err)
		}

		results := db.Search(args[0])
		if len(results) == 0 {
			fmt.Printf("%s 未找到匹配 %q 的条目\n", ui.IconWarn, args[0])
			return nil
		}

		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("搜索结果: %q (%d 条)", args[0], len(results))))
		fmt.Println()
		for _, e := range results {
			fmt.Printf("  %s %s\n", ui.LabelStyle.Render(e.GameCode+"-"+e.HeaderCRC), ui.ValueStyle.Render(e.Title))
			if e.Notes != "" {
				fmt.Printf("    %s\n", ui.DimStyle.Render(e.Notes))
			}
		}

		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbUpdateCmd)
	dbCmd.AddCommand(dbSearchCmd)
	Cmd.AddCommand(dbCmd)
}
