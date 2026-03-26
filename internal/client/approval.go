package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
)

// GetApprovalOptions represents optional filters for fetching an approval definition.
type GetApprovalOptions struct {
	Locale      string
	WithAdminID bool
	UserIDType  string
	WithOption  bool
	UserID      string
}

// ApprovalDefinition represents a simplified approval definition.
type ApprovalDefinition struct {
	ApprovalCode       string            `json:"approval_code"`
	ApprovalName       string            `json:"approval_name"`
	Status             string            `json:"status"`
	Form               any               `json:"form,omitempty"`
	NodeList           []*ApprovalNode   `json:"node_list,omitempty"`
	Viewers            []*ApprovalViewer `json:"viewers,omitempty"`
	ApprovalAdminIDs   []string          `json:"approval_admin_ids,omitempty"`
	FormWidgetRelation any               `json:"form_widget_relation,omitempty"`
}

// ApprovalNode represents a simplified approval node.
type ApprovalNode struct {
	Name                string `json:"name,omitempty"`
	NodeID              string `json:"node_id,omitempty"`
	CustomNodeID        string `json:"custom_node_id,omitempty"`
	NodeType            string `json:"node_type,omitempty"`
	NeedApprover        bool   `json:"need_approver,omitempty"`
	ApproverChosenMulti bool   `json:"approver_chosen_multi,omitempty"`
	RequireSignature    bool   `json:"require_signature,omitempty"`
}

