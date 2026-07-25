package auth

import "testing"

func TestSafeReturnTo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string defaults to /", "", "/"},
		{"valid root path", "/", "/"},
		{"valid nested path", "/dashboard", "/dashboard"},
		{"valid path with query", "/vehicles?id=123", "/vehicles?id=123"},
		{"valid complex path", "/vehicles/abc/mileage?tab=history", "/vehicles/abc/mileage?tab=history"},

		// Open redirect attacks
		{"absolute URL with https", "https://evil.com", "/"},
		{"absolute URL with http", "http://evil.com/steal", "/"},
		{"protocol-relative URL", "//evil.com", "/"},
		{"protocol-relative with path", "//evil.com/phish", "/"},
		{"javascript scheme", "javascript://alert(1)", "/"},
		{"data scheme", "data://text/html,<h1>pwned</h1>", "/"},

		// Edge cases
		{"backslash trick", "/\\evil.com", "/"},
		{"no leading slash", "evil.com", "/"},
		{"relative path without slash", "dashboard", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeReturnTo(tt.input)
			if got != tt.expected {
				t.Errorf("SafeReturnTo(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
