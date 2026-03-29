package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

// ==================== Bitable 数据结构 ====================

// BitableApp 多维表格基本信息
type BitableApp struct {
	AppToken string `json:"app_token"`
	Name     string `json:"name"`
	URL      string `json:"url,omitempty"`
	Revision int    `json:"revision,omitempty"`
}

// BitableTable 数据表信息
type BitableTable struct {
	TableID  string `json:"table_id"`
	Name     string `json:"name"`
	Revision int    `json:"revision,omitempty"`
}

// BitableField 字段信息
type BitableField struct {
	FieldID     string            `json:"field_id"`
	FieldName   string            `json:"field_name"`
	Type        int               `json:"type"`
	UIType      string            `json:"ui_type,omitempty"`
	IsPrimary   bool              `json:"is_primary,omitempty"`
	Description *BitableFieldDesc `json:"description,omitempty"`
	Property    json.RawMessage   `json:"property,omitempty"`
}

// BitableFieldDesc 字段描述，兼容字符串和对象两种 API 返回格式
type BitableFieldDesc struct {
	Text string
}

func (d *BitableFieldDesc) UnmarshalJSON(data []byte) error {
	// 尝试对象格式 {"text":"..."}
	var obj struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		d.Text = obj.Text
		return nil
	}
	// 回退到纯字符串格式
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.Text = s
		return nil
	}
	return nil
}

func (d BitableFieldDesc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Text string `json:"text"`
	}{Text: d.Text})
}

// BitableRecord 记录信息
type BitableRecord struct {
	RecordID string         `json:"record_id"`
	Fields   map[string]any `json:"fields"`
}

// BitableView 视图信息
type BitableView struct {
	ViewID   string `json:"view_id"`
	ViewName string `json:"view_name"`
	ViewType string `json:"view_type"`
}

// BitableSearchOptions 记录搜索选项
type BitableSearchOptions struct {
	PageSize   int               `json:"page_size,omitempty"`
	PageToken  string            `json:"page_token,omitempty"`
	Filter     *BitableFilter    `json:"filter,omitempty"`
	Sort       []BitableSortItem `json:"sort,omitempty"`
	FieldNames []string          `json:"field_names,omitempty"`
}

// BitableFilter 过滤条件
type BitableFilter struct {
	Conjunction string             `json:"conjunction"` // and / or
	Conditions  []BitableCondition `json:"conditions"`
}

// BitableCondition 过滤条件项
type BitableCondition struct {
	FieldName string `json:"field_name"`
	Operator  string `json:"operator"` // is, isNot, contains, doesNotContain, isEmpty, isNotEmpty, isGreater, isLess, etc.
	Value     []any  `json:"value,omitempty"`
}

// BitableSortItem 排序项
type BitableSortItem struct {
	FieldName string `json:"field_name"`
	Desc      bool   `json:"desc,omitempty"`
}

const bitableBase = "/open-apis/bitable/v1"

// ==================== 多维表格（App）操作 ====================

// CreateBitableApp 创建多维表格
func CreateBitableApp(name string, folderToken string, userAccessToken string) (*BitableApp, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"name": name,
	}
	if folderToken != "" {
		reqBody["folder_token"] = folderToken
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := bitableBase + "/apps"
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建多维表格失败: %w", err)
	}

	return parseBitableResponse[BitableApp](resp, "创建多维表格")
}

// GetBitableApp 获取多维表格元数据
func GetBitableApp(appToken string, userAccessToken string) (*BitableApp, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s", bitableBase, appToken)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取多维表格信息失败: %w", err)
	}

	return parseBitableResponse[BitableApp](resp, "获取多维表格信息")
}

// ==================== 数据表（Table）操作 ====================

// ListBitableTables 列出数据表
func ListBitableTables(appToken string, userAccessToken string) ([]BitableTable, error) {
	apiPath := fmt.Sprintf("%s/apps/%s/tables", bitableBase, appToken)
	return fetchAllBitableItems[BitableTable](apiPath, "列出数据表", userAccessToken)
}

