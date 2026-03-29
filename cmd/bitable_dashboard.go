package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 仪表盘（Dashboard）命令 ====================

var bitableDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "仪表盘管理",
	Long: `仪表盘管理命令组，支持仪表盘的增删改查。

子命令:
  list    列出仪表盘
  get     获取仪表盘详情
  create  创建仪表盘
  update  更新仪表盘
  delete  删除仪表盘`,
}

var bitableDashboardListCmd = &cobra.Command{
	Use:   "list <app_token>",
	Short: "列出仪表盘",
	Long:  "列出多维表格中的所有仪表盘",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		output, _ := cmd.Flags().GetString("output")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageToken, _ := cmd.Flags().GetString("page-token")
		userToken := resolveOptionalUserToken(cmd)

		dashboards, nextPageToken, err := client.ListBitableDashboards(appToken, pageSize, pageToken, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"dashboards": dashboards,
			}
			if nextPageToken != "" {
				result["page_token"] = nextPageToken
				result["has_more"] = true
			}
			return printJSON(result)
		}

		if len(dashboards) == 0 {
			fmt.Println("暂无仪表盘")
			return nil
		}

		fmt.Printf("共 %d 个仪表盘", len(dashboards))
		if nextPageToken != "" {
			fmt.Printf("（还有更多，page_token: %s）", nextPageToken)
		}
		fmt.Println("：")
		for i, d := range dashboards {
			name, _ := d["name"].(string)
			id, _ := d["dashboard_id"].(string)
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, name, id)
		}
		return nil
	},
}

var bitableDashboardGetCmd = &cobra.Command{
	Use:   "get <app_token> <dashboard_id>",
	Short: "获取仪表盘详情",
	Long:  "获取指定仪表盘的详情信息",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		dashboardID := args[1]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableDashboard(appToken, dashboardID, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		name, _ := data["name"].(string)
		fmt.Printf("Dashboard ID: %s\n", dashboardID)
		fmt.Printf("名称: %s\n", name)
		return nil
	},
}

