package nds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/apdb"
	"github.com/Daniel-Shi-233/CTR-Surgeon/internal/ui"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "AP patch database management",
}

var dbUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update the AP patch database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s using the embedded database; no manual update needed\n", ui.IconInfo)
		fmt.Println("  upgrade ctr-surgeon to get the latest database")
		return nil
	},
}

var dbSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "search the AP patch database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := apdb.NewDatabase()
		if err != nil {
			return fmt.Errorf("failed to load database: %w", err)
		}

		results := db.Search(args[0])
		if len(results) == 0 {
			fmt.Printf("%s no entries matching %q\n", ui.IconWarn, args[0])
			return nil
		}

		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("Search results: %q (%d)", args[0], len(results))))
		fmt.Println()
		for _, e := range results {
			fmt.Printf("  %s %s\n", ui.LabelStyle.Render(e.GameCode+"-"+e.HeaderCRC), ui.ValueStyle.Render(e.Title))
			if e.Notes != "" {
				fmt.Printf("    %s\n", ui.DimStyle.Render(e.Notes))
			}
		}

		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbUpdateCmd)
	dbCmd.AddCommand(dbSearchCmd)
	Cmd.AddCommand(dbCmd)
}