// CreateBitableTable 创建数据表
func CreateBitableTable(appToken string, name string, userAccessToken string) (*BitableTable, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"table": map[string]any{
			"name": name,
		},
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables", bitableBase, appToken)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建数据表失败: %w", err)
	}

	return parseBitableResponse[BitableTable](resp, "创建数据表")
}

// DeleteBitableTable 删除数据表
func DeleteBitableTable(appToken string, tableID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s", bitableBase, appToken, tableID)
	resp, err := client.Delete(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("删除数据表失败: %w", err)
	}

	return checkBitableError(resp, "删除数据表")
}

// RenameBitableTable 重命名数据表
func RenameBitableTable(appToken string, tableID string, name string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	reqBody := map[string]any{
		"name": name,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s", bitableBase, appToken, tableID)
	resp, err := client.Patch(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("重命名数据表失败: %w", err)
	}

	return checkBitableError(resp, "重命名数据表")
}

// ==================== 字段（Field）操作 ====================

// ListBitableFields 列出字段
func ListBitableFields(appToken string, tableID string, userAccessToken string) ([]BitableField, error) {
	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/fields", bitableBase, appToken, tableID)
	return fetchAllBitableItems[BitableField](apiPath, "列出字段", userAccessToken)
}

// CreateBitableField 创建字段
func CreateBitableField(appToken string, tableID string, fieldDef map[string]any, userAccessToken string) (*BitableField, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/fields", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, fieldDef, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建字段失败: %w", err)
	}

	return parseBitableFieldResponse(resp, "创建字段")
}

// UpdateBitableField 更新字段（注意：单选字段必须带完整 property）
func UpdateBitableField(appToken string, tableID string, fieldID string, fieldDef map[string]any, userAccessToken string) (*BitableField, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/fields/%s", bitableBase, appToken, tableID, fieldID)
	resp, err := client.Put(Context(), apiPath, fieldDef, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("更新字段失败: %w", err)
	}

	return parseBitableFieldResponse(resp, "更新字段")
}

// DeleteBitableField 删除字段
func DeleteBitableField(appToken string, tableID string, fieldID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/fields/%s", bitableBase, appToken, tableID, fieldID)
	resp, err := client.Delete(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("删除字段失败: %w", err)
	}

	return checkBitableError(resp, "删除字段")
}

// ==================== 记录（Record）操作 ====================

// CreateBitableRecord 创建单条记录
func CreateBitableRecord(appToken string, tableID string, fields map[string]any, userAccessToken string) (*BitableRecord, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"fields": fields,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建记录失败: %w", err)
	}

	return parseBitableRecordResponse(resp, "创建记录")
}

// BatchCreateBitableRecords 批量创建记录（最多 500 条）
func BatchCreateBitableRecords(appToken string, tableID string, records []map[string]any, userAccessToken string) ([]BitableRecord, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	// 构造请求体
	items := make([]map[string]any, len(records))
	for i, r := range records {
		items[i] = map[string]any{"fields": r}
	}
	reqBody := map[string]any{
		"records": items,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records/batch_create", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("批量创建记录失败: %w", err)
	}

	return parseBitableRecordsResponse(resp, "批量创建记录")
}

// UpdateBitableRecord 更新单条记录
func UpdateBitableRecord(appToken string, tableID string, recordID string, fields map[string]any, userAccessToken string) (*BitableRecord, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"fields": fields,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records/%s", bitableBase, appToken, tableID, recordID)
	resp, err := client.Put(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("更新记录失败: %w", err)
	}

	return parseBitableRecordResponse(resp, "更新记录")
}

// BatchUpdateBitableRecords 批量更新记录（最多 500 条）
func BatchUpdateBitableRecords(appToken string, tableID string, records []map[string]any, userAccessToken string) ([]BitableRecord, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"records": records,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records/batch_update", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("批量更新记录失败: %w", err)
	}

	return parseBitableRecordsResponse(resp, "批量更新记录")
}

// BatchDeleteBitableRecords 批量删除记录（最多 500 条）
func BatchDeleteBitableRecords(appToken string, tableID string, recordIDs []string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	reqBody := map[string]any{
		"records": recordIDs,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records/batch_delete", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("批量删除记录失败: %w", err)
	}

	return checkBitableError(resp, "批量删除记录")
}

