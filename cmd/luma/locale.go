package luma

import (
	"fmt"

	"github.com/spf13/cobra"
	lumalib "github.com/Daniel-Shi-233/CTR-Surgeon/internal/luma"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

var (
	localeRegion string
	localeLang   string
	localeSD     string
)

var localeCmd = &cobra.Command{
	Use:   "locale",
	Short: "Luma locale management",
}

var localeSetCmd = &cobra.Command{
	Use:   "set <title_id>",
	Short: "set a game's locale",
	Long:  "Create locale.txt for the given Title ID to force a specific region and language.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if localeSD == "" {
			return fmt.Errorf("SD card path required: --sd <path>")
		}

		titleID := args[0]
		if err := lumalib.GenerateLocale(localeSD, titleID, localeRegion, localeLang); err != nil {
			return err
		}

		fmt.Printf("%s locale set: %s → %s %s\n", ui.IconOK, titleID, localeRegion, localeLang)
		return nil
	},
}

var localeListCmd = &cobra.Command{
	Use:   "list <sd_root>",
	Short: "list configured locales",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		locales, err := lumalib.ListLocales(args[0])
		if err != nil {
			return err
		}

		if len(locales) == 0 {
			fmt.Printf("%s no locale configurations found\n", ui.IconWarn)
			return nil
		}

		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("Configured locales (%d)", len(locales))))
		fmt.Println()
		for titleID, locale := range locales {
			fmt.Printf("  %s %s\n", ui.LabelStyle.Render(titleID), ui.ValueStyle.Render(locale))
		}

		return nil
	},
}

func init() {
	localeSetCmd.Flags().StringVar(&localeRegion, "region", "JPN", "region code (JPN/USA/EUR/AUS/CHN/KOR/TWN)")
	localeSetCmd.Flags().StringVar(&localeLang, "lang", "JP", "language code (JP/EN/FR/DE/IT/ES/ZH/KO/NL/PT/RU/TW)")
	localeSetCmd.Flags().StringVar(&localeSD, "sd", "", "path to the SD card root")

	localeCmd.AddCommand(localeSetCmd)
	localeCmd.AddCommand(localeListCmd)
	Cmd.AddCommand(localeCmd)
}
