package client

import "testing"

func TestBuildDownloadMediaExtra(t *testing.T) {
	tests := []struct {
		name string
		opts DownloadMediaOptions
		want string
	}{
		{
			name: "empty options",
			opts: DownloadMediaOptions{},
			want: "",
		},
		{
			name: "doc token defaults to docx",
			opts: DownloadMediaOptions{DocToken: "doc_token_123"},
			want: `{"doc_token":"doc_token_123","doc_type":"docx"}`,
		},
		{
			name: "doc type can be overridden",
			opts: DownloadMediaOptions{DocToken: "doc_token_123", DocType: "doc"},
			want: `{"doc_token":"doc_token_123","doc_type":"doc"}`,
		},
		{
			name: "raw extra wins",
			opts: DownloadMediaOptions{
				DocToken: "doc_token_123",
				DocType:  "docx",
				Extra:    `{"doc_token":"override","doc_type":"docx"}`,
			},
			want: `{"doc_token":"override","doc_type":"docx"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildDownloadMediaExtra(tt.opts); got != tt.want {
				t.Fatalf("buildDownloadMediaExtra() = %q, want %q", got, tt.want)
			}
		})
	}
}
