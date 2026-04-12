package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const vcBase = "/open-apis/vc/v1"

// SearchMeetingsReq 会议搜索请求参数
// StartRFC3339/EndRFC3339 为空字符串表示不传；三个 ID 切片同理
type SearchMeetingsReq struct {
	Query          string
	StartRFC3339   string
	EndRFC3339     string
	OrganizerIDs   []string
	ParticipantIDs []string
	RoomIDs        []string
	PageSize       int
	PageToken      string
}

// SearchMeetings 搜索历史会议记录
// API: POST /open-apis/vc/v1/meetings/search
// 支持 query + 时间范围 + organizer_ids + participant_ids + open_room_ids 多维过滤
// 至少一个过滤条件由调用方保证
func SearchMeetings(req SearchMeetingsReq, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	// 构造请求体
	filter := map[string]any{}
	if req.StartRFC3339 != "" || req.EndRFC3339 != "" {
		startTime := map[string]string{}
		if req.StartRFC3339 != "" {
			startTime["start_time"] = req.StartRFC3339
		}
		if req.EndRFC3339 != "" {
			startTime["end_time"] = req.EndRFC3339
		}
		filter["start_time"] = startTime
	}
	if len(req.OrganizerIDs) > 0 {
		filter["organizer_ids"] = req.OrganizerIDs
	}
	if len(req.ParticipantIDs) > 0 {
		filter["participant_ids"] = req.ParticipantIDs
	}
	if len(req.RoomIDs) > 0 {
		filter["open_room_ids"] = req.RoomIDs
	}

	body := map[string]any{}
	if req.Query != "" {
		body["query"] = req.Query
	}
	if len(filter) > 0 {
		body["meeting_filter"] = filter
	}

	// 构造查询参数（分页）
	apiPath := fmt.Sprintf("%s/meetings/search", vcBase)
	params := url.Values{}
	if req.PageSize > 0 {
		params.Set("page_size", strconv.Itoa(req.PageSize))
	}
	if req.PageToken != "" {
		params.Set("page_token", req.PageToken)
	}
	if encoded := params.Encode(); encoded != "" {
		apiPath += "?" + encoded
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	resp, err := client.Post(Context(), apiPath, body, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("搜索会议失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索会议失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("搜索会议失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}
	return apiResp.Data, nil
}

// GetMeeting 获取会议详情
// API: GET /open-apis/vc/v1/meetings/{meeting_id}?with_participants=false&query_mode=0
// 返回 data 字段原始 JSON（含 meeting.note_id 等）
func GetMeeting(meetingID string, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)
	apiPath := fmt.Sprintf("%s/meetings/%s?with_participants=false&query_mode=0",
		vcBase, url.PathEscape(meetingID))

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取会议详情失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取会议详情失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("获取会议详情失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}
	return apiResp.Data, nil
}

// GetMeetingRecording 获取会议录制信息（含 minute 链接，可提取 minute_token）
// API: GET /open-apis/vc/v1/meetings/{meeting_id}/recording
func GetMeetingRecording(meetingID string, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)
	apiPath := fmt.Sprintf("%s/meetings/%s/recording", vcBase, url.PathEscape(meetingID))

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取会议录制失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取会议录制失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("获取会议录制失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}
	return apiResp.Data, nil
}

// GetMeetingNote 获取会议纪要文档引用
// API: GET /open-apis/vc/v1/notes/{note_id}
// 返回 data.note 原始 JSON（含 artifacts[].artifact_type/doc_token 和 references[].doc_token）
func GetMeetingNote(noteID string, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)
	apiPath := fmt.Sprintf("%s/notes/%s", vcBase, url.PathEscape(noteID))

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取会议纪要失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取会议纪要失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("获取会议纪要失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}
	return apiResp.Data, nil
}
