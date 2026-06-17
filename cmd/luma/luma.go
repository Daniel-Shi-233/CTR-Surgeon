package luma

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "luma",
	Short: "Luma3DS toolset",
	Long:  "Luma locale configuration, font installation, and checks.",
}
