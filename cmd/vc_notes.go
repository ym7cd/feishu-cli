package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var vcNotesCmd = &cobra.Command{
	Use:   "notes",
	Short: "获取会议纪要",
	Long: `获取会议的纪要信息，包括会议详情和关联的妙记。

支持两种输入方式（互斥）:
  --meeting-id     通过会议 ID 获取会议详情
  --minute-token   通过妙记 token 直接获取妙记内容

示例:
  # 通过会议 ID 获取会议详情（包含纪要信息）
  feishu-cli vc notes --meeting-id 69xxxx

  # 通过妙记 token 获取妙记内容
  feishu-cli vc notes --minute-token obcnxxxx

  # JSON 格式输出
  feishu-cli vc notes --meeting-id 69xxxx -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token := resolveOptionalUserToken(cmd)

		meetingID, _ := cmd.Flags().GetString("meeting-id")
		minuteToken, _ := cmd.Flags().GetString("minute-token")
		output, _ := cmd.Flags().GetString("output")

		if meetingID == "" && minuteToken == "" {
			return fmt.Errorf("请指定 --meeting-id 或 --minute-token 之一")
		}
		if meetingID != "" && minuteToken != "" {
			return fmt.Errorf("--meeting-id 和 --minute-token 不能同时使用")
		}

		if minuteToken != "" {
			// 通过妙记 token 获取
			data, err := client.GetMinute(minuteToken, token)
			if err != nil {
				return err
			}

			if output == "json" {
				return printJSON(json.RawMessage(data))
			}

			return printMinuteInfo(data)
		}

		// 通过会议 ID 获取会议详情
		data, err := client.GetMeeting(meetingID, token)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(json.RawMessage(data))
		}

		return printMeetingDetail(data)
	},
}

// printMeetingDetail 格式化输出会议详情
func printMeetingDetail(data json.RawMessage) error {
	var detail struct {
		Meeting struct {
			MeetingID string `json:"meeting_id"`
			Topic     string `json:"topic"`
			MeetingNo string `json:"meeting_no"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
			HostUser  *struct {
				UserID string `json:"user_id"`
				Name   string `json:"name"`
			} `json:"host_user"`
		} `json:"meeting"`
	}

	if err := json.Unmarshal(data, &detail); err != nil {
		// 解析失败，直接打印原始 JSON
		fmt.Println(string(data))
		return nil
	}

	m := detail.Meeting
	topic := m.Topic
	if topic == "" {
		topic = "(无标题)"
	}

	fmt.Printf("会议详情:\n\n")
	fmt.Printf("  标题:      %s\n", topic)
	fmt.Printf("  会议 ID:   %s\n", m.MeetingID)
	if m.MeetingNo != "" {
		fmt.Printf("  会议号:    %s\n", m.MeetingNo)
	}
	if m.StartTime != "" {
		fmt.Printf("  开始时间:  %s\n", formatVCTime(m.StartTime))
	}
	if m.EndTime != "" {
		fmt.Printf("  结束时间:  %s\n", formatVCTime(m.EndTime))
	}
	if m.HostUser != nil && m.HostUser.Name != "" {
		fmt.Printf("  主持人:    %s\n", m.HostUser.Name)
	}

	return nil
}

// printMinuteInfo 格式化输出妙记信息
func printMinuteInfo(data json.RawMessage) error {
	var minute struct {
		Minute struct {
			Token    string `json:"token"`
			Title    string `json:"title"`
			URL      string `json:"url"`
			CreateTime string `json:"create_time"`
			Owner    *struct {
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
		fmt.Printf("  创建时间:  %s\n", formatVCTime(m.CreateTime))
	}
	if m.Owner != nil && m.Owner.Name != "" {
		fmt.Printf("  创建者:    %s\n", m.Owner.Name)
	}
	if m.Duration != "" {
		fmt.Printf("  时长:      %s\n", m.Duration)
	}

	return nil
}

func init() {
	vcCmd.AddCommand(vcNotesCmd)
	vcNotesCmd.Flags().String("meeting-id", "", "会议 ID")
	vcNotesCmd.Flags().String("minute-token", "", "妙记 Token")
	vcNotesCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	vcNotesCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
}
