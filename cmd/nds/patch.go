package nds

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/apdb"
	ndslib "github.com/Daniel-Shi-233/CTR-Surgeon/internal/nds"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/patch"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

var (
	patchOutput string
	patchBackup bool
	patchDryRun bool
	patchIPS    string
)

var patchCmd = &cobra.Command{
	Use:   "patch <rom>",
	Short: "auto AP patch",
	Long:  "Auto-detect an NDS ROM and apply the matching AP patch. Supports a manually specified IPS file.",
	Args:  cobra.ExactArgs(1),
	RunE:  runPatch,
}

func init() {
	patchCmd.Flags().StringVarP(&patchOutput, "output", "o", "", "output file path (defaults to in-place)")
	patchCmd.Flags().BoolVar(&patchBackup, "backup", false, "create a .bak backup before modifying")
	patchCmd.Flags().BoolVar(&patchDryRun, "dry-run", false, "show patch info only, don't modify")
	patchCmd.Flags().StringVar(&patchIPS, "ips", "", "manually specify an IPS patch file")
	Cmd.AddCommand(patchCmd)
}

func runPatch(cmd *cobra.Command, args []string) error {
	romPath := args[0]

	// Parse ROM header.
	f, err := os.Open(romPath)
	if err != nil {
		return fmt.Errorf("failed to open ROM: %w", err)
	}
	defer f.Close()

	h, raw, err := ndslib.ParseHeader(f)
	if err != nil {
		return err
	}

	apID := h.APPatchID(raw)
	fmt.Printf("%s game: %s (%s)\n", ui.IconInfo, h.Title(), apID)

	var patchData []byte

	if patchIPS != "" {
		// Manual IPS file.
		patchData, err = patch.ReadIPSFile(patchIPS)
		if err != nil {
			return fmt.Errorf("failed to read IPS file: %w", err)
		}
		fmt.Printf("%s using manually specified IPS: %s\n", ui.IconInfo, patchIPS)
	} else {
		// Auto-lookup from database.
		db, err := apdb.NewDatabase()
		if err != nil {
			return fmt.Errorf("failed to load AP database: %w", err)
		}

		crc := fmt.Sprintf("%04X", ndslib.ComputeHeaderCRC(raw))
		entry, err := db.Lookup(h.Code(), crc)
		if err != nil {
			return fmt.Errorf("%s no AP patch found: %s\nuse --ips to specify a patch file manually", ui.IconFail, apID)
		}

		fmt.Printf("%s patch found: %s (%s)\n", ui.IconOK, entry.Title, entry.File)

		if patchDryRun {
			fmt.Printf("\n%s dry-run mode — no files will be modified\n", ui.IconWarn)
			fmt.Printf("  patch file: %s\n", entry.File)
			fmt.Printf("  cache path: %s/%s\n", db.CacheDir(), entry.File)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		patchData, err = db.FetchPatch(ctx, entry)
		if err != nil {
			return fmt.Errorf("failed to download patch: %w", err)
		}
		fmt.Printf("%s patch downloaded (%d bytes)\n", ui.IconOK, len(patchData))
	}

	// Apply IPS patch.
	patcher := patch.NewIPSPatcher(patchData)
	opts := patch.PatchOptions{
		DryRun:     patchDryRun,
		Backup:     patchBackup,
		OutputPath: patchOutput,
	}

	if err := patcher.Apply(romPath, opts); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	if !patchDryRun {
		target := romPath
		if patchOutput != "" {
			target = patchOutput
		}
		fmt.Printf("%s patch applied successfully: %s\n", ui.IconOK, target)
	}

	return nil
}
