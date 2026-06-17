package nds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ndslib "github.com/Daniel-Shi-233/CTR-Surgeon/internal/nds"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

var jsonOutput bool

var infoCmd = &cobra.Command{
	Use:   "info <rom>",
	Short: "show NDS ROM header info",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func init() {
	infoCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	Cmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("failed to open ROM: %w", err)
	}
	defer f.Close()

	h, raw, err := ndslib.ParseHeader(f)
	if err != nil {
		return err
	}

	crcValid := h.ValidateCRC(raw)

	fmt.Println(ui.TitleStyle.Render("NDS ROM info"))
	fmt.Println()
	printField("Game title", h.Title())
	printField("Game code", h.Code())
	printField("Maker code", h.Maker())
	printField("ROM version", fmt.Sprintf("%d", h.ROMVersion))
	printField("AP patch ID", h.APPatchID(raw))
	printField("Header CRC", fmt.Sprintf("0x%04X", h.HeaderCRC))

	if crcValid {
		printField("CRC check", ui.IconOK+" passed")
	} else {
		computed := ndslib.ComputeHeaderCRC(raw)
		printField("CRC check", fmt.Sprintf("%s failed (expected 0x%04X, computed 0x%04X)", ui.IconFail, h.HeaderCRC, computed))
	}

	printField("ARM9 offset", fmt.Sprintf("0x%08X", h.ARM9ROMOffset))
	printField("ARM9 size", fmt.Sprintf("0x%08X (%d bytes)", h.ARM9Size, h.ARM9Size))
	printField("ARM7 offset", fmt.Sprintf("0x%08X", h.ARM7ROMOffset))
	printField("ARM7 size", fmt.Sprintf("0x%08X (%d bytes)", h.ARM7Size, h.ARM7Size))
	printField("Total ROM size", fmt.Sprintf("0x%08X (%d bytes)", h.TotalUsedROMSize, h.TotalUsedROMSize))

	return nil
}

func printField(label, value string) {
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render(label), ui.ValueStyle.Render(value))
}
