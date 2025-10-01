package ai

import (
	"errors"
	"testing"
)

func TestFormatAIError(t *testing.T) {
	tests := []struct {
		name  string
		input error
		want  string
	}{
		{"rate limit", errors.New("rate limit exceeded"), "AI Error: Rate limit reached. Please wait and try again."},
		{"quota", errors.New("quota exceeded"), "AI Error: Quota reached. Try later."},
		{"invalid prompt", errors.New("invalid prompt"), "AI Error: Your prompt is invalid."},
		{"other error", errors.New("API key missing"), "AI Error: Please try again later."},
		{"nil error", nil, "AI features enabled"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := FormatAIError(tc.input) // hypothetical function
			if got != tc.want {
				t.Errorf("FormatAIError(%v) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
