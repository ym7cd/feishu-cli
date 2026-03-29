package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/spf13/cobra"
)

// ==================== 高级权限（Advanced Permissions）命令 ====================

var bitableAdvpermCmd = &cobra.Command{
	Use:   "advperm",
	Short: "高级权限管理",
	Long: `多维表格高级权限管理。

子命令:
  enable   启用高级权限
  disable  禁用高级权限

启用高级权限后，可以通过角色（role）对不同协作者设置不同的数据访问权限。

示例:
  # 启用高级权限
  feishu-cli bitable advperm enable APP_TOKEN

  # 禁用高级权限
  feishu-cli bitable advperm disable APP_TOKEN`,
}

var bitableAdvpermEnableCmd = &cobra.Command{
	Use:   "enable <app_token>",
	Short: "启用高级权限",
	Long: `启用多维表格的高级权限。

启用后可以创建角色并分配不同的数据访问权限。

示例:
  feishu-cli bitable advperm enable APP_TOKEN`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		userToken := resolveOptionalUserToken(cmd)

		if err := client.UpdateBitableAdvancedPermissions(appToken, true, userToken); err != nil {
			return err
		}

		fmt.Println("高级权限已启用")
		return nil
	},
}

var bitableAdvpermDisableCmd = &cobra.Command{
	Use:   "disable <app_token>",
	Short: "禁用高级权限",
	Long: `禁用多维表格的高级权限。

禁用后所有角色权限将失效，协作者恢复为统一的权限级别。

示例:
  feishu-cli bitable advperm disable APP_TOKEN`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appToken := args[0]
		userToken := resolveOptionalUserToken(cmd)

		if err := client.UpdateBitableAdvancedPermissions(appToken, false, userToken); err != nil {
			return err
		}

		fmt.Println("高级权限已禁用")
		return nil
	},
}

func init() {
	bitableCmd.AddCommand(bitableAdvpermCmd)

	bitableAdvpermCmd.AddCommand(bitableAdvpermEnableCmd)
	bitableAdvpermCmd.AddCommand(bitableAdvpermDisableCmd)

	// enable
	bitableAdvpermEnableCmd.Flags().String("user-access-token", "", "User Access Token（可选）")

	// disable
	bitableAdvpermDisableCmd.Flags().String("user-access-token", "", "User Access Token（可选）")
}
