package cmd

import (
	"strings"
	"testing"
)

func TestValidateApprovalCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid uuid-like code",
			code:    "7C468A54-8745-2245-9675-08B7C63E7A85",
			wantErr: false,
		},
		{
			name:    "valid short code",
			code:    "approval_123",
			wantErr: false,
		},
		{
			name:    "invalid blank code",
			code:    "",
			wantErr: true,
		},
		{
			name:    "invalid code with slash",
			code:    "approval/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApprovalCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateApprovalCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

func TestApprovalGetOutputFlagSupportsRawJSON(t *testing.T) {
	flag := approvalGetCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("approvalGetCmd should register --output")
	}
	if got := flag.Usage; got != "输出格式（json/raw-json）" {
		t.Fatalf("--output usage = %q, want %q", got, "输出格式（json/raw-json）")
	}
}

func TestApprovalGetLongHelpMentionsRawJSONAndAlignedExamples(t *testing.T) {
	if !strings.Contains(approvalGetCmd.Long, "--output raw-json") {
		t.Fatalf("approvalGetCmd.Long should mention raw-json, got %q", approvalGetCmd.Long)
	}
	if !strings.Contains(approvalGetCmd.Long, "\n示例:\n") {
		t.Fatalf("approvalGetCmd.Long should align 示例 header with other commands, got %q", approvalGetCmd.Long)
	}
}
