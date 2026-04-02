package cmd

import "github.com/spf13/cobra"

var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "审批相关命令",
	Long: `审批相关命令，用于查看审批定义、实例和任务。

当前已提供：
  - 审批定义查询（approval get）
  - 审批任务查询（approval task query）

示例:
  # 查看审批定义详情
  feishu-cli approval get <approval_code>

  # 查看当前登录用户的待我审批任务
  feishu-cli approval task query --topic todo`,
}

func init() {
	rootCmd.AddCommand(approvalCmd)
}
