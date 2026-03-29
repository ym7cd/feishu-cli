package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 工作流（Workflow）命令 ====================

var bitableWorkflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "工作流管理",
	Long: `工作流管理命令组。

子命令:
  list     列出工作流
  get      获取工作流详情
  enable   启用工作流
  disable  禁用工作流`,
}

var bitableWorkflowListCmd = &cobra.Command{
	Use:   "list <app_token>",
	Short: "列出工作流",
	Long:  "列出多维表格中的所有工作流",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		output, _ := cmd.Flags().GetString("output")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageToken, _ := cmd.Flags().GetString("page-token")
		userToken := resolveOptionalUserToken(cmd)

		workflows, nextPageToken, err := client.ListBitableWorkflows(appToken, pageSize, pageToken, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"workflows": workflows,
			}
			if nextPageToken != "" {
				result["page_token"] = nextPageToken
				result["has_more"] = true
			}
			return printJSON(result)
		}

		if len(workflows) == 0 {
			fmt.Println("暂无工作流")
			return nil
		}

		fmt.Printf("共 %d 个工作流", len(workflows))
		if nextPageToken != "" {
			fmt.Printf("（还有更多，page_token: %s）", nextPageToken)
		}
		fmt.Println("：")
		for i, w := range workflows {
			name, _ := w["name"].(string)
			id, _ := w["workflow_id"].(string)
			status, _ := w["status"].(string)
			fmt.Printf("  %d. %s (状态: %s, ID: %s)\n", i+1, name, status, id)
		}
		return nil
	},
}

var bitableWorkflowGetCmd = &cobra.Command{
	Use:   "get <app_token> <workflow_id>",
	Short: "获取工作流详情",
	Long:  "获取指定工作流的详情信息",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		workflowID := args[1]
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		data, err := client.GetBitableWorkflow(appToken, workflowID, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		name, _ := data["name"].(string)
		status, _ := data["status"].(string)
		fmt.Printf("Workflow ID: %s\n", workflowID)
		fmt.Printf("名称: %s\n", name)
		fmt.Printf("状态: %s\n", status)
		return nil
	},
}

var bitableWorkflowEnableCmd = &cobra.Command{
	Use:   "enable <app_token> <workflow_id>",
	Short: "启用工作流",
	Long:  "启用指定工作流",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		workflowID := args[1]
		userToken := resolveOptionalUserToken(cmd)

		if err := client.EnableBitableWorkflow(appToken, workflowID, userToken); err != nil {
			return err
		}

		fmt.Println("工作流已启用")
		return nil
	},
}

var bitableWorkflowDisableCmd = &cobra.Command{
	Use:   "disable <app_token> <workflow_id>",
	Short: "禁用工作流",
	Long:  "禁用指定工作流",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		workflowID := args[1]
		userToken := resolveOptionalUserToken(cmd)

		if err := client.DisableBitableWorkflow(appToken, workflowID, userToken); err != nil {
			return err
		}

		fmt.Println("工作流已禁用")
		return nil
	},
}

func init() {
	bitableCmd.AddCommand(bitableWorkflowCmd)

	bitableWorkflowCmd.AddCommand(bitableWorkflowListCmd)
	bitableWorkflowCmd.AddCommand(bitableWorkflowGetCmd)
	bitableWorkflowCmd.AddCommand(bitableWorkflowEnableCmd)
	bitableWorkflowCmd.AddCommand(bitableWorkflowDisableCmd)

	// workflow list
	bitableWorkflowListCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableWorkflowListCmd.Flags().Int("page-size", 20, "每页数量")
	bitableWorkflowListCmd.Flags().String("page-token", "", "分页标记")
	bitableWorkflowListCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// workflow get
	bitableWorkflowGetCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableWorkflowGetCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// workflow enable
	bitableWorkflowEnableCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// workflow disable
	bitableWorkflowDisableCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
