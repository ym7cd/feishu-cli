package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var vcSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "搜索历史会议记录",
	Long: `搜索已结束的历史会议记录。

可选参数:
  --start          起始日期，格式 YYYY-MM-DD（默认 7 天前）
  --end            结束日期，格式 YYYY-MM-DD（默认今天）
  --meeting-no     按会议号过滤
  --meeting-status 会议状态（1=进行中, 2=未开始, 3=已结束，默认 3）
  --page-size      每页数量
  --page-token     分页标记
  --output, -o     输出格式（json）

示例:
  # 搜索最近 7 天已结束的会议
  feishu-cli vc search

  # 按时间范围搜索
  feishu-cli vc search --start "2026-03-20" --end "2026-03-28"

  # 按会议号搜索
  feishu-cli vc search --meeting-no "123456789"

  # JSON 格式输出
  feishu-cli vc search -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token := resolveOptionalUserToken(cmd)

		startStr, _ := cmd.Flags().GetString("start")
		endStr, _ := cmd.Flags().GetString("end")
		meetingNo, _ := cmd.Flags().GetString("meeting-no")
		meetingStatus, _ := cmd.Flags().GetInt("meeting-status")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageToken, _ := cmd.Flags().GetString("page-token")
		output, _ := cmd.Flags().GetString("output")

		// 解析时间范围
		now := time.Now()
		loc := now.Location()

		var startTime, endTime time.Time

		if startStr != "" {
			t, err := time.ParseInLocation("2006-01-02", startStr, loc)
			if err != nil {
				return fmt.Errorf("解析起始日期失败（格式应为 YYYY-MM-DD）: %w", err)
			}
			startTime = t
		} else {
			// 默认 7 天前
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -7)
		}

		if endStr != "" {
			t, err := time.ParseInLocation("2006-01-02", endStr, loc)
			if err != nil {
				return fmt.Errorf("解析结束日期失败（格式应为 YYYY-MM-DD）: %w", err)
			}
			// 结束日期设为当天 23:59:59，使查询包含该天
			endTime = t.Add(24*time.Hour - time.Second)
		} else {
			// 默认今天结束
			endTime = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
		}

		meetingList, nextToken, hasMore, err := client.SearchMeetings(
			startTime.Unix(),
			endTime.Unix(),
			meetingStatus,
			meetingNo,
			pageSize,
			pageToken,
			token,
		)
		if err != nil {
			return err
		}

		if output == "json" {
			result := map[string]any{
				"meeting_list": json.RawMessage(meetingList),
				"has_more":     hasMore,
			}
			if nextToken != "" {
				result["page_token"] = nextToken
			}
			return printJSON(result)
		}

		// 解析会议列表用于格式化输出
		var meetings []struct {
			MeetingID string `json:"meeting_id"`
			Topic     string `json:"topic"`
			MeetingNo string `json:"meeting_no"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		}
		if meetingList != nil {
			if err := json.Unmarshal(meetingList, &meetings); err != nil {
				// 如果解析失败，直接打印原始 JSON
				fmt.Println(string(meetingList))
				return nil
			}
		}

		if len(meetings) == 0 {
			fmt.Printf("在 %s ~ %s 期间无会议记录\n",
				startTime.Format("2006-01-02"),
				endTime.Format("2006-01-02"))
			return nil
		}

		fmt.Printf("会议列表（%s ~ %s，共 %d 个）:\n\n",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"),
			len(meetings))

		for i, m := range meetings {
			topic := m.Topic
			if topic == "" {
				topic = "(无标题)"
			}

			fmt.Printf("[%d] %s\n", i+1, topic)
			fmt.Printf("    会议 ID:   %s\n", m.MeetingID)
			if m.MeetingNo != "" {
				fmt.Printf("    会议号:    %s\n", m.MeetingNo)
			}

			// 尝试解析 Unix 时间戳为可读时间
			if m.StartTime != "" {
				fmt.Printf("    开始时间:  %s\n", formatVCTime(m.StartTime))
			}
			if m.EndTime != "" {
				fmt.Printf("    结束时间:  %s\n", formatVCTime(m.EndTime))
			}
			fmt.Println()
		}

		if hasMore {
			fmt.Printf("还有更多会议，使用 --page-token %s 获取下一页\n", nextToken)
		}

		return nil
	},
}

// formatVCTime 尝试将 Unix 秒字符串转为可读时间
func formatVCTime(ts string) string {
	var sec int64
	if _, err := fmt.Sscanf(ts, "%d", &sec); err == nil && sec > 0 {
		return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
	}
	return ts
}

func init() {
	vcCmd.AddCommand(vcSearchCmd)
	vcSearchCmd.Flags().String("start", "", "起始日期，格式 YYYY-MM-DD（默认 7 天前）")
	vcSearchCmd.Flags().String("end", "", "结束日期，格式 YYYY-MM-DD（默认今天）")
	vcSearchCmd.Flags().String("meeting-no", "", "按会议号过滤")
	vcSearchCmd.Flags().Int("meeting-status", 3, "会议状态（1=进行中, 2=未开始, 3=已结束）")
	vcSearchCmd.Flags().Int("page-size", 0, "每页数量")
	vcSearchCmd.Flags().String("page-token", "", "分页标记")
	vcSearchCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	vcSearchCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
}
