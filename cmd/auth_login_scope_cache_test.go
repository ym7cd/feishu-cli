package cmd

import "testing"

func TestLoginRequestedScopeCacheRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const (
		deviceCode     = "device-code-1"
		requestedScope = "auth:user.id:read search:docs:read"
	)

	if err := saveLoginRequestedScope(deviceCode, requestedScope); err != nil {
		t.Fatalf("saveLoginRequestedScope() error = %v", err)
	}

	got, err := loadLoginRequestedScope(deviceCode)
	if err != nil {
		t.Fatalf("loadLoginRequestedScope() error = %v", err)
	}
	if got != requestedScope {
		t.Fatalf("loadLoginRequestedScope() = %q, want %q", got, requestedScope)
	}

	if err := removeLoginRequestedScope(deviceCode); err != nil {
		t.Fatalf("removeLoginRequestedScope() error = %v", err)
	}

	got, err = loadLoginRequestedScope(deviceCode)
	if err != nil {
		t.Fatalf("loadLoginRequestedScope() after remove error = %v", err)
	}
	if got != "" {
		t.Fatalf("loadLoginRequestedScope() after remove = %q, want empty", got)
	}
}
