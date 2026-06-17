package fat32

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "fat32",
	Short: "FAT32 SD card tools",
	Long:  "SD card health checks and filesystem validation.",
}
