package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Set by goreleaser via ldflags.
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ctr-surgeon %s (commit: %s, built: %s)\n", version, commit, date)
	},
}
