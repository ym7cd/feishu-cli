package cmd

import (
	"testing"
	"time"
)

func TestParseVCTimestamp(t *testing.T) {
	t.Run("seconds", func(t *testing.T) {
		got, ok := parseVCTimestamp("1775725174")
		if !ok {
			t.Fatal("expected seconds timestamp to parse")
		}
		want := time.Unix(1775725174, 0)
		if !got.Equal(want) {
			t.Fatalf("seconds mismatch: got %v want %v", got, want)
		}
	})

	t.Run("milliseconds", func(t *testing.T) {
		got, ok := parseVCTimestamp("1775725174389")
		if !ok {
			t.Fatal("expected milliseconds timestamp to parse")
		}
		want := time.UnixMilli(1775725174389)
		if !got.Equal(want) {
			t.Fatalf("milliseconds mismatch: got %v want %v", got, want)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, ok := parseVCTimestamp("not-a-timestamp"); ok {
			t.Fatal("expected invalid timestamp to fail")
		}
	})
}

func TestFormatVCTimeMilliseconds(t *testing.T) {
	input := "1775725174389"
	want := time.UnixMilli(1775725174389).In(time.Local).Format("2006-01-02 15:04:05")

	if got := formatVCTime(input); got != want {
		t.Fatalf("formatVCTime(%q) = %q, want %q", input, got, want)
	}
}
