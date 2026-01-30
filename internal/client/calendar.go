package client

import (
	"fmt"
	"strconv"
	"time"

	larkcalendar "github.com/larksuite/oapi-sdk-go/v3/service/calendar/v4"
)

// Calendar 日历信息
type Calendar struct {
	CalendarID   string `json:"calendar_id"`
	Summary      string `json:"summary"`
	Description  string `json:"description,omitempty"`
	Permissions  string `json:"permissions,omitempty"`
	Type         string `json:"type,omitempty"`
	Color        int    `json:"color,omitempty"`
	Role         string `json:"role,omitempty"`
	SummaryAlias string `json:"summary_alias,omitempty"`
	IsDeleted    bool   `json:"is_deleted,omitempty"`
	IsThirdParty bool   `json:"is_third_party,omitempty"`
}

// CalendarEvent 日程信息
type CalendarEvent struct {
	EventID     string `json:"event_id"`
	OrganizerID string `json:"organizer_calendar_id,omitempty"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	TimeZone    string `json:"time_zone,omitempty"`
	Location    string `json:"location,omitempty"`
	Status      string `json:"status,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	CreateTime  string `json:"create_time,omitempty"`
	RecurringID string `json:"recurring_event_id,omitempty"`
	IsException bool   `json:"is_exception,omitempty"`
	AppLink     string `json:"app_link,omitempty"`
	Color       int    `json:"color,omitempty"`
}

