package autoconfig

import (
	"testing"
)

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple HTML",
			input:    "<html><body><p>Test</p></body></html>",
			expected: "Test",
			wantErr:  false,
		},
		{
			name:     "HTML with scripts",
			input:    "<html><head><script>alert('test')</script></head><body>Content</body></html>",
			expected: "Content",
			wantErr:  false,
		},
		{
			name:     "empty HTML",
			input:    "",
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CleanHTML(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("CleanHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}

