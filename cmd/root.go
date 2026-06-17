package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Daniel-Shi-233/CTR-Surgeon/cmd/fat32"
	"github.com/Daniel-Shi-233/CTR-Surgeon/cmd/luma"
	"github.com/Daniel-Shi-233/CTR-Surgeon/cmd/nds"
)

var (
	verbose bool
	sdRoot  string
)

var rootCmd = &cobra.Command{
	Use:   "ctr-surgeon",
	Short: "CTR-Surgeon — 3DS/NDS first-aid kit",
	Long:  "Cross-platform CLI for automating font installation, AP patch injection, and flashcart kernel upgrades.",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVar(&sdRoot, "sd-root", "", "path to the SD card root")

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
