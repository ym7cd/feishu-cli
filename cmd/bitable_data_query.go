package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 数据聚合查询（Data Query）命令 ====================

var bitableDataQueryCmd = &cobra.Command{
	Use:   "data-query <app_token> <table_id>",
	Short: "数据聚合查询",
	Long: `对数据表执行聚合查询。

通过 records/search API 结合聚合参数进行数据分析。
请求体通过 --data（JSON 字符串）或 --data-file（JSON 文件）传入。

请求体示例（使用 records/search 的完整参数）:
  {
    "filter": {"conjunction":"and","conditions":[{"field_name":"状态","operator":"is","value":["进行中"]}]},
    "sort": [{"field_name":"创建时间","desc":true}],
    "field_names": ["状态","金额"],
    "page_size": 500
  }`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		dataJSON, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		// 加载请求体
		jsonStr := dataJSON
		if dataFile != "" {
			data, err := os.ReadFile(dataFile)
			if err != nil {
				return fmt.Errorf("读取数据文件失败: %w", err)
			}
			jsonStr = string(data)
		}
		if jsonStr == "" {
			return fmt.Errorf("请通过 --data 或 --data-file 提供查询参数 JSON")
		}

		var reqBody map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &reqBody); err != nil {
			return fmt.Errorf("解析查询参数 JSON 失败: %w", err)
		}

		// 解析为 BitableSearchOptions
		opts := client.BitableSearchOptions{}

		if ps, ok := reqBody["page_size"].(float64); ok {
			opts.PageSize = int(ps)
		}
		if pt, ok := reqBody["page_token"].(string); ok {
			opts.PageToken = pt
		}

		// 解析 filter
		if filterRaw, ok := reqBody["filter"]; ok {
			filterBytes, _ := json.Marshal(filterRaw)
			var filter client.BitableFilter
			if err := json.Unmarshal(filterBytes, &filter); err != nil {
				return fmt.Errorf("解析 filter 失败: %w", err)
			}
			opts.Filter = &filter
		}

		// 解析 sort
		if sortRaw, ok := reqBody["sort"]; ok {
			sortBytes, _ := json.Marshal(sortRaw)
			var sort []client.BitableSortItem
			if err := json.Unmarshal(sortBytes, &sort); err != nil {
				return fmt.Errorf("解析 sort 失败: %w", err)
			}
			opts.Sort = sort
		}

		// 解析 field_names
		if fnRaw, ok := reqBody["field_names"]; ok {
			fnBytes, _ := json.Marshal(fnRaw)
			var fieldNames []string
			if err := json.Unmarshal(fnBytes, &fieldNames); err != nil {
				return fmt.Errorf("解析 field_names 失败: %w", err)
			}
			opts.FieldNames = fieldNames
		}

		records, nextPageToken, total, err := client.SearchBitableRecords(appToken, tableID, opts, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"total":   total,
				"records": records,
			}
			if nextPageToken != "" {
				result["page_token"] = nextPageToken
				result["has_more"] = true
			}
			return printJSON(result)
		}

		fmt.Printf("共 %d 条记录", total)
		if nextPageToken != "" {
			fmt.Printf("（还有更多，page_token: %s）", nextPageToken)
		}
		fmt.Println()

		for i, r := range records {
			fmt.Printf("\n--- 记录 %d (ID: %s) ---\n", i+1, r.RecordID)
			for k, v := range r.Fields {
				fmt.Printf("  %s: %v\n", k, formatFieldValue(v))
			}
		}
		return nil
	},
}

func init() {
	bitableCmd.AddCommand(bitableDataQueryCmd)

	bitableDataQueryCmd.Flags().String("data", "", "查询参数 JSON")
	bitableDataQueryCmd.Flags().String("data-file", "", "查询参数 JSON 文件")
	bitableDataQueryCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDataQueryCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
