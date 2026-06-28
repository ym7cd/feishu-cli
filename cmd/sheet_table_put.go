package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var sheetTablePutCmd = &cobra.Command{
	Use:   "table-put <spreadsheet_token> <sheet_id>",
	Short: "按列 dtype 类型保真写入整表（V3 typed cell）",
	Long: `按列 dtype 把 DataFrame 形状的 JSON 写入电子表格，保证数字/日期列不被当文本。

核心价值：日期列写 Excel 序列号 + 写后给该列设日期 formatter（yyyy/MM/dd），
飞书识别为「真日期」（可排序/可透视/ISNUMBER=TRUE），而非被当文本。
数字列保持数值类型，文本列用 formatter "@" 防止 ID/邮编等数字串被识别为数字。

输入格式（--sheets 或 --sheets-file，对齐 pandas to_json(orient="split") 形状）:
  {
    "sheets": [
      {
        "name": "可选 sheet 名（写入时以 <sheet_id> 为准）",
        "columns": ["id", "name", "amount", "date"],
        "data": [
          ["A001", "张三", 1200.50, "2024-01-15"],
          ["A002", "李四", 980,    "2024-02-20"]
        ],
        "dtypes":  {"amount": "float64", "date": "datetime64[ns]"},
        "formats": {"amount": "#,##0.00"}
      }
    ]
  }

dtype 映射（缺省/未知 → string）:
  - int*/uint*/float*/complex*   → number（interval* 除外，按文本处理）
  - bool/boolean                 → bool
  - datetime*                    → date（写 Excel 序列号 + 日期 formatter）
  - 其他（object/string 等）      → string（文本格式 @）

使用限制:
  - 当前仅支持单 sheet 一次写入（payload 含多 sheet 会报错，请逐个写入）
  - 就地写入 A1 起的矩形区域（覆盖前 N 行），不清除该区域之外的既有旧行，
    也不自动扩容：写入前确保目标 sheet 已有足够行（用 sheet add-rows 预扩容）
  - 单批 ≤ 5000 单元格，超出自动按行分批

示例:
  feishu-cli sheet table-put shtcnxxx 0b12 --sheets-file table.json
  feishu-cli sheet table-put shtcnxxx 0b12 --sheets '{"sheets":[{"columns":["x"],"data":[[1]],"dtypes":{"x":"int64"}}]}'
  feishu-cli sheet table-put shtcnxxx 0b12 --sheets-file table.json --header=false`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}
		spreadsheetToken := args[0]
		sheetID := args[1]
		sheetsJSON, _ := cmd.Flags().GetString("sheets")
		sheetsFile, _ := cmd.Flags().GetString("sheets-file")
		header, _ := cmd.Flags().GetBool("header")
		userIDType, _ := cmd.Flags().GetString("user-id-type")
		userAccessToken := resolveOptionalUserTokenWithFallback(cmd)

		raw, err := loadJSONInput(sheetsJSON, sheetsFile, "sheets", "sheets-file", "表格数据")
		if err != nil {
			return err
		}
		payload, err := client.ParseTablePutPayload([]byte(raw))
		if err != nil {
			return err
		}
		if len(payload.Sheets) > 1 {
			return fmt.Errorf("table-put 当前仅支持单 sheet（payload 含 %d 个，请逐个写入）", len(payload.Sheets))
		}
		spec := payload.Sheets[0]
		numCols := len(spec.Columns)
		if numCols < 1 {
			return fmt.Errorf("columns 为空")
		}

		// 构造元素矩阵（行 → 列 → 元素数组）。BuildTypedCell 对空值返回空文本元素，
		// 每格恒有一个元素（V3 不接受空元素数组，否则整批写入 500）。
		var rows [][][]*client.CellElement
		if header {
			hdr := make([][]*client.CellElement, numCols)
			for c, col := range spec.Columns {
				hdr[c] = []*client.CellElement{{Type: "text", Text: &client.TextElement{Text: col.Name}}}
			}
			rows = append(rows, hdr)
		}
		for ri, row := range spec.Rows {
			cells := make([][]*client.CellElement, numCols)
			for c := 0; c < numCols; c++ {
				cell, err := client.BuildTypedCell(spec.Columns[c], row[c])
				if err != nil {
					return fmt.Errorf("构造单元格失败（数据行 %d 列 %q）: %w", ri+1, spec.Columns[c].Name, err)
				}
				cells[c] = []*client.CellElement{cell}
			}
			rows = append(rows, cells)
		}
		if len(rows) == 0 {
			return fmt.Errorf("无数据可写入（data 为空且未写表头）")
		}

		// 前置 formatter：V3 单元格元素不支持 cell_styles，须先用 V2 style 接口给各列数据区
		// 设 formatter，再写值。顺序很关键——若先写值，飞书后端会对 text 元素做类型推断
		// （如 "007" 被存成数字 7、前导零丢失），formatter 后置只改显示、无法挽回；
		// 先设 @ / 日期 formatter 再写值，数字串才按文本保真、序列号才渲染为真日期（已实测）。
		dataStartRow := 1
		if header {
			dataStartRow = 2 // 表头占第 1 行，数据从第 2 行起
		}
		dataEndRow := len(rows)
		if dataEndRow >= dataStartRow {
			for c, col := range spec.Columns {
				formatter := client.FormatterForType(col)
				if formatter == "" {
					continue
				}
				colL := client.IndexToColumn(c)
				styleRange := fmt.Sprintf("%s!%s%d:%s%d", sheetID, colL, dataStartRow, colL, dataEndRow)
				style := &client.CellStyle{Formatter: formatter}
				if err := client.SetCellStyle(client.Context(), spreadsheetToken, styleRange, style, userAccessToken); err != nil {
					return fmt.Errorf("设置列 %q 格式（%s）失败: %w", col.Name, formatter, err)
				}
			}
		}

		// 分批写入（V3 单批 ≤ 5000 cell）
		const maxCellsPerWrite = 5000
		batchRows := maxCellsPerWrite / numCols
		if batchRows < 1 {
			batchRows = 1
		}
		lastColLetter := client.IndexToColumn(numCols - 1)
		for start := 0; start < len(rows); start += batchRows {
			end := start + batchRows
			if end > len(rows) {
				end = len(rows)
			}
			rng := fmt.Sprintf("%s!A%d:%s%d", sheetID, start+1, lastColLetter, end)
			vr := &client.ValueRangeV3{Range: rng, Values: rows[start:end]}
			if err := client.WriteCellsV3(client.Context(), spreadsheetToken, sheetID, []*client.ValueRangeV3{vr}, userIDType, userAccessToken); err != nil {
				return fmt.Errorf("写入第 %d-%d 行失败: %w", start+1, end, err)
			}
		}

		fmt.Printf("table-put 完成：sheet=%s，%d 列 × %d 行（含表头 %v）已写入\n", sheetID, numCols, len(rows), header)
		return nil
	},
}

func init() {
	sheetCmd.AddCommand(sheetTablePutCmd)
	sheetTablePutCmd.Flags().String("sheets", "", "表格数据 JSON（与 --sheets-file 二选一）")
	sheetTablePutCmd.Flags().String("sheets-file", "", "表格数据 JSON 文件路径")
	sheetTablePutCmd.Flags().Bool("header", true, "首行写入列名（默认开启，--header=false 关闭）")
	sheetTablePutCmd.Flags().String("user-id-type", "", "用户 ID 类型: open_id, union_id, user_id")
	sheetTablePutCmd.Flags().String("user-access-token", "", "User Access Token（可选，用于访问无 App 权限的表格）")
}
