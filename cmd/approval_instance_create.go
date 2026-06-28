package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var approvalInstanceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建审批实例（发起审批）",
	Long: `创建一条审批实例。需要 tenant_access_token + scope approval:approval。

参数:
  --approval-code   审批定义 code（必填）
  --user-id         发起人 ID（必填，open_id 或 user_id，对应 --user-id-type）
  --form            表单数据 JSON 字符串（与 --form-file 二选一）
  --form-file       表单数据 JSON 文件路径（与 --form 二选一）
  --user-id-type    open_id（默认）/user_id（v4/instances endpoint 不支持 union_id）
  --department-id   发起人部门 ID（可选）
  --open-chat-id    审批结果推送到的群（可选）
  --node-approver   节点指定审批人 JSON，如 [{"node_id":"n1","value":["ou_xxx"]}]（可选，与 --node-approver-file 二选一）
  --node-cc         节点指定抄送人 JSON，格式同 --node-approver（可选，与 --node-cc-file 二选一）
  --output, -o      输出格式：json

示例:
  # 通过文件提交表单
  feishu-cli approval instance create --approval-code <code> --user-id ou_xxx --form-file form.json

	# 直接传 JSON 字符串
  feishu-cli approval instance create --approval-code <code> --user-id ou_xxx \
    --form '[{"id":"widget_1","type":"input","value":"内容"}]'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		approvalCode, _ := cmd.Flags().GetString("approval-code")
		if err := validateApprovalCode(approvalCode); err != nil {
			return err
		}

		userID, _ := cmd.Flags().GetString("user-id")
		if strings.TrimSpace(userID) == "" {
			return fmt.Errorf("--user-id 不能为空")
		}

		formInline, _ := cmd.Flags().GetString("form")
		formFile, _ := cmd.Flags().GetString("form-file")
		formData, err := loadJSONInput(formInline, formFile, "form", "form-file", "表单数据")
		if err != nil {
			return err
		}
		// 校验为合法 JSON 数组，飞书 form 必须是数组（[{"id":...}]），
		// 否则服务端会返回不友好的参数错误。
		var arr []any
		if err := json.Unmarshal([]byte(formData), &arr); err != nil {
			return fmt.Errorf("表单数据必须是 JSON 数组，解析失败: %w", err)
		}

		userIDType, _ := cmd.Flags().GetString("user-id-type")
		if err := validateApprovalCreateUserIDType(userIDType); err != nil {
			return err
		}
		departmentID, _ := cmd.Flags().GetString("department-id")
		openChatID, _ := cmd.Flags().GetString("open-chat-id")
		nodeApproverRaw, nodeCCRaw, err := loadNodeApproverCC(cmd)
		if err != nil {
			return err
		}
		output, _ := cmd.Flags().GetString("output")

		opts := client.CreateApprovalInstanceOptions{
			ApprovalCode:           approvalCode,
			UserID:                 userID,
			Form:                   formData,
			UserIDType:             userIDType,
			DepartmentID:           departmentID,
			OpenChatID:             openChatID,
			NodeApproverUserIDList: nodeApproverRaw,
			NodeCCUserIDList:       nodeCCRaw,
		}

		if err := config.Validate(); err != nil {
			return err
		}

		result, err := client.CreateApprovalInstance(opts, "")
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(result)
		}

		fmt.Printf("审批实例已创建\n  instance_code: %s\n", result.InstanceCode)
		return nil
	},
}

func init() {
	approvalInstanceCmd.AddCommand(approvalInstanceCreateCmd)

	approvalInstanceCreateCmd.Flags().String("approval-code", "", "审批定义 code（必填）")
	approvalInstanceCreateCmd.Flags().String("user-id", "", "发起人用户 ID（必填）")
	approvalInstanceCreateCmd.Flags().String("form", "", "表单数据 JSON 字符串（与 --form-file 二选一）")
	approvalInstanceCreateCmd.Flags().String("form-file", "", "表单数据 JSON 文件路径")
	approvalInstanceCreateCmd.Flags().String("user-id-type", "open_id", "用户 ID 类型：open_id/user_id（endpoint 不支持 union_id）")
	approvalInstanceCreateCmd.Flags().String("department-id", "", "发起人部门 ID（可选）")
	approvalInstanceCreateCmd.Flags().String("open-chat-id", "", "结果推送的群 ID（可选）")
	approvalInstanceCreateCmd.Flags().String("node-approver", "", "节点指定审批人 JSON，如 [{\"node_id\":\"n1\",\"value\":[\"ou_xxx\"]}]（可选）")
	approvalInstanceCreateCmd.Flags().String("node-approver-file", "", "节点指定审批人 JSON 文件路径")
	approvalInstanceCreateCmd.Flags().String("node-cc", "", "节点指定抄送人 JSON，格式同 --node-approver（可选）")
	approvalInstanceCreateCmd.Flags().String("node-cc-file", "", "节点指定抄送人 JSON 文件路径")
	approvalInstanceCreateCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	mustMarkFlagRequired(approvalInstanceCreateCmd, "approval-code", "user-id")
}

// loadNodeApproverCC 读取 --node-approver/--node-approver-file 与 --node-cc/--node-cc-file，
// 返回节点指定审批人/抄送人的 JSON 原文；两者都为空时对应返回 nil（不写入 body）。
// 格式：[{"node_id":"<节点ID>","value":["ou_xxx", ...]}]。
func loadNodeApproverCC(cmd *cobra.Command) (json.RawMessage, json.RawMessage, error) {
	approverRaw, err := loadNodeJSONArrayFlag(cmd, "node-approver", "node-approver-file", "节点指定审批人")
	if err != nil {
		return nil, nil, err
	}
	ccRaw, err := loadNodeJSONArrayFlag(cmd, "node-cc", "node-cc-file", "节点指定抄送人")
	if err != nil {
		return nil, nil, err
	}
	return approverRaw, ccRaw, nil
}

// loadNodeJSONArrayFlag 读取一对 inline/file flag，校验其内容为 JSON 数组并返回原文；
// 两者都为空时返回 nil（不写入 body）。
func loadNodeJSONArrayFlag(cmd *cobra.Command, inlineFlag, fileFlag, label string) (json.RawMessage, error) {
	inline, _ := cmd.Flags().GetString(inlineFlag)
	file, _ := cmd.Flags().GetString(fileFlag)
	if inline == "" && file == "" {
		return nil, nil
	}
	s, err := loadJSONInput(inline, file, inlineFlag, fileFlag, label)
	if err != nil {
		return nil, err
	}
	var arr []any
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, fmt.Errorf("--%s 必须是 JSON 数组: %w", inlineFlag, err)
	}
	return json.RawMessage(s), nil
}
