package cmd

import (
	"reflect"
	"testing"
)

func TestParseTableColumnWidthFlag(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		wantMode  string
		wantVals  []int
		wantError bool
	}{
		{"empty -> auto", "", "auto", nil, false},
		{"auto", "auto", "auto", nil, false},
		{"fixed", "fixed", "fixed", nil, false},
		{"explicit ints", "80,200,120", "explicit", []int{80, 200, 120}, false},
		{"with spaces", " 80 , 200 , 120 ", "explicit", []int{80, 200, 120}, false},
		{"star placeholder", "80,*,120", "explicit", []int{80, 0, 120}, false},
		{"empty between commas", "80,,120", "explicit", []int{80, 0, 120}, false},
		{"invalid token", "80,xxx,120", "", nil, true},
		{"negative", "80,-50,120", "", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mode, values, err := parseTableColumnWidthFlag(tc.in)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil (mode=%q, vals=%v)", mode, values)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mode != tc.wantMode {
				t.Errorf("mode=%q, want %q", mode, tc.wantMode)
			}
			if !reflect.DeepEqual(values, tc.wantVals) {
				t.Errorf("values=%v, want %v", values, tc.wantVals)
			}
		})
	}
}