// ApprovalViewer represents a simplified approval viewer.
type ApprovalViewer struct {
	Type   string `json:"type,omitempty"`
	ID     string `json:"id,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// ApprovalTaskQueryOptions represents options for querying approval tasks.
type ApprovalTaskQueryOptions struct {
	PageSize   int
	PageToken  string
	UserID     string
	Topic      string
	UserIDType string
}

// ApprovalTaskQueryResult represents a simplified approval task query result.
type ApprovalTaskQueryResult struct {
	Tasks     []*ApprovalTaskInfo `json:"tasks"`
	PageToken string              `json:"page_token,omitempty"`
	HasMore   bool                `json:"has_more"`
	Count     *ApprovalTaskCount  `json:"count,omitempty"`
}

// ApprovalTaskCount represents summary count information returned on the first page.
type ApprovalTaskCount struct {
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// ApprovalTaskInfo represents a simplified approval task.
type ApprovalTaskInfo struct {
	Topic               string   `json:"topic,omitempty"`
	UserID              string   `json:"user_id,omitempty"`
	Title               string   `json:"title,omitempty"`
	HelpdeskURL         string   `json:"helpdesk_url,omitempty"`
	MobileURL           string   `json:"mobile_url,omitempty"`
	PCURL               string   `json:"pc_url,omitempty"`
	ProcessExternalID   string   `json:"process_external_id,omitempty"`
	TaskExternalID      string   `json:"task_external_id,omitempty"`
	Status              string   `json:"status,omitempty"`
	ProcessStatus       string   `json:"process_status,omitempty"`
	DefinitionCode      string   `json:"definition_code,omitempty"`
	DefinitionName      string   `json:"definition_name,omitempty"`
	DefinitionID        string   `json:"definition_id,omitempty"`
	DefinitionGroupID   string   `json:"definition_group_id,omitempty"`
	DefinitionGroupName string   `json:"definition_group_name,omitempty"`
	Initiators          []string `json:"initiators,omitempty"`
	InitiatorNames      []string `json:"initiator_names,omitempty"`
	TaskID              string   `json:"task_id,omitempty"`
	ProcessID           string   `json:"process_id,omitempty"`
	ProcessCode         string   `json:"process_code,omitempty"`
}

type approvalTaskString string

func (s *approvalTaskString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*s = ""
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = approvalTaskString(str)
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var number json.Number
	if err := decoder.Decode(&number); err == nil {
		*s = approvalTaskString(number.String())
		return nil
	}

	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		*s = approvalTaskString(strconv.FormatBool(boolean))
		return nil
	}

	return fmt.Errorf("不支持的审批任务字段类型: %s", string(data))
}

func (s approvalTaskString) String() string {
	return string(s)
}

type approvalTaskQueryAPIResp struct {
	Code int                       `json:"code"`
	Msg  string                    `json:"msg"`
	Data *approvalTaskQueryAPIData `json:"data"`
}

type approvalTaskQueryAPIData struct {
	Tasks     []*approvalTaskAPIInfo `json:"tasks,omitempty"`
	PageToken string                 `json:"page_token,omitempty"`
	HasMore   bool                   `json:"has_more,omitempty"`
	Count     *ApprovalTaskCount     `json:"count,omitempty"`
}

type approvalTaskAPIInfo struct {
	Topic               approvalTaskString `json:"topic,omitempty"`
	UserID              approvalTaskString `json:"user_id,omitempty"`
	Title               approvalTaskString `json:"title,omitempty"`
	Urls                *approvalTaskURLs  `json:"urls,omitempty"`
	ProcessExternalID   approvalTaskString `json:"process_external_id,omitempty"`
	TaskExternalID      approvalTaskString `json:"task_external_id,omitempty"`
	Status              approvalTaskString `json:"status,omitempty"`
	ProcessStatus       approvalTaskString `json:"process_status,omitempty"`
	DefinitionCode      approvalTaskString `json:"definition_code,omitempty"`
	DefinitionName      approvalTaskString `json:"definition_name,omitempty"`
	DefinitionID        approvalTaskString `json:"definition_id,omitempty"`
	DefinitionGroupID   approvalTaskString `json:"definition_group_id,omitempty"`
	DefinitionGroupName approvalTaskString `json:"definition_group_name,omitempty"`
	Initiators          []string           `json:"initiators,omitempty"`
	InitiatorNames      []string           `json:"initiator_names,omitempty"`
	TaskID              approvalTaskString `json:"task_id,omitempty"`
	ProcessID           approvalTaskString `json:"process_id,omitempty"`
	ProcessCode         approvalTaskString `json:"process_code,omitempty"`
}

type approvalTaskURLs struct {
	Helpdesk approvalTaskString `json:"helpdesk,omitempty"`
	Mobile   approvalTaskString `json:"mobile,omitempty"`
	Pc       approvalTaskString `json:"pc,omitempty"`
}

// GetApprovalDefinition retrieves approval definition details by approval code.
func GetApprovalDefinition(approvalCode string, opts GetApprovalOptions) (*ApprovalDefinition, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBuilder := larkapproval.NewGetApprovalReqBuilder().
		ApprovalCode(approvalCode)

	if opts.Locale != "" {
		reqBuilder.Locale(opts.Locale)
	}
	if opts.WithAdminID {
		reqBuilder.WithAdminId(true)
	}
	if opts.UserIDType != "" {
		reqBuilder.UserIdType(opts.UserIDType)
	}
	if opts.WithOption {
		reqBuilder.WithOption(true)
	}
	if opts.UserID != "" {
		reqBuilder.UserId(opts.UserID)
	}

	resp, err := client.Approval.V4.Approval.Get(Context(), reqBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("获取审批定义失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取审批定义失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("获取审批定义返回数据为空")
	}

	result := &ApprovalDefinition{
		ApprovalCode:       approvalCode,
		ApprovalName:       StringVal(resp.Data.ApprovalName),
		Status:             StringVal(resp.Data.Status),
		Form:               parseEmbeddedJSON(StringVal(resp.Data.Form)),
		ApprovalAdminIDs:   resp.Data.ApprovalAdminIds,
		FormWidgetRelation: parseEmbeddedJSON(StringVal(resp.Data.FormWidgetRelation)),
	}

	if len(resp.Data.NodeList) > 0 {
		result.NodeList = make([]*ApprovalNode, 0, len(resp.Data.NodeList))
		for _, node := range resp.Data.NodeList {
			if node == nil {
				continue
			}
			result.NodeList = append(result.NodeList, &ApprovalNode{
				Name:                StringVal(node.Name),
				NodeID:              StringVal(node.NodeId),
				CustomNodeID:        StringVal(node.CustomNodeId),
				NodeType:            StringVal(node.NodeType),
				NeedApprover:        BoolVal(node.NeedApprover),
				ApproverChosenMulti: BoolVal(node.ApproverChosenMulti),
				RequireSignature:    BoolVal(node.RequireSignature),
			})
		}
	}

	if len(resp.Data.Viewers) > 0 {
		result.Viewers = make([]*ApprovalViewer, 0, len(resp.Data.Viewers))
		for _, viewer := range resp.Data.Viewers {
			if viewer == nil {
				continue
			}
			result.Viewers = append(result.Viewers, &ApprovalViewer{
				Type:   StringVal(viewer.Type),
				ID:     StringVal(viewer.Id),
				UserID: StringVal(viewer.UserId),
			})
		}
	}

	return result, nil
}

func queryApprovalTasksRawBody(opts ApprovalTaskQueryOptions, userAccessToken string) ([]byte, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	req := &larkcore.ApiReq{
		HttpMethod:                http.MethodGet,
		ApiPath:                   "/open-apis/approval/v4/tasks/query",
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.QueryParams.Set("user_id", opts.UserID)
	req.QueryParams.Set("topic", opts.Topic)

	if opts.PageSize > 0 {
		req.QueryParams.Set("page_size", strconv.Itoa(opts.PageSize))
	}
	if opts.PageToken != "" {
		req.QueryParams.Set("page_token", opts.PageToken)
	}
	if opts.UserIDType != "" {
		req.QueryParams.Set("user_id_type", opts.UserIDType)
	}

	resp, err := client.Do(Context(), req, UserTokenOption(userAccessToken)...)
	if err != nil {
		return nil, fmt.Errorf("查询审批任务失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询审批任务失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	return resp.RawBody, nil
}

// QueryApprovalTasksRaw retrieves the raw approval task response body from the API.
func QueryApprovalTasksRaw(opts ApprovalTaskQueryOptions, userAccessToken string) ([]byte, error) {
	return queryApprovalTasksRawBody(opts, userAccessToken)
}

// QueryApprovalTasks retrieves approval tasks for a user.
func QueryApprovalTasks(opts ApprovalTaskQueryOptions, userAccessToken string) (*ApprovalTaskQueryResult, error) {
	body, err := queryApprovalTasksRawBody(opts, userAccessToken)
	if err != nil {
		return nil, err
	}

	return parseApprovalTaskQueryResponse(body)
}

func parseEmbeddedJSON(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var value any
	if err := json.Unmarshal([]byte(raw), &value); err == nil {
		return value
	}

	return raw
}

func approvalTaskToInfo(task *larkapproval.Task) *ApprovalTaskInfo {
	info := &ApprovalTaskInfo{
		Topic:               StringVal(task.Topic),
		UserID:              StringVal(task.UserId),
		Title:               StringVal(task.Title),
		ProcessExternalID:   StringVal(task.ProcessExternalId),
		TaskExternalID:      StringVal(task.TaskExternalId),
		Status:              StringVal(task.Status),
		ProcessStatus:       StringVal(task.ProcessStatus),
		DefinitionCode:      StringVal(task.DefinitionCode),
		DefinitionName:      StringVal(task.DefinitionName),
		DefinitionID:        StringVal(task.DefinitionId),
		DefinitionGroupID:   StringVal(task.DefinitionGroupId),
		DefinitionGroupName: StringVal(task.DefinitionGroupName),
		Initiators:          task.Initiators,
		InitiatorNames:      task.InitiatorNames,
		TaskID:              StringVal(task.TaskId),
		ProcessID:           StringVal(task.ProcessId),
		ProcessCode:         StringVal(task.ProcessCode),
	}

	if task.Urls != nil {
		info.HelpdeskURL = StringVal(task.Urls.Helpdesk)
		info.MobileURL = StringVal(task.Urls.Mobile)
		info.PCURL = StringVal(task.Urls.Pc)
	}

	return info
}

func parseApprovalTaskQueryResponse(body []byte) (*ApprovalTaskQueryResult, error) {
	var apiResp approvalTaskQueryAPIResp
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析审批任务响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("查询审批任务失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}

	result := &ApprovalTaskQueryResult{
		Tasks: make([]*ApprovalTaskInfo, 0),
	}

	if apiResp.Data == nil {
		return result, nil
	}

	result.PageToken = apiResp.Data.PageToken
	result.HasMore = apiResp.Data.HasMore
	result.Count = apiResp.Data.Count

	if len(apiResp.Data.Tasks) > 0 {
		result.Tasks = make([]*ApprovalTaskInfo, 0, len(apiResp.Data.Tasks))
		for _, task := range apiResp.Data.Tasks {
			if task == nil {
				continue
			}
			result.Tasks = append(result.Tasks, approvalTaskAPIToInfo(task))
		}
	}

	return result, nil
}

func approvalTaskAPIToInfo(task *approvalTaskAPIInfo) *ApprovalTaskInfo {
	info := &ApprovalTaskInfo{
		Topic:               task.Topic.String(),
		UserID:              task.UserID.String(),
		Title:               task.Title.String(),
		ProcessExternalID:   task.ProcessExternalID.String(),
		TaskExternalID:      task.TaskExternalID.String(),
		Status:              task.Status.String(),
		ProcessStatus:       task.ProcessStatus.String(),
		DefinitionCode:      task.DefinitionCode.String(),
		DefinitionName:      task.DefinitionName.String(),
		DefinitionID:        task.DefinitionID.String(),
		DefinitionGroupID:   task.DefinitionGroupID.String(),
		DefinitionGroupName: task.DefinitionGroupName.String(),
		Initiators:          task.Initiators,
		InitiatorNames:      task.InitiatorNames,
		TaskID:              task.TaskID.String(),
		ProcessID:           task.ProcessID.String(),
		ProcessCode:         task.ProcessCode.String(),
	}

	if task.Urls != nil {
		info.HelpdeskURL = task.Urls.Helpdesk.String()
		info.MobileURL = task.Urls.Mobile.String()
		info.PCURL = task.Urls.Pc.String()
	}

	return info
}
