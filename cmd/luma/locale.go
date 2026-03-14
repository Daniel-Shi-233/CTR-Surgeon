package luma

import (
	"fmt"

	"github.com/spf13/cobra"
	lumalib "github.com/xingshiyu/ctr-surgeon/internal/luma"
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

var (
	localeRegion string
	localeLang   string
	localeSD     string
)

var localeCmd = &cobra.Command{
	Use:   "locale",
	Short: "Luma locale 管理",
}

var localeSetCmd = &cobra.Command{
	Use:   "set <title_id>",
	Short: "设置游戏 locale",
	Long:  "为指定 Title ID 创建 locale.txt，强制游戏使用指定区域和语言。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if localeSD == "" {
			return fmt.Errorf("请指定 SD 卡路径: --sd <path>")
		}

		titleID := args[0]
		if err := lumalib.GenerateLocale(localeSD, titleID, localeRegion, localeLang); err != nil {
			return err
		}

		fmt.Printf("%s locale 已设置: %s → %s %s\n", ui.IconOK, titleID, localeRegion, localeLang)
		return nil
	},
}

var localeListCmd = &cobra.Command{
	Use:   "list <sd_root>",
	Short: "列出已配置的 locale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		locales, err := lumalib.ListLocales(args[0])
		if err != nil {
			return err
		}

		if len(locales) == 0 {
			fmt.Printf("%s 未找到任何 locale 配置\n", ui.IconWarn)
			return nil
		}

		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("已配置的 Locale (%d 个)", len(locales))))
		fmt.Println()
		for titleID, locale := range locales {
			fmt.Printf("  %s %s\n", ui.LabelStyle.Render(titleID), ui.ValueStyle.Render(locale))
		}

		return nil
	},
}

func init() {
	localeSetCmd.Flags().StringVar(&localeRegion, "region", "JPN", "区域代码 (JPN/USA/EUR/AUS/CHN/KOR/TWN)")
	localeSetCmd.Flags().StringVar(&localeLang, "lang", "JP", "语言代码 (JP/EN/FR/DE/IT/ES/ZH/KO/NL/PT/RU/TW)")
	localeSetCmd.Flags().StringVar(&localeSD, "sd", "", "SD 卡根目录路径")

	localeCmd.AddCommand(localeSetCmd)
	localeCmd.AddCommand(localeListCmd)
	Cmd.AddCommand(localeCmd)
}
