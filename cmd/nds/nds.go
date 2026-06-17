package nds

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "nds",
	Short: "NDS ROM toolset",
	Long:  "Inspect NDS ROM info, apply AP patches, manage the database, and upgrade flashcart kernels.",
}
