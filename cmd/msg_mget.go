package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var msgMgetCmd = &cobra.Command{
	Use:   "mget",
	Short: "批量获取消息详情",
	Long: `批量获取多条消息的详情。

选项:
  --message-ids  消息 ID 列表（逗号分隔，必填）

示例:
  # 获取多条消息
  feishu-cli msg mget --message-ids om_xxx,om_yyy,om_zzz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		messageIDsStr, _ := cmd.Flags().GetString("message-ids")
		userToken := resolveOptionalUserToken(cmd)

		messageIDs := splitAndTrim(messageIDsStr)
		if len(messageIDs) == 0 {
			return fmt.Errorf("请提供至少一个消息 ID")
		}

		messages, err := client.BatchGetMessages(messageIDs, userToken)
		if err != nil {
			return err
		}

		return printJSON(messages)
	},
}

func init() {
	msgCmd.AddCommand(msgMgetCmd)
	msgMgetCmd.Flags().String("message-ids", "", "消息 ID 列表（逗号分隔）")
	msgMgetCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
	mustMarkFlagRequired(msgMgetCmd, "message-ids")
}
