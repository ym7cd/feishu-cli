package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

// minutesCmd 妙记父命令
var minutesCmd = &cobra.Command{
	Use:   "minutes",
	Short: "妙记操作命令",
	Long: `妙记相关操作，包括获取妙记信息等。

子命令:
  get    获取妙记信息

示例:
  feishu-cli minutes get obcnxxxx`,
}

var minutesGetCmd = &cobra.Command{
	Use:   "get <minute_token>",
	Short: "获取妙记信息",
	Long: `通过妙记 Token 获取妙记基础信息，包括标题、链接、创建时间、时长等。

参数:
  minute_token  妙记 Token

示例:
  # 获取妙记信息
  feishu-cli minutes get obcnxxxx

  # JSON 格式输出
  feishu-cli minutes get obcnxxxx -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token := resolveOptionalUserToken(cmd)
		minuteToken := args[0]
		output, _ := cmd.Flags().GetString("output")

		data, err := client.GetMinute(minuteToken, token)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(json.RawMessage(data))
		}

		// 格式化输出
		var minute struct {
			Minute struct {
				Token      string `json:"token"`
				Title      string `json:"title"`
				URL        string `json:"url"`
				CreateTime string `json:"create_time"`
				Owner      *struct {
					UserID string `json:"user_id"`
					Name   string `json:"name"`
				} `json:"owner"`
				Duration string `json:"duration"`
			} `json:"minute"`
		}

		if err := json.Unmarshal(data, &minute); err != nil {
			// 解析失败，直接打印原始 JSON
			fmt.Println(string(data))
			return nil
		}

		m := minute.Minute
		title := m.Title
		if title == "" {
			title = "(无标题)"
		}

		fmt.Printf("妙记信息:\n\n")
		fmt.Printf("  标题:      %s\n", title)
		fmt.Printf("  Token:     %s\n", m.Token)
		if m.URL != "" {
			fmt.Printf("  链接:      %s\n", m.URL)
		}
		if m.CreateTime != "" {
			fmt.Printf("  创建时间:  %s\n", formatMinuteTime(m.CreateTime))
		}
		if m.Owner != nil && m.Owner.Name != "" {
			fmt.Printf("  创建者:    %s\n", m.Owner.Name)
		}
		if m.Duration != "" {
			fmt.Printf("  时长:      %s\n", m.Duration)
		}

		return nil
	},
}

// formatMinuteTime 尝试将 Unix 秒字符串转为可读时间
func formatMinuteTime(ts string) string {
	// 复用 vc_search.go 中的 formatVCTime
	return formatVCTime(ts)
}

func init() {
	rootCmd.AddCommand(minutesCmd)
	minutesCmd.AddCommand(minutesGetCmd)
	minutesGetCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	minutesGetCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
}
