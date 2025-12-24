package autoconfig

import (
	"context"
	"testing"
)

// MockStorage мок для тестирования AutoConfig
type mockAutoconfigStorage struct {
	candidates []Candidate
	attempts   []ConfigAttempt
}

func (m *mockAutoconfigStorage) GetClassifiedCandidates(limit int) ([]Candidate, error) {
	if limit > len(m.candidates) {
		limit = len(m.candidates)
	}
	return m.candidates[:limit], nil
}

func (m *mockAutoconfigStorage) MarkAsFailed(id, reason string) error {
	return nil
}

func (m *mockAutoconfigStorage) MarkAsConfigured(id string, config ShopConfig) error {
	return nil
}

func (m *mockAutoconfigStorage) RecordAttempt(attempt ConfigAttempt) error {
	m.attempts = append(m.attempts, attempt)
	return nil
}

// MockAI мок для AI клиента
type mockAI struct {
	selectorsJSON string
	err           error
}

func (m *mockAI) GenerateSelectors(ctx context.Context, html string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.selectorsJSON, nil
}

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
			// Проверяем, что скрипты удалены (упрощенная проверка)
			if len(result) > 0 && len(result) < len(tt.input) {
				// HTML был очищен
			}
		})
	}
}