// SearchBitableRecords 搜索记录
func SearchBitableRecords(appToken string, tableID string, opts BitableSearchOptions, userAccessToken string) ([]BitableRecord, string, int, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", 0, err
	}

	reqBody := map[string]any{}
	if opts.PageSize > 0 {
		reqBody["page_size"] = opts.PageSize
	}
	if opts.PageToken != "" {
		reqBody["page_token"] = opts.PageToken
	}
	if opts.Filter != nil {
		reqBody["filter"] = opts.Filter
	}
	if len(opts.Sort) > 0 {
		reqBody["sort"] = opts.Sort
	}
	if len(opts.FieldNames) > 0 {
		reqBody["field_names"] = opts.FieldNames
	}

	tokenType, reqOpts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records/search", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, reqOpts...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("搜索记录失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", 0, fmt.Errorf("搜索记录失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items     []BitableRecord `json:"items"`
			PageToken string          `json:"page_token"`
			HasMore   bool            `json:"has_more"`
			Total     int             `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, "", 0, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, "", 0, fmt.Errorf("搜索记录失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}

	nextPageToken := ""
	if apiResp.Data.HasMore {
		nextPageToken = apiResp.Data.PageToken
	}

	return apiResp.Data.Items, nextPageToken, apiResp.Data.Total, nil
}

// GetBitableRecord 获取单条记录
func GetBitableRecord(appToken string, tableID string, recordID string, userAccessToken string) (*BitableRecord, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/records/%s", bitableBase, appToken, tableID, recordID)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取记录失败: %w", err)
	}

	return parseBitableRecordResponse(resp, "获取记录")
}

// ==================== 视图（View）操作 ====================

// ListBitableViews 列出视图
func ListBitableViews(appToken string, tableID string, pageSize int, pageToken string, userAccessToken string) ([]BitableView, string, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := buildBitablePagePath(
		fmt.Sprintf("%s/apps/%s/tables/%s/views", bitableBase, appToken, tableID),
		pageSize,
		pageToken,
	)

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("列出视图失败: %w", err)
	}

	return parseBitablePagedListResponse[BitableView](resp, "列出视图")
}

// CreateBitableView 创建视图
func CreateBitableView(appToken string, tableID string, viewName string, viewType string, userAccessToken string) (*BitableView, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"view_name": viewName,
		"view_type": viewType,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/views", bitableBase, appToken, tableID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建视图失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("创建视图失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			View BitableView `json:"view"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("创建视图失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}

	return &apiResp.Data.View, nil
}

// DeleteBitableView 删除视图
func DeleteBitableView(appToken string, tableID string, viewID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/views/%s", bitableBase, appToken, tableID, viewID)
	resp, err := client.Delete(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("删除视图失败: %w", err)
	}

	return checkBitableError(resp, "删除视图")
}

// ==================== 内部辅助函数 ====================

// checkBitableError 检查 Bitable API 响应错误
func checkBitableError(resp *larkcore.ApiResp, action string) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	return nil
}

// fetchAllBitableItems 自动翻页获取全部列表数据，避免 tables/fields 结果被截断。
func fetchAllBitableItems[T any](basePath string, action string, userAccessToken string) ([]T, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, reqOpts := resolveTokenOpts(userAccessToken)
	var allItems []T
	nextPageToken := ""

	for {
		apiPath := buildBitablePagePath(basePath, 0, nextPageToken)
		resp, err := client.Get(Context(), apiPath, nil, tokenType, reqOpts...)
		if err != nil {
			return nil, fmt.Errorf("%s失败: %w", action, err)
		}

		items, token, err := parseBitablePagedListResponse[T](resp, action)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, items...)
		if token == "" {
			return allItems, nil
		}

		nextPageToken = token
	}
}

func buildBitablePagePath(basePath string, pageSize int, pageToken string) string {
	params := url.Values{}
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		params.Set("page_token", pageToken)
	}
	if len(params) == 0 {
		return basePath
	}
	return basePath + "?" + params.Encode()
}

