package utils

import (
	"testing"
	"time"
)

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		input time.Time
		want  string
	}{
		{time.Now().Add(-500 * time.Millisecond), "just now"},
		{time.Now().Add(-1 * time.Second), "1s ago"},
		{time.Now().Add(-5 * time.Second), "5s ago"},
		{time.Now().Add(-1 * time.Minute), "1m ago"},
		{time.Now().Add(-5 * time.Minute), "5m ago"},
		{time.Now().Add(-1 * time.Hour), "1h ago"},
		{time.Now().Add(-3 * time.Hour), "3h ago"},
		{time.Now().Add(-24 * time.Hour), "1d ago"},
		{time.Now().Add(-3 * 24 * time.Hour), "3d ago"},
	}

	for _, tc := range tests {
		t.Run(tc.input.String(), func(t *testing.T) {
			got := FormatTimeAgo(tc.input)
			if got != tc.want {
				t.Errorf("FormatTimeAgo(%v) = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}
func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		input    string
		wantStr  string
		wantBool bool
	}{
		{"testname", "", true},
		{"&_-#+#-$- ", "Username can only contain letters, numbers, - and _", false},
		{"this name is supposed to be greater than 20 characters", "Username too long (max 20 characters)", false},
		{"", "Username too short (min 2 characters)", false},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			gotBool, gotStr := IsValidUsername(tc.input)
			if gotStr != tc.wantStr || gotBool != tc.wantBool {
				t.Errorf("IsValidUsername(%q) = gotBool: %v, gotStr: %q; want gotBool: %v, wantStr: %q",
					tc.input, gotBool, gotStr, tc.wantBool, tc.wantStr)
			}
		})
	}
}
