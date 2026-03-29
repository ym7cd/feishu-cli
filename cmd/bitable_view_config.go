package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 视图过滤（View Filter）命令 ====================

var bitableViewFilterCmd = &cobra.Command{
	Use:   "view-filter",
	Short: "视图过滤条件管理",
	Long: `视图过滤条件管理命令组。

子命令:
  get  获取视图过滤条件
  set  设置视图过滤条件`,
}

var bitableViewFilterGetCmd = &cobra.Command{
	Use:   "get <app_token> <table_id> <view_id>",
	Short: "获取视图过滤条件",
	Long:  "获取视图的过滤条件配置",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		viewID := args[2]
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableViewConfig(appToken, tableID, viewID, "filter_info", userToken)
		if err != nil {
			return err
		}

		return printJSON(data)
	},
}

var bitableViewFilterSetCmd = &cobra.Command{
	Use:   "set <app_token> <table_id> <view_id>",
	Short: "设置视图过滤条件",
	Long: `设置视图的过滤条件配置。

通过 --config 传入 JSON 字符串或 --config-file 传入 JSON 文件。`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		viewID := args[2]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		configBody, err := loadViewConfigInput(cmd)
		if err != nil {
			return err
		}

		data, err := client.SetBitableViewConfig(appToken, tableID, viewID, "filter_info", configBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		fmt.Println("设置成功！")
		return nil
	},
}

// ==================== 视图排序（View Sort）命令 ====================

var bitableViewSortCmd = &cobra.Command{
	Use:   "view-sort",
	Short: "视图排序管理",
	Long: `视图排序管理命令组。

子命令:
  get  获取视图排序配置
  set  设置视图排序配置`,
}

var bitableViewSortGetCmd = &cobra.Command{
	Use:   "get <app_token> <table_id> <view_id>",
	Short: "获取视图排序配置",
	Long:  "获取视图的排序配置",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		viewID := args[2]
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableViewConfig(appToken, tableID, viewID, "sort_info", userToken)
		if err != nil {
			return err
		}

		return printJSON(data)
	},
}

var bitableViewSortSetCmd = &cobra.Command{
	Use:   "set <app_token> <table_id> <view_id>",
	Short: "设置视图排序配置",
	Long: `设置视图的排序配置。

通过 --config 传入 JSON 字符串或 --config-file 传入 JSON 文件。`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		viewID := args[2]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		configBody, err := loadViewConfigInput(cmd)
		if err != nil {
			return err
		}

		data, err := client.SetBitableViewConfig(appToken, tableID, viewID, "sort_info", configBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		fmt.Println("设置成功！")
		return nil
	},
}

// ==================== 视图分组（View Group）命令 ====================

var bitableViewGroupCmd = &cobra.Command{
	Use:   "view-group",
	Short: "视图分组管理",
	Long: `视图分组管理命令组。

子命令:
  get  获取视图分组配置
  set  设置视图分组配置`,
}

var bitableViewGroupGetCmd = &cobra.Command{
	Use:   "get <app_token> <table_id> <view_id>",
	Short: "获取视图分组配置",
	Long:  "获取视图的分组配置",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		viewID := args[2]
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableViewConfig(appToken, tableID, viewID, "group_info", userToken)
		if err != nil {
			return err
		}

		return printJSON(data)
	},
}

var bitableViewGroupSetCmd = &cobra.Command{
	Use:   "set <app_token> <table_id> <view_id>",
	Short: "设置视图分组配置",
	Long: `设置视图的分组配置。

通过 --config 传入 JSON 字符串或 --config-file 传入 JSON 文件。`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		viewID := args[2]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		configBody, err := loadViewConfigInput(cmd)
		if err != nil {
			return err
		}

		data, err := client.SetBitableViewConfig(appToken, tableID, viewID, "group_info", configBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		fmt.Println("设置成功！")
		return nil
	},
}

// loadViewConfigInput 从 --config 或 --config-file 加载视图配置 JSON
func loadViewConfigInput(cmd *cobra.Command) (any, error) {
	configJSON, _ := cmd.Flags().GetString("config")
	configFile, _ := cmd.Flags().GetString("config-file")

	if configJSON == "" && configFile == "" {
		return nil, fmt.Errorf("请通过 --config 或 --config-file 提供配置 JSON")
	}
	if configJSON != "" && configFile != "" {
		return nil, fmt.Errorf("--config 和 --config-file 不能同时使用")
	}

	jsonStr := configJSON
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		jsonStr = string(data)
	}

	var body any
	if err := json.Unmarshal([]byte(jsonStr), &body); err != nil {
		return nil, fmt.Errorf("解析配置 JSON 失败: %w", err)
	}

	return body, nil
}

func init() {
	bitableCmd.AddCommand(bitableViewFilterCmd)
	bitableCmd.AddCommand(bitableViewSortCmd)
	bitableCmd.AddCommand(bitableViewGroupCmd)

	// === View Filter 子命令 ===
	bitableViewFilterCmd.AddCommand(bitableViewFilterGetCmd)
	bitableViewFilterCmd.AddCommand(bitableViewFilterSetCmd)

	addViewConfigFlags(bitableViewFilterGetCmd)
	addViewConfigSetFlags(bitableViewFilterSetCmd)

	// === View Sort 子命令 ===
	bitableViewSortCmd.AddCommand(bitableViewSortGetCmd)
	bitableViewSortCmd.AddCommand(bitableViewSortSetCmd)

	addViewConfigFlags(bitableViewSortGetCmd)
	addViewConfigSetFlags(bitableViewSortSetCmd)

	// === View Group 子命令 ===
	bitableViewGroupCmd.AddCommand(bitableViewGroupGetCmd)
	bitableViewGroupCmd.AddCommand(bitableViewGroupSetCmd)

	addViewConfigFlags(bitableViewGroupGetCmd)
	addViewConfigSetFlags(bitableViewGroupSetCmd)
}

// addViewConfigFlags 为 get 命令添加通用 flags
func addViewConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	cmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}

// addViewConfigSetFlags 为 set 命令添加通用 flags
func addViewConfigSetFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "", "配置 JSON 字符串")
	cmd.Flags().String("config-file", "", "配置 JSON 文件路径")
	cmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	cmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
