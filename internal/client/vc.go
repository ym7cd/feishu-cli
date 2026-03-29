package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// ==================== VC 数据结构 ====================

// VCMeeting 会议信息
type VCMeeting struct {
	MeetingID   string `json:"meeting_id"`
	MeetingNo   string `json:"meeting_no"`
	Topic       string `json:"topic"`
	HostUser    any    `json:"host_user,omitempty"`
	Status      int    `json:"status,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	ParticipantCount string `json:"participant_count,omitempty"`
	URL         string `json:"url,omitempty"`
}

// VCMeetingDetail 会议详情
type VCMeetingDetail struct {
	Meeting      json.RawMessage `json:"meeting,omitempty"`
	MeetingRecord json.RawMessage `json:"meeting_record,omitempty"`
}

const vcBase = "/open-apis/vc/v1"

// ==================== 会议列表 ====================

// SearchMeetings 搜索会议列表
// 使用 GET /open-apis/vc/v1/meeting_list 获取会议列表
func SearchMeetings(startTime, endTime int64, meetingStatus int, meetingNo string, pageSize int, pageToken string, userAccessToken string) (json.RawMessage, string, bool, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", false, err
	}

	apiPath := fmt.Sprintf("%s/meeting_list?start_time=%s&end_time=%s",
		vcBase,
		strconv.FormatInt(startTime, 10),
		strconv.FormatInt(endTime, 10),
	)

	if meetingStatus > 0 {
		apiPath += fmt.Sprintf("&meeting_status=%d", meetingStatus)
	}
	if meetingNo != "" {
		apiPath += "&meeting_no=" + meetingNo
	}
	if pageSize > 0 {
		apiPath += fmt.Sprintf("&page_size=%d", pageSize)
	}
	if pageToken != "" {
		apiPath += "&page_token=" + pageToken
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, "", false, fmt.Errorf("搜索会议列表失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("搜索会议列表失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MeetingList json.RawMessage `json:"meeting_list"`
			PageToken   string          `json:"page_token"`
			HasMore     bool            `json:"has_more"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.RawBody, &apiResp); err != nil {
		return nil, "", false, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, "", false, fmt.Errorf("搜索会议列表失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}

	return apiResp.Data.MeetingList, apiResp.Data.PageToken, apiResp.Data.HasMore, nil
}

// ==================== 会议详情 ====================

// GetMeeting 获取会议详情
// 使用 GET /open-apis/vc/v1/meetings/:meeting_id
func GetMeeting(meetingID string, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("%s/meetings/%s", vcBase, meetingID)
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

// ==================== 妙记 ====================

// GetMinute 获取妙记信息
// 使用 GET /open-apis/minutes/v1/minutes/:minute_token
func GetMinute(minuteToken string, userAccessToken string) (json.RawMessage, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	tokenType, opts := resolveTokenOpts(userAccessToken)

	apiPath := fmt.Sprintf("/open-apis/minutes/v1/minutes/%s", minuteToken)
	resp, err := client.Get(Context(), apiPath, nil, tokenType, opts...)
	if err != nil {
		return nil, fmt.Errorf("获取妙记信息失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取妙记信息失败: HTTP %d, body: %s", resp.StatusCode, string(resp.RawBody))
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
		return nil, fmt.Errorf("获取妙记信息失败: code=%d, msg=%s", apiResp.Code, apiResp.Msg)
	}

	return apiResp.Data, nil
}