// ListCalendars 列出日历
func ListCalendars(pageSize int, pageToken string) ([]*Calendar, string, bool, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", false, err
	}

	reqBuilder := larkcalendar.NewListCalendarReqBuilder()
	if pageSize > 0 {
		reqBuilder.PageSize(pageSize)
	}
	if pageToken != "" {
		reqBuilder.PageToken(pageToken)
	}

	resp, err := client.Calendar.Calendar.List(Context(), reqBuilder.Build())
	if err != nil {
		return nil, "", false, fmt.Errorf("获取日历列表失败: %w", err)
	}

	if !resp.Success() {
		return nil, "", false, fmt.Errorf("获取日历列表失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var calendars []*Calendar
	if resp.Data != nil && resp.Data.CalendarList != nil {
		for _, item := range resp.Data.CalendarList {
			calendars = append(calendars, &Calendar{
				CalendarID:   StringVal(item.CalendarId),
				Summary:      StringVal(item.Summary),
				Description:  StringVal(item.Description),
				Permissions:  StringVal(item.Permissions),
				Type:         StringVal(item.Type),
				Color:        IntVal(item.Color),
				Role:         StringVal(item.Role),
				SummaryAlias: StringVal(item.SummaryAlias),
				IsDeleted:    BoolVal(item.IsDeleted),
				IsThirdParty: BoolVal(item.IsThirdParty),
			})
		}
	}

	var nextPageToken string
	var hasMore bool
	if resp.Data != nil {
		nextPageToken = StringVal(resp.Data.PageToken)
		hasMore = BoolVal(resp.Data.HasMore)
	}

	return calendars, nextPageToken, hasMore, nil
}

// CreateEventParams 创建日程的参数
type CreateEventParams struct {
	CalendarID  string
	Summary     string
	Description string
	StartTime   string // RFC3339 格式
	EndTime     string // RFC3339 格式
	TimeZone    string
	Location    string
}

// CreateEvent 创建日程
func CreateEvent(params *CreateEventParams) (*CalendarEvent, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	startTs, err := parseTimeToTimestamp(params.StartTime)
	if err != nil {
		return nil, fmt.Errorf("解析开始时间失败: %w", err)
	}
	endTs, err := parseTimeToTimestamp(params.EndTime)
	if err != nil {
		return nil, fmt.Errorf("解析结束时间失败: %w", err)
	}

	startTime := larkcalendar.NewTimeInfoBuilder().
		Timestamp(startTs).
		Build()
	endTime := larkcalendar.NewTimeInfoBuilder().
		Timestamp(endTs).
		Build()

	eventBuilder := larkcalendar.NewCalendarEventBuilder().
		Summary(params.Summary).
		StartTime(startTime).
		EndTime(endTime)

	if params.Description != "" {
		eventBuilder.Description(params.Description)
	}

	if params.Location != "" {
		location := larkcalendar.NewEventLocationBuilder().
			Name(params.Location).
			Build()
		eventBuilder.Location(location)
	}

	req := larkcalendar.NewCreateCalendarEventReqBuilder().
		CalendarId(params.CalendarID).
		CalendarEvent(eventBuilder.Build()).
		Build()

	resp, err := client.Calendar.CalendarEvent.Create(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("创建日程失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("创建日程失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data == nil || resp.Data.Event == nil {
		return nil, fmt.Errorf("创建日程成功但未返回日程信息")
	}

	return convertEvent(resp.Data.Event), nil
}

// GetEvent 获取日程详情
func GetEvent(calendarID, eventID string) (*CalendarEvent, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	req := larkcalendar.NewGetCalendarEventReqBuilder().
		CalendarId(calendarID).
		EventId(eventID).
		Build()

	resp, err := client.Calendar.CalendarEvent.Get(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("获取日程详情失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取日程详情失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data == nil || resp.Data.Event == nil {
		return nil, fmt.Errorf("日程不存在")
	}

	return convertEvent(resp.Data.Event), nil
}

// ListEventsParams 列出日程的参数
type ListEventsParams struct {
	CalendarID string
	StartTime  string // RFC3339 格式，可选
	EndTime    string // RFC3339 格式，可选
	PageSize   int
	PageToken  string
}

// ListEvents 列出日程
func ListEvents(params *ListEventsParams) ([]*CalendarEvent, string, bool, error) {
	client, err := GetClient()
	if err != nil {
		return nil, "", false, err
	}

	reqBuilder := larkcalendar.NewListCalendarEventReqBuilder().
		CalendarId(params.CalendarID)

	if params.StartTime != "" {
		startTs, err := parseTimeToTimestamp(params.StartTime)
		if err != nil {
			return nil, "", false, fmt.Errorf("解析开始时间失败: %w", err)
		}
		reqBuilder.StartTime(startTs)
	}

	if params.EndTime != "" {
		endTs, err := parseTimeToTimestamp(params.EndTime)
		if err != nil {
			return nil, "", false, fmt.Errorf("解析结束时间失败: %w", err)
		}
		reqBuilder.EndTime(endTs)
	}

	if params.PageSize > 0 {
		reqBuilder.PageSize(params.PageSize)
	}

	if params.PageToken != "" {
		reqBuilder.PageToken(params.PageToken)
	}

	resp, err := client.Calendar.CalendarEvent.List(Context(), reqBuilder.Build())
	if err != nil {
		return nil, "", false, fmt.Errorf("获取日程列表失败: %w", err)
	}

	if !resp.Success() {
		return nil, "", false, fmt.Errorf("获取日程列表失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	var events []*CalendarEvent
	if resp.Data != nil && resp.Data.Items != nil {
		for _, item := range resp.Data.Items {
			events = append(events, convertEvent(item))
		}
	}

	var nextPageToken string
	var hasMore bool
	if resp.Data != nil {
		nextPageToken = StringVal(resp.Data.PageToken)
		hasMore = BoolVal(resp.Data.HasMore)
	}

	return events, nextPageToken, hasMore, nil
}

// UpdateEventParams 更新日程的参数
type UpdateEventParams struct {
	CalendarID  string
	EventID     string
	Summary     string
	Description string
	StartTime   string // RFC3339 格式
	EndTime     string // RFC3339 格式
	Location    string
}

// UpdateEvent 更新日程（使用 Patch 方式）
func UpdateEvent(params *UpdateEventParams) (*CalendarEvent, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	eventBuilder := larkcalendar.NewCalendarEventBuilder()

	if params.Summary != "" {
		eventBuilder.Summary(params.Summary)
	}

	if params.Description != "" {
		eventBuilder.Description(params.Description)
	}

	if params.StartTime != "" {
		startTs, err := parseTimeToTimestamp(params.StartTime)
		if err != nil {
			return nil, fmt.Errorf("解析开始时间失败: %w", err)
		}
		startTime := larkcalendar.NewTimeInfoBuilder().
			Timestamp(startTs).
			Build()
		eventBuilder.StartTime(startTime)
	}

	if params.EndTime != "" {
		endTs, err := parseTimeToTimestamp(params.EndTime)
		if err != nil {
			return nil, fmt.Errorf("解析结束时间失败: %w", err)
		}
		endTime := larkcalendar.NewTimeInfoBuilder().
			Timestamp(endTs).
			Build()
		eventBuilder.EndTime(endTime)
	}

	if params.Location != "" {
		location := larkcalendar.NewEventLocationBuilder().
			Name(params.Location).
			Build()
		eventBuilder.Location(location)
	}

	req := larkcalendar.NewPatchCalendarEventReqBuilder().
		CalendarId(params.CalendarID).
		EventId(params.EventID).
		CalendarEvent(eventBuilder.Build()).
		Build()

	resp, err := client.Calendar.CalendarEvent.Patch(Context(), req)
	if err != nil {
		return nil, fmt.Errorf("更新日程失败: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("更新日程失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	if resp.Data == nil || resp.Data.Event == nil {
		return nil, fmt.Errorf("更新日程成功但未返回日程信息")
	}

	return convertEvent(resp.Data.Event), nil
}

// DeleteEvent 删除日程
func DeleteEvent(calendarID, eventID string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	req := larkcalendar.NewDeleteCalendarEventReqBuilder().
		CalendarId(calendarID).
		EventId(eventID).
		Build()

	resp, err := client.Calendar.CalendarEvent.Delete(Context(), req)
	if err != nil {
		return fmt.Errorf("删除日程失败: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("删除日程失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// 辅助函数：将 RFC3339 时间格式转换为时间戳字符串
func parseTimeToTimestamp(timeStr string) (string, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(t.Unix(), 10), nil
}

// 辅助函数：将时间戳字符串转换为 RFC3339 格式
func timestampToRFC3339(ts string, tz string) string {
	if ts == "" {
		return ""
	}
	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return ts
	}

	loc := time.Local
	if tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}

	return time.Unix(timestamp, 0).In(loc).Format(time.RFC3339)
}

// 辅助函数：转换日程对象
func convertEvent(event *larkcalendar.CalendarEvent) *CalendarEvent {
	if event == nil {
		return nil
	}

	result := &CalendarEvent{
		EventID:     StringVal(event.EventId),
		OrganizerID: StringVal(event.OrganizerCalendarId),
		Summary:     StringVal(event.Summary),
		Description: StringVal(event.Description),
		Status:      StringVal(event.Status),
		Visibility:  StringVal(event.Visibility),
		RecurringID: StringVal(event.RecurringEventId),
		IsException: BoolVal(event.IsException),
		AppLink:     StringVal(event.AppLink),
		Color:       IntVal(event.Color),
	}

	// 时区
	tz := ""
	if event.StartTime != nil && event.StartTime.Timezone != nil {
		tz = *event.StartTime.Timezone
		result.TimeZone = tz
	}

	// 时间转换
	if event.StartTime != nil && event.StartTime.Timestamp != nil {
		result.StartTime = timestampToRFC3339(*event.StartTime.Timestamp, tz)
	}
	if event.EndTime != nil && event.EndTime.Timestamp != nil {
		result.EndTime = timestampToRFC3339(*event.EndTime.Timestamp, tz)
	}
	if event.Location != nil && event.Location.Name != nil {
		result.Location = *event.Location.Name
	}
	if event.CreateTime != nil {
		result.CreateTime = timestampToRFC3339(*event.CreateTime, tz)
	}

	return result
}