// parseBitableResponse 解析 Bitable API 响应（data 直接是对象）
func parseBitableResponse[T any](resp *larkcore.ApiResp, action string) (*T, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			App   *T `json:"app"`
			Table *T `json:"table"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	// 尝试不同的 data 字段名
	if apiResp.Data.App != nil {
		return apiResp.Data.App, nil
	}
	if apiResp.Data.Table != nil {
		return apiResp.Data.Table, nil
	}

	// 直接尝试从 data 解析
	var directResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data T      `json:"data"`
	}
	if err := json.Unmarshal(resp.RawBody, &directResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &directResp.Data, nil
}

// parseBitablePagedListResponse 解析带分页信息的 Bitable API 列表响应。
func parseBitablePagedListResponse[T any](resp *larkcore.ApiResp, action string) ([]T, string, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items     []T    `json:"items"`
			PageToken string `json:"page_token"`
			HasMore   bool   `json:"has_more"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, "", fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, "", fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	nextPageToken := ""
	if apiResp.Data.HasMore {
		nextPageToken = apiResp.Data.PageToken
	}

	return apiResp.Data.Items, nextPageToken, nil
}

// parseBitableFieldResponse 解析字段响应
func parseBitableFieldResponse(resp *larkcore.ApiResp, action string) (*BitableField, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Field BitableField `json:"field"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	return &apiResp.Data.Field, nil
}

// parseBitableRecordResponse 解析单条记录响应
func parseBitableRecordResponse(resp *larkcore.ApiResp, action string) (*BitableRecord, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Record BitableRecord `json:"record"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	return &apiResp.Data.Record, nil
}

