package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 表单（Form）命令 ====================

var bitableFormCmd = &cobra.Command{
	Use:   "form",
	Short: "表单管理",
	Long: `表单管理命令组。

子命令:
  list   列出表单
  get    获取表单详情
  patch  更新表单`,
}

var bitableFormListCmd = &cobra.Command{
	Use:   "list <app_token> <table_id>",
	Short: "列出表单",
	Long:  "列出数据表中的所有表单",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		output, _ := cmd.Flags().GetString("output")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageToken, _ := cmd.Flags().GetString("page-token")
		userToken := resolveOptionalUserToken(cmd)

		forms, nextPageToken, err := client.ListBitableForms(appToken, tableID, pageSize, pageToken, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"forms": forms,
			}
			if nextPageToken != "" {
				result["page_token"] = nextPageToken
				result["has_more"] = true
			}
			return printJSON(result)
		}

		if len(forms) == 0 {
			fmt.Println("暂无表单")
			return nil
		}

		fmt.Printf("共 %d 个表单", len(forms))
		if nextPageToken != "" {
			fmt.Printf("（还有更多，page_token: %s）", nextPageToken)
		}
		fmt.Println("：")
		for i, f := range forms {
			name, _ := f["name"].(string)
			id, _ := f["form_id"].(string)
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, name, id)
		}
		return nil
	},
}

var bitableFormGetCmd = &cobra.Command{
	Use:   "get <app_token> <table_id> <form_id>",
	Short: "获取表单详情",
	Long:  "获取指定表单的详情信息",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		formID := args[2]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableForm(appToken, tableID, formID, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		name, _ := data["name"].(string)
		description, _ := data["description"].(string)
		fmt.Printf("Form ID: %s\n", formID)
		fmt.Printf("名称: %s\n", name)
		if description != "" {
			fmt.Printf("描述: %s\n", description)
		}
		return nil
	},
}

var bitableFormPatchCmd = &cobra.Command{
	Use:   "patch <app_token> <table_id> <form_id>",
	Short: "更新表单",
	Long:  "更新表单的名称或描述",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		tableID := args[1]
		formID := args[2]
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		reqBody := map[string]any{}
		if name != "" {
			reqBody["name"] = name
		}
		if description != "" {
			reqBody["description"] = description
		}
		if len(reqBody) == 0 {
			return fmt.Errorf("请至少指定 --name 或 --description")
		}

		data, err := client.PatchBitableForm(appToken, tableID, formID, reqBody, userToken)
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

func init() {
	bitableCmd.AddCommand(bitableFormCmd)

	bitableFormCmd.AddCommand(bitableFormListCmd)
	bitableFormCmd.AddCommand(bitableFormGetCmd)
	bitableFormCmd.AddCommand(bitableFormPatchCmd)

	// form list
	bitableFormListCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableFormListCmd.Flags().Int("page-size", 20, "每页数量")
	bitableFormListCmd.Flags().String("page-token", "", "分页标记")
	bitableFormListCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// form get
	bitableFormGetCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableFormGetCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// form patch
	bitableFormPatchCmd.Flags().StringP("name", "n", "", "表单名称")
	bitableFormPatchCmd.Flags().StringP("description", "d", "", "表单描述")
	bitableFormPatchCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableFormPatchCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
