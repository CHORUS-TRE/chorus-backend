package middleware

import (
	"testing"
)

func TestIsOriginInAllowedList(t *testing.T) {
	tests := []struct {
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{"http://example.com", []string{"http://example.com"}, true},
		{"http://example.com", []string{"http://example.org"}, false},
		{"http://example.com/abc/def", []string{"http://example.com", "http://example.org"}, true},
		{"", []string{"http://example.com"}, false},
	}

	for _, test := range tests {
		result := isOriginInAllowedList(test.origin, test.allowedOrigins)
		if result != test.expected {
			t.Errorf("Expected %v, got %v for origin %s with allowed origins %v", test.expected, result, test.origin, test.allowedOrigins)
		}
	}
}
