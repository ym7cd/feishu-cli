package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var userSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "通过邮箱或手机号查询用户 ID",
	Long: `通过邮箱或手机号批量查询用户 ID。至少需要指定一个查询条件。

参数:
  --email     邮箱列表，逗号分隔
  --mobile    手机号列表，逗号分隔

示例:
  feishu-cli user search --email user@example.com
  feishu-cli user search --mobile +8613800138000
  feishu-cli user search --email a@example.com,b@example.com -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		emailStr, _ := cmd.Flags().GetString("email")
		mobileStr, _ := cmd.Flags().GetString("mobile")
		output, _ := cmd.Flags().GetString("output")

		if emailStr == "" && mobileStr == "" {
			return fmt.Errorf("至少需要指定 --email 或 --mobile")
		}

		var emails, mobiles []string
		if emailStr != "" {
			emails = splitAndTrim(emailStr)
		}
		if mobileStr != "" {
			mobiles = splitAndTrim(mobileStr)
		}

		result, err := client.BatchGetUserID(emails, mobiles)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(result)
		}

		if len(result) == 0 {
			fmt.Println("未找到匹配的用户")
			return nil
		}

		fmt.Printf("查询结果（共 %d 条）:\n\n", len(result))
		for i, item := range result {
			fmt.Printf("[%d] 用户 ID: %s\n", i+1, item.UserID)
			if item.Email != "" {
				fmt.Printf("    邮箱: %s\n", item.Email)
			}
			if item.Mobile != "" {
				fmt.Printf("    手机号: %s\n", item.Mobile)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	userCmd.AddCommand(userSearchCmd)
	userSearchCmd.Flags().String("email", "", "邮箱列表，逗号分隔")
	userSearchCmd.Flags().String("mobile", "", "手机号列表，逗号分隔")
	userSearchCmd.Flags().StringP("output", "o", "", "输出格式（json）")
}