// parseBitableRecordsResponse 解析多条记录响应
func parseBitableRecordsResponse(resp *larkcore.ApiResp, action string) ([]BitableRecord, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Records []BitableRecord `json:"records"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	return apiResp.Data.Records, nil
}

// ==================== 仪表盘（Dashboard）操作 ====================

// ListBitableDashboards 列出仪表盘
func ListBitableDashboards(appToken string, pageSize int, pageToken string, userAccessToken string) ([]map[string]any, string, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := buildBitablePagePath(
		fmt.Sprintf("%s/apps/%s/dashboards", bitableBase, appToken),
		pageSize,
		pageToken,
	)

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("列出仪表盘失败: %w", err)
	}

	return parseBitableRawPagedListResponse(resp, "列出仪表盘", "dashboards")
}

// GetBitableDashboard 获取仪表盘
// 注意：飞书 API 没有单独的获取仪表盘接口，通过 list 接口过滤实现
func GetBitableDashboard(appToken string, dashboardID string, userAccessToken string) (map[string]any, error) {
	// 通过 list 接口获取所有仪表盘，然后从中查找匹配的
	dashboards, _, err := ListBitableDashboards(appToken, 500, "", userAccessToken)
	if err != nil {
		return nil, fmt.Errorf("获取仪表盘失败: %w", err)
	}

	for _, d := range dashboards {
		if id, _ := d["block_id"].(string); id == dashboardID {
			return d, nil
		}
	}

	return nil, fmt.Errorf("获取仪表盘失败: 未找到 ID 为 %s 的仪表盘", dashboardID)
}

// CreateBitableDashboard 复制仪表盘创建新仪表盘
// 注意：飞书 API 不支持直接创建仪表盘，只支持复制现有仪表盘
// reqBody 需包含 "name"（新仪表盘名称）和 "source_block_id"（源仪表盘 block_id）
func CreateBitableDashboard(appToken string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	sourceBlockID, _ := reqBody["source_block_id"].(string)
	if sourceBlockID == "" {
		return nil, fmt.Errorf("飞书 API 不支持直接创建仪表盘，请通过 --source-block-id 指定要复制的源仪表盘 block_id")
	}

	// 只传 name 到 API
	copyBody := map[string]any{}
	if name, ok := reqBody["name"]; ok {
		copyBody["name"] = name
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	// 正确路径：POST /dashboards/:block_id/copy
	apiPath := fmt.Sprintf("%s/apps/%s/dashboards/%s/copy", bitableBase, appToken, sourceBlockID)
	resp, err := client.Post(Context(), apiPath, copyBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("复制仪表盘失败: %w", err)
	}

	return parseBitableRawResponse(resp, "复制仪表盘")
}

// UpdateBitableDashboard 更新仪表盘
// 注意：飞书 API 不支持更新仪表盘，此函数保留签名兼容但返回错误提示
func UpdateBitableDashboard(appToken string, dashboardID string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	return nil, fmt.Errorf("飞书 API 不支持更新仪表盘（仅支持 list 和 copy 操作）")
}

// DeleteBitableDashboard 删除仪表盘
// 注意：飞书 API 不支持删除仪表盘，此函数保留签名兼容但返回错误提示
func DeleteBitableDashboard(appToken string, dashboardID string, userAccessToken string) error {
	return fmt.Errorf("飞书 API 不支持删除仪表盘（仅支持 list 和 copy 操作）")
}

// ==================== 仪表盘 Block 操作（飞书 API 不支持） ====================
// 注意：飞书开放 API 不提供仪表盘 Block 的 CRUD 接口，以下函数保留签名兼容但返回错误提示

// ListBitableDashboardBlocks 列出仪表盘 Block（API 不支持）
func ListBitableDashboardBlocks(appToken string, dashboardID string, pageSize int, pageToken string, userAccessToken string) ([]map[string]any, string, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := buildBitablePagePath(
		fmt.Sprintf("%s/apps/%s/dashboards/%s/blocks", bitableBase, appToken, dashboardID),
		pageSize,
		pageToken,
	)

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("列出仪表盘 Block 失败: %w", err)
	}

	return parseBitableRawPagedListResponse(resp, "列出仪表盘 Block", "blocks")
}

// GetBitableDashboardBlock 获取仪表盘 Block
func GetBitableDashboardBlock(appToken string, dashboardID string, blockID string, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/dashboards/%s/blocks/%s", bitableBase, appToken, dashboardID, blockID)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取仪表盘 Block 失败: %w", err)
	}

	return parseBitableRawResponse(resp, "获取仪表盘 Block")
}

// CreateBitableDashboardBlock 创建仪表盘 Block
func CreateBitableDashboardBlock(appToken string, dashboardID string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/dashboards/%s/blocks", bitableBase, appToken, dashboardID)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建仪表盘 Block 失败: %w", err)
	}

	return parseBitableRawResponse(resp, "创建仪表盘 Block")
}

// UpdateBitableDashboardBlock 更新仪表盘 Block
func UpdateBitableDashboardBlock(appToken string, dashboardID string, blockID string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/dashboards/%s/blocks/%s", bitableBase, appToken, dashboardID, blockID)
	resp, err := client.Patch(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("更新仪表盘 Block 失败: %w", err)
	}

	return parseBitableRawResponse(resp, "更新仪表盘 Block")
}

// DeleteBitableDashboardBlock 删除仪表盘 Block
func DeleteBitableDashboardBlock(appToken string, dashboardID string, blockID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/dashboards/%s/blocks/%s", bitableBase, appToken, dashboardID, blockID)
	resp, err := client.Delete(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("删除仪表盘 Block 失败: %w", err)
	}

	return checkBitableError(resp, "删除仪表盘 Block")
}

// ==================== 视图高级配置 ====================

// GetBitableViewConfig 获取视图配置（filter_info/sort_info/group_info）
// 通过 GET /views/:view_id 获取完整视图信息，然后从 property 中提取对应配置
func GetBitableViewConfig(appToken string, tableID string, viewID string, configType string, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	// 正确路径：GET /views/:view_id，无独立的 filter_info/sort_info/group_info 子路径
	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/views/%s", bitableBase, appToken, tableID, viewID)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取视图配置失败: %w", err)
	}

	data, err := parseBitableRawResponse(resp, "获取视图配置")
	if err != nil {
		return nil, err
	}

	// 从 view.property 中提取对应的配置
	view, _ := data["view"].(map[string]any)
	if view == nil {
		return data, nil
	}
	property, _ := view["property"].(map[string]any)
	if property == nil {
		return map[string]any{}, nil
	}
	if configData, ok := property[configType]; ok {
		if m, ok := configData.(map[string]any); ok {
			return m, nil
		}
	}

	return map[string]any{}, nil
}

// SetBitableViewConfig 设置视图配置（filter_info/sort_info/group_info）
// 通过 PATCH /views/:view_id 更新视图属性
func SetBitableViewConfig(appToken string, tableID string, viewID string, configType string, reqBody any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	// 正确路径：PATCH /views/:view_id，将配置包装在 property 中
	patchBody := map[string]any{
		"property": map[string]any{
			configType: reqBody,
		},
	}

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/views/%s", bitableBase, appToken, tableID, viewID)
	resp, err := client.Patch(Context(), apiPath, patchBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("设置视图配置失败: %w", err)
	}

	return parseBitableRawResponse(resp, "设置视图配置")
}

// ==================== 工作流（Workflow）操作 ====================

// ListBitableWorkflows 列出工作流
func ListBitableWorkflows(appToken string, pageSize int, pageToken string, userAccessToken string) ([]map[string]any, string, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := buildBitablePagePath(
		fmt.Sprintf("%s/apps/%s/workflows", bitableBase, appToken),
		pageSize,
		pageToken,
	)

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("列出工作流失败: %w", err)
	}

	return parseBitableRawPagedListResponse(resp, "列出工作流", "workflows")
}

// GetBitableWorkflow 获取工作流
func GetBitableWorkflow(appToken string, workflowID string, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/workflows/%s", bitableBase, appToken, workflowID)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取工作流失败: %w", err)
	}

	return parseBitableRawResponse(resp, "获取工作流")
}

// EnableBitableWorkflow 启用工作流
// 正确 API：PUT /workflows/:workflow_id，请求体 {"status": "Enable"}
func EnableBitableWorkflow(appToken string, workflowID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	reqBody := map[string]any{
		"status": "Enable",
	}

	apiPath := fmt.Sprintf("%s/apps/%s/workflows/%s", bitableBase, appToken, workflowID)
	resp, err := client.Put(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("启用工作流失败: %w", err)
	}

	return checkBitableError(resp, "启用工作流")
}

// DisableBitableWorkflow 禁用工作流
// 正确 API：PUT /workflows/:workflow_id，请求体 {"status": "Disable"}
func DisableBitableWorkflow(appToken string, workflowID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	reqBody := map[string]any{
		"status": "Disable",
	}

	apiPath := fmt.Sprintf("%s/apps/%s/workflows/%s", bitableBase, appToken, workflowID)
	resp, err := client.Put(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("禁用工作流失败: %w", err)
	}

	return checkBitableError(resp, "禁用工作流")
}

// ==================== 表单（Form）操作 ====================

// ListBitableForms 列出表单
// 注意：飞书 API 没有独立的 list forms 接口，通过 list views 过滤 form 类型实现
func ListBitableForms(appToken string, tableID string, pageSize int, pageToken string, userAccessToken string) ([]map[string]any, string, error) {
	// 通过列出视图接口获取所有视图，然后过滤出 form 类型的
	views, nextToken, err := ListBitableViews(appToken, tableID, 100, pageToken, userAccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("列出表单失败: %w", err)
	}

	var forms []map[string]any
	for _, v := range views {
		if v.ViewType == "form" {
			forms = append(forms, map[string]any{
				"form_id":   v.ViewID,
				"name":      v.ViewName,
				"view_type": v.ViewType,
			})
		}
	}

	return forms, nextToken, nil
}

// GetBitableForm 获取表单
func GetBitableForm(appToken string, tableID string, formID string, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/forms/%s", bitableBase, appToken, tableID, formID)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取表单失败: %w", err)
	}

	return parseBitableRawResponse(resp, "获取表单")
}

// PatchBitableForm 更新表单
func PatchBitableForm(appToken string, tableID string, formID string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/tables/%s/forms/%s", bitableBase, appToken, tableID, formID)
	resp, err := client.Patch(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("更新表单失败: %w", err)
	}

	return parseBitableRawResponse(resp, "更新表单")
}

// ==================== 角色（Role）操作 ====================

// ListBitableRoles 列出角色（使用 base/v2 新版 API）
func ListBitableRoles(appToken string, pageSize int, pageToken string, userAccessToken string) ([]map[string]any, string, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	// 新版 API 路径：/open-apis/base/v2/apps/:app_token/roles
	apiPath := buildBitablePagePath(
		fmt.Sprintf("/open-apis/base/v2/apps/%s/roles", appToken),
		pageSize,
		pageToken,
	)

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("列出角色失败: %w", err)
	}

	return parseBitableRawPagedListResponse(resp, "列出角色", "items")
}

// CreateBitableRole 创建角色（使用 base/v2 新版 API）
func CreateBitableRole(appToken string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	// 新版 API 路径：/open-apis/base/v2/apps/:app_token/roles
	apiPath := fmt.Sprintf("/open-apis/base/v2/apps/%s/roles", appToken)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	return parseBitableRawResponse(resp, "创建角色")
}

// DeleteBitableRole 删除角色
func DeleteBitableRole(appToken string, roleID string, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/roles/%s", bitableBase, appToken, roleID)
	resp, err := client.Delete(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	return checkBitableError(resp, "删除角色")
}

// ==================== 复制（Copy）操作 ====================

// CopyBitableApp 复制多维表格
func CopyBitableApp(appToken string, reqBody map[string]any, userAccessToken string) (map[string]any, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/apps/%s/copy", bitableBase, appToken)
	resp, err := client.Post(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("复制多维表格失败: %w", err)
	}

	return parseBitableRawResponse(resp, "复制多维表格")
}

// ==================== 高级权限（Advanced Permissions）操作 ====================

// UpdateBitableAdvancedPermissions 启用或禁用多维表格高级权限
// 正确 API：PUT /open-apis/bitable/v1/apps/:app_token，请求体 {"is_advanced": true/false}
func UpdateBitableAdvancedPermissions(appToken string, enabled bool, userAccessToken string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	reqBody := map[string]any{
		"is_advanced": enabled,
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	// 正确路径：PUT /apps/:app_token（更新多维表格元数据）
	apiPath := fmt.Sprintf("%s/apps/%s", bitableBase, appToken)
	resp, err := client.Put(Context(), apiPath, reqBody, tokenType, opts...)
	if err != nil {
		action := "启用"
		if !enabled {
			action = "禁用"
		}
		return fmt.Errorf("%s高级权限失败: %w", action, err)
	}

	action := "启用"
	if !enabled {
		action = "禁用"
	}
	return checkBitableError(resp, action+"高级权限")
}

// ==================== 通用 Raw 响应解析 ====================

// parseBitableRawResponse 解析 Bitable API 响应，返回 data 字段的原始 map
func parseBitableRawResponse(resp *larkcore.ApiResp, action string) (map[string]any, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data map[string]any `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	return apiResp.Data, nil
}

