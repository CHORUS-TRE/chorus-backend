package service

import (
	"testing"
)

func TestGetWorkspaceID(t *testing.T) {
	s := &WorkbenchService{}
	tests := []struct {
		name      string
		input     string
		expected  uint64
		expectErr bool
	}{
		{
			name:      "valid workspace name",
			input:     "workspace123",
			expected:  123,
			expectErr: false,
		},
		{
			name:      "invalid workspace name",
			input:     "invalid123",
			expected:  0,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := s.getWorkspaceID(test.input)
			if (err != nil) != test.expectErr {
				t.Errorf("expected error: %v, got: %v", test.expectErr, err)
			}
			if result != test.expected {
				t.Errorf("expected: %d, got: %d", test.expected, result)
			}
		})
	}
}
