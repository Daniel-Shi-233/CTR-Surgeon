package fat32

import (
	"fmt"

	"github.com/spf13/cobra"
	fat32lib "github.com/Daniel-Shi-233/CTR-Surgeon/internal/fat32"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

var (
	checkAuto bool
	checkJSON bool
)

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "SD card health check",
	Long:  "Verify that the SD card's critical directories and files are present.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&checkAuto, "auto", false, "auto-detect SD cards")
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "output as JSON")
	Cmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	var paths []string

	if len(args) > 0 {
		paths = append(paths, args[0])
	} else if checkAuto {
		paths = fat32lib.DetectSDCards()
		if len(paths) == 0 {
			fmt.Printf("%s no SD card detected\n", ui.IconWarn)
			return nil
		}
	} else {
		return fmt.Errorf("specify an SD card path or use --auto to detect one")
	}

	for _, path := range paths {
		report, err := fat32lib.CheckSD(path)
		if err != nil {
			fmt.Printf("%s %s: %v\n", ui.IconFail, path, err)
			continue
		}

		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("SD card check: %s", report.Path)))
		fmt.Println()

		for _, c := range report.Checks {
			icon := ui.IconOK
			if !c.OK {
				icon = ui.IconFail
			}
			fmt.Printf("  %s %s — %s\n", icon, ui.LabelStyle.Render(c.Name), c.Message)
		}

		total := len(report.Checks)
		passed := report.PassCount()
		fmt.Println()
		if passed == total {
			fmt.Printf("  %s all passed (%d/%d)\n", ui.IconOK, passed, total)
		} else {
			fmt.Printf("  %s passed %d/%d checks\n", ui.IconWarn, passed, total)
		}
		fmt.Println()
	}

	return nil
}
