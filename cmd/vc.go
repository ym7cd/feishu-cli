package cmd

import (
	"github.com/spf13/cobra"
)

// vcCmd 视频会议父命令
var vcCmd = &cobra.Command{
	Use:   "vc",
	Short: "视频会议操作命令",
	Long: `视频会议相关操作，包括搜索历史会议、获取会议纪要等。

子命令:
  search    搜索历史会议记录
  notes     获取会议纪要

示例:
  feishu-cli vc search --start "2026-03-20" --end "2026-03-28"
  feishu-cli vc notes --meeting-id 69xxxx`,
}

func init() {
	rootCmd.AddCommand(vcCmd)
}
