package cmd

import "github.com/spf13/cobra"

var approvalTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "审批任务相关命令",
	Long: `审批任务相关命令，用于查询当前 auth 登录用户待处理、已处理、已发起或抄送的审批任务。

示例:
  # 查询待我审批的任务
  feishu-cli approval task query --topic todo

  # 查询我发起的审批
  feishu-cli approval task query --topic started --output json`,
}

func init() {
	approvalCmd.AddCommand(approvalTaskCmd)
}
