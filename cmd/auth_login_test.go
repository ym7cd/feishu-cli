package cmd

import "testing"

// TestFormatUserCode 测试 Device Flow 用户码格式化
func TestFormatUserCode(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"ABCD1234", "ABCD-1234"},
		{"ABCD-1234", "ABCD-1234"},
		{"ABC", "ABC"},
		{"ABCDEFGHIJ", "ABCDEFGHIJ"},
	}
	for _, c := range cases {
		got := formatUserCode(c.input)
		if got != c.want {
			t.Errorf("formatUserCode(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestResolveRequestedScope(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		domains   []string
		recommend bool
		want      string
	}{
		{
			name:  "explicit scope adds core scope",
			input: "minutes:minutes.basic:read minutes:minute:download",
			want:  "auth:user.id:read minutes:minutes.basic:read minutes:minute:download",
		},
		{
			name:    "domain all (no recommend)",
			domains: []string{"search"},
			want:    "auth:user.id:read search:docs:read search:message",
		},
	}

	for _, c := range cases {
		got, err := resolveRequestedScope(c.input, c.domains, c.recommend, true)
		if err != nil {
			t.Errorf("%s: resolveRequestedScope() error = %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: resolveRequestedScope(%q) = %q, want %q", c.name, c.input, got, c.want)
		}
	}
}

func TestResolveRequestedScopeRejectsMixedInput(t *testing.T) {
	_, err := resolveRequestedScope("search:docs:read", []string{"search"}, true, true)
	if err == nil {
		t.Fatal("expected error when --scope and --domain/--recommend are mixed")
	}
}