var bitableDashboardCreateCmd = &cobra.Command{
	Use:   "create <app_token>",
	Short: "复制创建仪表盘",
	Long: `通过复制现有仪表盘创建新的仪表盘。

注意：飞书 API 不支持直接创建空仪表盘，只能通过复制现有仪表盘来创建。

示例:
  feishu-cli bitable dashboard create APP_TOKEN --source-block-id blkXXX --name "新仪表盘"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		name, _ := cmd.Flags().GetString("name")
		sourceBlockID, _ := cmd.Flags().GetString("source-block-id")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		reqBody := map[string]any{}
		if name != "" {
			reqBody["name"] = name
		}
		if sourceBlockID != "" {
			reqBody["source_block_id"] = sourceBlockID
		}

		data, err := client.CreateBitableDashboard(appToken, reqBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		if id, ok := data["block_id"].(string); ok {
			fmt.Printf("复制成功！Dashboard Block ID: %s\n", id)
		} else {
			fmt.Println("复制成功！")
			return printJSON(data)
		}
		return nil
	},
}

var bitableDashboardUpdateCmd = &cobra.Command{
	Use:   "update <app_token> <dashboard_id>",
	Short: "更新仪表盘（API 不支持）",
	Long:  "飞书 API 不支持更新仪表盘，仅支持 list 和 copy（create）操作",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("飞书 API 不支持更新仪表盘（仅支持 list 和 copy 操作）")
	},
}

var bitableDashboardDeleteCmd = &cobra.Command{
	Use:   "delete <app_token> <dashboard_id>",
	Short: "删除仪表盘（API 不支持）",
	Long:  "飞书 API 不支持删除仪表盘，仅支持 list 和 copy（create）操作",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("飞书 API 不支持删除仪表盘（仅支持 list 和 copy 操作）")
	},
}

// ==================== 仪表盘 Block 命令 ====================

var bitableDashboardBlockCmd = &cobra.Command{
	Use:   "dashboard-block",
	Short: "仪表盘 Block 管理（API 不支持）",
	Long: `仪表盘 Block 管理命令组。

注意：飞书开放 API 不提供仪表盘 Block 的 CRUD 接口，
仪表盘仅支持 list（列出）和 copy（复制创建）操作。`,
}

var bitableDashboardBlockListCmd = &cobra.Command{
	Use:   "list <app_token> <dashboard_id>",
	Short: "列出仪表盘 Block",
	Long:  "列出仪表盘中的所有 Block",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		dashboardID := args[1]
		output, _ := cmd.Flags().GetString("output")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageToken, _ := cmd.Flags().GetString("page-token")
		userToken := resolveOptionalUserToken(cmd)

		blocks, nextPageToken, err := client.ListBitableDashboardBlocks(appToken, dashboardID, pageSize, pageToken, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"blocks": blocks,
			}
			if nextPageToken != "" {
				result["page_token"] = nextPageToken
				result["has_more"] = true
			}
			return printJSON(result)
		}

		if len(blocks) == 0 {
			fmt.Println("暂无 Block")
			return nil
		}

		fmt.Printf("共 %d 个 Block", len(blocks))
		if nextPageToken != "" {
			fmt.Printf("（还有更多，page_token: %s）", nextPageToken)
		}
		fmt.Println("：")
		for i, b := range blocks {
			name, _ := b["name"].(string)
			id, _ := b["block_id"].(string)
			blockType, _ := b["type"].(string)
			fmt.Printf("  %d. %s (类型: %s, ID: %s)\n", i+1, name, blockType, id)
		}
		return nil
	},
}

var bitableDashboardBlockGetCmd = &cobra.Command{
	Use:   "get <app_token> <dashboard_id> <block_id>",
	Short: "获取仪表盘 Block 详情",
	Long:  "获取仪表盘中指定 Block 的详情",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		dashboardID := args[1]
		blockID := args[2]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableDashboardBlock(appToken, dashboardID, blockID, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		name, _ := data["name"].(string)
		blockType, _ := data["type"].(string)
		fmt.Printf("Block ID: %s\n", blockID)
		fmt.Printf("名称: %s\n", name)
		fmt.Printf("类型: %s\n", blockType)
		return nil
	},
}

var bitableDashboardBlockCreateCmd = &cobra.Command{
	Use:   "create <app_token> <dashboard_id>",
	Short: "创建仪表盘 Block",
	Long:  "在仪表盘中创建 Block",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		dashboardID := args[1]
		name, _ := cmd.Flags().GetString("name")
		blockType, _ := cmd.Flags().GetString("type")
		dataConfigJSON, _ := cmd.Flags().GetString("data-config")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		reqBody := map[string]any{}
		if name != "" {
			reqBody["name"] = name
		}
		if blockType != "" {
			reqBody["type"] = blockType
		}
		if dataConfigJSON != "" {
			var dataConfig map[string]any
			if err := json.Unmarshal([]byte(dataConfigJSON), &dataConfig); err != nil {
				return fmt.Errorf("解析 data-config JSON 失败: %w", err)
			}
			reqBody["data_config"] = dataConfig
		}

		data, err := client.CreateBitableDashboardBlock(appToken, dashboardID, reqBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		if id, ok := data["block_id"].(string); ok {
			fmt.Printf("创建成功！Block ID: %s\n", id)
		} else {
			fmt.Println("创建成功！")
			return printJSON(data)
		}
		return nil
	},
}

var bitableDashboardBlockUpdateCmd = &cobra.Command{
	Use:   "update <app_token> <dashboard_id> <block_id>",
	Short: "更新仪表盘 Block",
	Long:  "更新仪表盘中的 Block",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		dashboardID := args[1]
		blockID := args[2]
		name, _ := cmd.Flags().GetString("name")
		blockType, _ := cmd.Flags().GetString("type")
		dataConfigJSON, _ := cmd.Flags().GetString("data-config")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		reqBody := map[string]any{}
		if name != "" {
			reqBody["name"] = name
		}
		if blockType != "" {
			reqBody["type"] = blockType
		}
		if dataConfigJSON != "" {
			var dataConfig map[string]any
			if err := json.Unmarshal([]byte(dataConfigJSON), &dataConfig); err != nil {
				return fmt.Errorf("解析 data-config JSON 失败: %w", err)
			}
			reqBody["data_config"] = dataConfig
		}

		if len(reqBody) == 0 {
			return fmt.Errorf("请至少指定 --name、--type 或 --data-config")
		}

		data, err := client.UpdateBitableDashboardBlock(appToken, dashboardID, blockID, reqBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		fmt.Println("更新成功！")
		return nil
	},
}

var bitableDashboardBlockDeleteCmd = &cobra.Command{
	Use:   "delete <app_token> <dashboard_id> <block_id>",
	Short: "删除仪表盘 Block",
	Long:  "删除仪表盘中的指定 Block",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		dashboardID := args[1]
		blockID := args[2]
		userToken := resolveOptionalUserToken(cmd)

		if err := client.DeleteBitableDashboardBlock(appToken, dashboardID, blockID, userToken); err != nil {
			return err
		}

		fmt.Println("删除成功")
		return nil
	},
}

func init() {
	bitableCmd.AddCommand(bitableDashboardCmd)
	bitableCmd.AddCommand(bitableDashboardBlockCmd)

	// === Dashboard 子命令 ===
	bitableDashboardCmd.AddCommand(bitableDashboardListCmd)
	bitableDashboardCmd.AddCommand(bitableDashboardGetCmd)
	bitableDashboardCmd.AddCommand(bitableDashboardCreateCmd)
	bitableDashboardCmd.AddCommand(bitableDashboardUpdateCmd)
	bitableDashboardCmd.AddCommand(bitableDashboardDeleteCmd)

	// dashboard list
	bitableDashboardListCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardListCmd.Flags().Int("page-size", 20, "每页数量")
	bitableDashboardListCmd.Flags().String("page-token", "", "分页标记")
	bitableDashboardListCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// dashboard get
	bitableDashboardGetCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardGetCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// dashboard create (copy)
	bitableDashboardCreateCmd.Flags().StringP("name", "n", "", "新仪表盘名称（必填）")
	bitableDashboardCreateCmd.Flags().String("source-block-id", "", "源仪表盘 block_id（必填，通过 dashboard list 获取）")
	bitableDashboardCreateCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardCreateCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
	mustMarkFlagRequired(bitableDashboardCreateCmd, "name")
	mustMarkFlagRequired(bitableDashboardCreateCmd, "source-block-id")

	// dashboard update（API 不支持，保留命令以给出友好错误提示）

	// dashboard delete（API 不支持，保留命令以给出友好错误提示）

	// === Dashboard Block 子命令 ===
	bitableDashboardBlockCmd.AddCommand(bitableDashboardBlockListCmd)
	bitableDashboardBlockCmd.AddCommand(bitableDashboardBlockGetCmd)
	bitableDashboardBlockCmd.AddCommand(bitableDashboardBlockCreateCmd)
	bitableDashboardBlockCmd.AddCommand(bitableDashboardBlockUpdateCmd)
	bitableDashboardBlockCmd.AddCommand(bitableDashboardBlockDeleteCmd)

	// dashboard-block list
	bitableDashboardBlockListCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardBlockListCmd.Flags().Int("page-size", 20, "每页数量")
	bitableDashboardBlockListCmd.Flags().String("page-token", "", "分页标记")
	bitableDashboardBlockListCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// dashboard-block get
	bitableDashboardBlockGetCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardBlockGetCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// dashboard-block create
	bitableDashboardBlockCreateCmd.Flags().StringP("name", "n", "", "Block 名称")
	bitableDashboardBlockCreateCmd.Flags().StringP("type", "t", "", "Block 类型")
	bitableDashboardBlockCreateCmd.Flags().String("data-config", "", "数据配置 JSON")
	bitableDashboardBlockCreateCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardBlockCreateCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// dashboard-block update
	bitableDashboardBlockUpdateCmd.Flags().StringP("name", "n", "", "Block 名称")
	bitableDashboardBlockUpdateCmd.Flags().StringP("type", "t", "", "Block 类型")
	bitableDashboardBlockUpdateCmd.Flags().String("data-config", "", "数据配置 JSON")
	bitableDashboardBlockUpdateCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableDashboardBlockUpdateCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// dashboard-block delete
	bitableDashboardBlockDeleteCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
