package cmd

import (
	"github.com/spf13/cobra"
)

var bitableCmd = &cobra.Command{
	Use:     "bitable",
	Aliases: []string{"base"},
	Short:   "多维表格操作",
	Long: `多维表格操作命令组，支持数据表、字段、记录、视图的增删改查，
以及仪表盘、工作流、表单、角色等高级功能。

子命令:
  create          创建多维表格
  get             获取多维表格信息
  copy            复制多维表格
  tables          列出数据表
  create-table    创建数据表
  delete-table    删除数据表
  rename-table    重命名数据表
  fields          列出字段
  create-field    创建字段
  update-field    更新字段
  delete-field    删除字段
  records         搜索/列出记录
  get-record      获取单条记录
  add-record      创建记录
  add-records     批量创建记录
  update-record   更新记录
  update-records  批量更新记录
  delete-records  批量删除记录
  data-query      数据聚合查询
  views           列出视图
  create-view     创建视图
  delete-view     删除视图
  view-filter     视图过滤条件管理
  view-sort       视图排序管理
  view-group      视图分组管理
  dashboard       仪表盘管理
  dashboard-block 仪表盘 Block 管理
  workflow        工作流管理
  form            表单管理
  role            角色管理

示例:
  # 创建多维表格
  feishu-cli bitable create --name "项目管理"

  # 复制多维表格
  feishu-cli bitable copy <app_token> --name "副本"

  # 列出数据表
  feishu-cli bitable tables <app_token>

  # 搜索记录
  feishu-cli bitable records <app_token> <table_id>

  # 创建记录
  feishu-cli bitable add-record <app_token> <table_id> --fields '{"名称":"测试","金额":100}'

  # 仪表盘管理
  feishu-cli bitable dashboard list <app_token>

  # 工作流管理
  feishu-cli bitable workflow list <app_token>

  # 视图过滤条件
  feishu-cli bitable view-filter get <app_token> <table_id> <view_id>`,
}

func init() {
	rootCmd.AddCommand(bitableCmd)
}
