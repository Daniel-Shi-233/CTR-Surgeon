package luma

import (
	"fmt"

	"github.com/spf13/cobra"
	lumalib "github.com/Daniel-Shi-233/CTR-Surgeon/internal/luma"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

var fontSD string

var fontCmd = &cobra.Command{
	Use:   "font",
	Short: "font management",
}

var fontInstallCmd = &cobra.Command{
	Use:   "install <bcfnt>",
	Short: "install a custom font to the SD card",
	Long:  "Validate a .bcfnt file and copy it into the Luma font directory.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if fontSD == "" {
			return fmt.Errorf("SD card path required: --sd <path>")
		}

		fontPath := args[0]
		if err := lumalib.InjectFont(fontSD, fontPath); err != nil {
			return err
		}

		fmt.Printf("%s font installed to %s/luma/font.bcfnt\n", ui.IconOK, fontSD)
		return nil
	},
}

var fontCheckCmd = &cobra.Command{
	Use:   "check <sd_root>",
	Short: "check the SD card font status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, path, _ := lumalib.CheckFont(args[0])
		if installed {
			fmt.Printf("%s font installed: %s\n", ui.IconOK, path)
		} else {
			fmt.Printf("%s no font installed: %s\n", ui.IconWarn, path)
		}
		return nil
	},
}

func init() {
	fontInstallCmd.Flags().StringVar(&fontSD, "sd", "", "path to the SD card root")

	fontCmd.AddCommand(fontInstallCmd)
	fontCmd.AddCommand(fontCheckCmd)
	Cmd.AddCommand(fontCmd)
}
