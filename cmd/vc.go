package cmd

import (
	"github.com/spf13/cobra"
)

// vcCmd 视频会议父命令
var vcCmd = &cobra.Command{
	Use:   "vc",
	Short: "视频会议操作命令",
	Long: `视频会议相关操作，包括搜索历史会议、获取会议纪要、查询会议录制等。

所有 vc 子命令均需 User Access Token（先 feishu-cli auth login）。

子命令:
  search     搜索历史会议记录（支持 query/时间/组织者/参会者/会议室多维过滤）
  notes      获取会议纪要（支持 meeting-ids / minute-tokens / calendar-event-ids 三路径）
  recording  查询会议录制并提取 minute_token

示例:
  feishu-cli vc search --query "周会" --start 2026-03-01
  feishu-cli vc notes --minute-tokens obcnxxxx --with-artifacts
  feishu-cli vc recording --meeting-ids 69xxxx`,
}

func init() {
	rootCmd.AddCommand(vcCmd)
}
