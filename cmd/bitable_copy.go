package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 多维表格复制（Copy）命令 ====================

var bitableCopyCmd = &cobra.Command{
	Use:   "copy <app_token>",
	Short: "复制多维表格",
	Long: `复制多维表格到指定位置。

可选参数:
  --name            新的多维表格名称
  --folder-token    目标文件夹 Token
  --without-content 是否不复制内容（仅复制结构）`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		name, _ := cmd.Flags().GetString("name")
		folderToken, _ := cmd.Flags().GetString("folder-token")
		withoutContent, _ := cmd.Flags().GetBool("without-content")
		output, _ := cmd.Flags().GetString("output")
		userToken := resolveOptionalUserToken(cmd)

		reqBody := map[string]any{}
		if name != "" {
			reqBody["name"] = name
		}
		if folderToken != "" {
			reqBody["folder_token"] = folderToken
		}
		if withoutContent {
			reqBody["without_content"] = true
		}

		data, err := client.CopyBitableApp(appToken, reqBody, userToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(data)
		}

		if app, ok := data["app"].(map[string]any); ok {
			if newToken, ok := app["app_token"].(string); ok {
				fmt.Printf("复制成功！\n")
				fmt.Printf("  App Token: %s\n", newToken)
				if n, ok := app["name"].(string); ok {
					fmt.Printf("  名称: %s\n", n)
				}
				if u, ok := app["url"].(string); ok {
					fmt.Printf("  URL: %s\n", u)
				}
				return nil
			}
		}

		fmt.Println("复制成功！")
		return printJSON(data)
	},
}

func init() {
	bitableCmd.AddCommand(bitableCopyCmd)

	bitableCopyCmd.Flags().StringP("name", "n", "", "新的多维表格名称")
	bitableCopyCmd.Flags().String("folder-token", "", "目标文件夹 Token")
	bitableCopyCmd.Flags().Bool("without-content", false, "不复制内容（仅复制结构）")
	bitableCopyCmd.Flags().StringP("output", "o", "text", "输出格式: text, json")
	bitableCopyCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
