package cmd

import (
	"github.com/spf13/cobra"
)

var bitableCmd = &cobra.Command{
	Use:     "bitable",
	Aliases: []string{"base"},
	Short:   "多维表格（Base/Bitable）操作",
	Long: `多维表格操作命令组，底层调用飞书 base/v3 API（/open-apis/base/v3/bases/{base_token}/...）。

⚠️ 重要：本命令已从旧 bitable/v1 API 切换到 base/v3 API，支持的能力大幅增强：
  - 视图完整配置读写（filter/sort/group/visible-fields/timebar/card）
  - 记录 upsert + 修改历史
  - 角色完整 CRUD + 高级权限
  - 工作流查询

命令分为以下子组：
  bitable <create|get|copy>             基础：创建/获取/复制多维表格
  bitable table <list|get|create|...>   数据表 CRUD
  bitable field <list|get|create|...>   字段 CRUD + search-options
  bitable record <list|get|search|...>  记录 CRUD + upsert + batch + history
  bitable view <list|get|create|...>    视图 CRUD + rename
  bitable view-<filter|sort|group|visible-fields|timebar|card> <get|set>  视图配置
  bitable role <list|get|create|...>    角色 CRUD
  bitable advperm <enable|disable>      高级权限开关
  bitable data-query                    数据聚合查询
  bitable workflow list                 工作流列表

所有命令默认使用 User Access Token（先 feishu-cli auth login）。
使用 --base-token 传入多维表格 token（从 URL 里的 /base/{token} 片段获取）。

示例:
  feishu-cli bitable create --name "项目管理"
  feishu-cli bitable table list --base-token bscnxxxx
  feishu-cli bitable record list --base-token bscnxxxx --table-id tblxxx
  feishu-cli bitable view set-sort --base-token bscnxxxx --table-id tblxxx --view-id viewxxx --config '{"sort_info":[{"field_id":"fld1","desc":false}]}'`,
}

func init() {
	rootCmd.AddCommand(bitableCmd)
}