// parseBitableRawPagedListResponse 解析带分页的 Bitable API 列表响应，返回原始 map 切片
func parseBitableRawPagedListResponse(resp *larkcore.ApiResp, action string, itemsKey string) ([]map[string]any, string, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("%s失败: HTTP %d, body: %s", action, resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items     json.RawMessage `json:"items"`
			PageToken string          `json:"page_token"`
			HasMore   bool            `json:"has_more"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, "", fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, "", fmt.Errorf("%s失败: code=%d, msg=%s", action, apiResp.Code, apiResp.Msg)
	}

	// 优先尝试从指定 key 获取列表
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(resp.RawBody, &struct {
		Data *map[string]json.RawMessage `json:"data"`
	}{Data: &rawData}); err == nil {
		if raw, ok := rawData[itemsKey]; ok {
			var items []map[string]any
			if err := json.Unmarshal(raw, &items); err == nil {
				nextPageToken := ""
				if apiResp.Data.HasMore {
					nextPageToken = apiResp.Data.PageToken
				}
				return items, nextPageToken, nil
			}
		}
	}

	// 回退到 items 字段
	var items []map[string]any
	if apiResp.Data.Items != nil {
		if err := json.Unmarshal(apiResp.Data.Items, &items); err != nil {
			return nil, "", fmt.Errorf("解析列表数据失败: %w", err)
		}
	}

	nextPageToken := ""
	if apiResp.Data.HasMore {
		nextPageToken = apiResp.Data.PageToken
	}

	return items, nextPageToken, nil
}
