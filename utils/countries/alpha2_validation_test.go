package countries

import (
	"testing"
)

func TestIsValidAlpha2Code(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"MY", true},  // Malaysia - valid Alpha-2 code
		{"US", true},  // United States - valid Alpha-2 code
		{"RU", true},  // Russia - valid Alpha-2 code
		{"ZZ", false}, // Invalid country code
		{"IN", true},  // India - valid Alpha-2 code
		{"GB", true},  // United Kingdom - valid Alpha-2 code
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := IsValidAlpha2Code(tt.code)
			if result != tt.expected {
				t.Errorf("isValidAlpha2Code(%v) = %v; want %v", tt.code, result, tt.expected)
			}
		})
	}
}
