package client

import (
	"encoding/json"
	"reflect"
	"testing"

	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
)

func TestParseEmbeddedJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{
			name:  "json object",
			input: `{"name":"leave"}`,
			want:  map[string]any{"name": "leave"},
		},
		{
			name:  "json array",
			input: `[{"id":"widget_1"}]`,
			want:  []any{map[string]any{"id": "widget_1"}},
		},
		{
			name:  "plain string fallback",
			input: `not-json`,
			want:  "not-json",
		},
		{
			name:  "empty string",
			input: `   `,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEmbeddedJSON(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseEmbeddedJSON(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestApprovalTaskToInfo(t *testing.T) {
	task := &larkapproval.Task{
		Topic:               stringPtr("1"),
		UserId:              stringPtr("ou_user"),
		Title:               stringPtr("审批标题"),
		Status:              stringPtr("PENDING"),
		ProcessStatus:       stringPtr("RUNNING"),
		DefinitionCode:      stringPtr("approval_code"),
		DefinitionName:      stringPtr("文档权限申请"),
		DefinitionId:        stringPtr("definition_id"),
		DefinitionGroupId:   stringPtr("group_id"),
		DefinitionGroupName: stringPtr("权限类"),
		TaskId:              stringPtr("task_id"),
		ProcessId:           stringPtr("process_id"),
		ProcessCode:         stringPtr("process_code"),
		Initiators:          []string{"ou_initiator"},
		InitiatorNames:      []string{"A"},
		Urls: &larkapproval.TaskUrls{
			Pc:     stringPtr("https://pc"),
			Mobile: stringPtr("https://mobile"),
		},
	}

	info := approvalTaskToInfo(task)

	if info.Topic != "1" {
		t.Fatalf("Topic = %q, want %q", info.Topic, "1")
	}
	if info.Title != "审批标题" {
		t.Fatalf("Title = %q, want %q", info.Title, "审批标题")
	}
	if info.PCURL != "https://pc" {
		t.Fatalf("PCURL = %q, want %q", info.PCURL, "https://pc")
	}
	if info.MobileURL != "https://mobile" {
		t.Fatalf("MobileURL = %q, want %q", info.MobileURL, "https://mobile")
	}
	if !reflect.DeepEqual(info.InitiatorNames, []string{"A"}) {
		t.Fatalf("InitiatorNames = %#v, want %#v", info.InitiatorNames, []string{"A"})
	}
}

func TestApprovalTaskStringUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "string",
			input: `"RUNNING"`,
			want:  "RUNNING",
		},
		{
			name:  "number",
			input: `42`,
			want:  "42",
		},
		{
			name:  "boolean",
			input: `true`,
			want:  "true",
		},
		{
			name:  "null",
			input: `null`,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got approvalTaskString
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("json.Unmarshal(%s) error = %v", tt.input, err)
			}
			if got.String() != tt.want {
				t.Fatalf("json.Unmarshal(%s) = %q, want %q", tt.input, got.String(), tt.want)
			}
		})
	}
}

func TestParseApprovalTaskQueryResponseHandlesNumericProcessStatus(t *testing.T) {
	body := []byte(`{
		"code": 0,
		"msg": "success",
		"data": {
			"page_token": "next_page",
			"has_more": true,
			"count": {
				"total": 1,
				"has_more": false
			},
			"tasks": [
				{
					"topic": 1,
					"user_id": "ou_user",
					"title": "审批标题",
					"status": "PENDING",
					"process_status": 12,
					"definition_name": "文档权限申请",
					"task_id": "task_id",
					"process_id": "process_id",
					"initiator_names": ["A"],
					"urls": {
						"pc": "https://pc"
					}
				}
			]
		}
	}`)

	result, err := parseApprovalTaskQueryResponse(body)
	if err != nil {
		t.Fatalf("parseApprovalTaskQueryResponse() error = %v", err)
	}

	if !result.HasMore {
		t.Fatalf("HasMore = %v, want true", result.HasMore)
	}
	if result.PageToken != "next_page" {
		t.Fatalf("PageToken = %q, want %q", result.PageToken, "next_page")
	}
	if result.Count == nil || result.Count.Total != 1 {
		t.Fatalf("Count = %#v, want total 1", result.Count)
	}
	if len(result.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, want 1", len(result.Tasks))
	}
	if result.Tasks[0].Topic != "1" {
		t.Fatalf("Topic = %q, want %q", result.Tasks[0].Topic, "1")
	}
	if result.Tasks[0].ProcessStatus != "12" {
		t.Fatalf("ProcessStatus = %q, want %q", result.Tasks[0].ProcessStatus, "12")
	}
	if result.Tasks[0].PCURL != "https://pc" {
		t.Fatalf("PCURL = %q, want %q", result.Tasks[0].PCURL, "https://pc")
	}
}

func stringPtr(v string) *string {
	return &v
}
