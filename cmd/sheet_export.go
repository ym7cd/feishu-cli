package cmd

import (
	"fmt"
	"os"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var sheetExportCmd = &cobra.Command{
	Use:   "export <spreadsheet_token>",
	Short: "导出电子表格为 XLSX 或 CSV 文件",
	Long: `导出电子表格为文件。支持 XLSX 和 CSV 两种格式。

CSV 格式导出时必须指定 --sheet-id 参数（只能导出单个工作表）。

示例:
  # 导出为 XLSX
  feishu-cli sheet export SPREADSHEET_TOKEN -o output.xlsx

  # 导出为 CSV（需指定工作表）
  feishu-cli sheet export SPREADSHEET_TOKEN --format csv --sheet-id SHEET_ID -o output.csv

  # 导出为 XLSX（自定义重试次数）
  feishu-cli sheet export SPREADSHEET_TOKEN -o report.xlsx --max-retries 60`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		spreadsheetToken := args[0]
		format, _ := cmd.Flags().GetString("format")
		sheetID, _ := cmd.Flags().GetString("sheet-id")
		outputPath, _ := cmd.Flags().GetString("output")
		maxRetries, _ := cmd.Flags().GetInt("max-retries")

		// CSV 格式必须指定 sheet-id
		if format == "csv" && sheetID == "" {
			return fmt.Errorf("CSV 格式导出必须指定 --sheet-id 参数（使用 feishu-cli sheet list-sheets <token> 查看工作表 ID）")
		}

		// 默认输出文件名
		if outputPath == "" {
			outputPath = spreadsheetToken + "." + format
		}

		// 获取可选的 User Access Token
		userAccessToken := resolveOptionalUserTokenWithFallback(cmd)

		// 创建导出任务
		fmt.Fprintf(os.Stderr, "正在创建导出任务...\n")
		ticket, err := client.CreateExportTaskWithSubId(spreadsheetToken, "sheet", format, sheetID, userAccessToken)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "  任务 ID: %s\n", ticket)

		// 轮询等待任务完成
		fmt.Fprintf(os.Stderr, "正在等待导出完成...\n")
		fileToken, err := client.WaitExportTask(ticket, spreadsheetToken, userAccessToken, maxRetries)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "  导出文件 Token: %s\n", fileToken)

		// 下载导出文件
		fmt.Fprintf(os.Stderr, "正在下载文件...\n")
		if err := client.DownloadExportFile(fileToken, outputPath, userAccessToken); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "导出成功！\n")
		fmt.Fprintf(os.Stderr, "  保存路径: %s\n", outputPath)

		return nil
	},
}

func init() {
	sheetCmd.AddCommand(sheetExportCmd)
	sheetExportCmd.Flags().StringP("format", "f", "xlsx", "导出格式（xlsx/csv）")
	sheetExportCmd.Flags().String("sheet-id", "", "工作表 ID（CSV 格式必填）")
	sheetExportCmd.Flags().StringP("output", "o", "", "输出文件路径")
	sheetExportCmd.Flags().Int("max-retries", 30, "最大轮询重试次数")
	sheetExportCmd.Flags().String("user-access-token", "", "User Access Token")
}
