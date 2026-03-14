package nds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ndslib "github.com/xingshiyu/ctr-surgeon/internal/nds"
	"github.com/xingshiyu/ctr-surgeon/internal/ui"
)

var jsonOutput bool

var infoCmd = &cobra.Command{
	Use:   "info <rom>",
	Short: "显示 NDS ROM 头信息",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func init() {
	infoCmd.Flags().BoolVar(&jsonOutput, "json", false, "以 JSON 格式输出")
	Cmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("打开 ROM 失败: %w", err)
	}
	defer f.Close()

	h, raw, err := ndslib.ParseHeader(f)
	if err != nil {
		return err
	}

	crcValid := h.ValidateCRC(raw)

	fmt.Println(ui.TitleStyle.Render("NDS ROM 信息"))
	fmt.Println()
	printField("游戏标题", h.Title())
	printField("游戏代码", h.Code())
	printField("制造商代码", h.Maker())
	printField("ROM 版本", fmt.Sprintf("%d", h.ROMVersion))
	printField("AP 补丁 ID", h.APPatchID(raw))
	printField("Header CRC", fmt.Sprintf("0x%04X", h.HeaderCRC))

	if crcValid {
		printField("CRC 校验", ui.IconOK+" 通过")
	} else {
		computed := ndslib.ComputeHeaderCRC(raw)
		printField("CRC 校验", fmt.Sprintf("%s 失败 (期望 0x%04X, 计算 0x%04X)", ui.IconFail, h.HeaderCRC, computed))
	}

	printField("ARM9 偏移", fmt.Sprintf("0x%08X", h.ARM9ROMOffset))
	printField("ARM9 大小", fmt.Sprintf("0x%08X (%d bytes)", h.ARM9Size, h.ARM9Size))
	printField("ARM7 偏移", fmt.Sprintf("0x%08X", h.ARM7ROMOffset))
	printField("ARM7 大小", fmt.Sprintf("0x%08X (%d bytes)", h.ARM7Size, h.ARM7Size))
	printField("ROM 总大小", fmt.Sprintf("0x%08X (%d bytes)", h.TotalUsedROMSize, h.TotalUsedROMSize))

	return nil
}

func printField(label, value string) {
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render(label), ui.ValueStyle.Render(value))
}
