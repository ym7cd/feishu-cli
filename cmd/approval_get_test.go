package cmd

import "testing"

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
