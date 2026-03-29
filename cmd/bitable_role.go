package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 角色（Role）命令 ====================

var bitableRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "角色管理",
	Long: `角色管理命令组。

子命令:
  list    列出角色
  create  创建角色
  delete  删除角色`,
}

var bitableRoleListCmd = &cobra.Command{
	Use:   "list <app_token>",
	Short: "列出角色",
	Long:  "列出多维表格中的所有角色",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		output, _ := cmd.Flags().GetString("output")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageToken, _ := cmd.Flags().GetString("page-token")
		userToken := resolveOptionalUserToken(cmd)

		roles, nextPageToken, err := client.ListBitableRoles(appToken, pageSize, pageToken, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"roles": roles,
			}
			if nextPageToken != "" {
				result["page_token"] = nextPageToken
				result["has_more"] = true
			}
			return printJSON(result)
		}

		if len(roles) == 0 {
			fmt.Println("暂无角色")
			return nil
		}

		fmt.Printf("共 %d 个角色", len(roles))
		if nextPageToken != "" {
			fmt.Printf("（还有更多，page_token: %s）", nextPageToken)
		}
		fmt.Println("：")
		for i, r := range roles {
			name, _ := r["role_name"].(string)
			id, _ := r["role_id"].(string)
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, name, id)
		}
		return nil
	},
}

var bitableRoleCreateCmd = &cobra.Command{
	Use:   "create <app_token>",
	Short: "创建角色",
	Long: `创建角色。

通过 --name 指定角色名称，--config 指定角色配置 JSON。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		name, _ := cmd.Flags().GetString("name")
		configJSON, _ := cmd.Flags().GetString("config")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		reqBody := map[string]any{}
		if name != "" {
			reqBody["role_name"] = name
		}
		if configJSON != "" {
			var cfg map[string]any
			if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
				return fmt.Errorf("解析 config JSON 失败: %w", err)
			}
			// 合并 config 字段到请求体
			for k, v := range cfg {
				reqBody[k] = v
			}
		}

		data, err := client.CreateBitableRole(appToken, reqBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		if id, ok := data["role_id"].(string); ok {
			fmt.Printf("创建成功！Role ID: %s\n", id)
		} else {
			fmt.Println("创建成功！")
			return printJSON(data)
		}
		return nil
	},
}

var bitableRoleDeleteCmd = &cobra.Command{
	Use:   "delete <app_token> <role_id>",
	Short: "删除角色",
	Long:  "删除指定角色",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		roleID := args[1]
		userToken := resolveOptionalUserToken(cmd)

		if err := client.DeleteBitableRole(appToken, roleID, userToken); err != nil {
			return err
		}

		fmt.Println("删除成功")
		return nil
	},
}

func init() {
	bitableCmd.AddCommand(bitableRoleCmd)

	bitableRoleCmd.AddCommand(bitableRoleListCmd)
	bitableRoleCmd.AddCommand(bitableRoleCreateCmd)
	bitableRoleCmd.AddCommand(bitableRoleDeleteCmd)

	// role list
	bitableRoleListCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableRoleListCmd.Flags().Int("page-size", 20, "每页数量")
	bitableRoleListCmd.Flags().String("page-token", "", "分页标记")
	bitableRoleListCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// role create
	bitableRoleCreateCmd.Flags().StringP("name", "n", "", "角色名称")
	bitableRoleCreateCmd.Flags().String("config", "", "角色配置 JSON")
	bitableRoleCreateCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableRoleCreateCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
	mustMarkFlagRequired(bitableRoleCreateCmd, "name")

	// role delete
	bitableRoleDeleteCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
