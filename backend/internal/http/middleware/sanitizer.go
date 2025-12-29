package middleware

import (
	"html"
	"strings"
	"unicode"
)

// Sanitizer функции для очистки входных данных
type Sanitizer struct{}

// TrimWhitespace удаляет пробелы в начале и конце
func (s *Sanitizer) TrimWhitespace(str string) string {
	return strings.TrimSpace(str)
}

// HTMLEscape экранирует HTML специальные символы
func (s *Sanitizer) HTMLEscape(str string) string {
	return html.EscapeString(str)
}

// StripHTML удаляет HTML теги
func (s *Sanitizer) StripHTML(str string) string {
	// Простой метод - удаляем все в <...>
	var result strings.Builder
	inTag := false

	for _, r := range str {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// RemoveControlCharacters удаляет управляющие символы
func (s *Sanitizer) RemoveControlCharacters(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, str)
}

// NormalizeWhitespace заменяет множественные пробелы на один
func (s *Sanitizer) NormalizeWhitespace(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

// SanitizeString полная очистка строки
func (s *Sanitizer) SanitizeString(str string) string {
	// 1. Trim whitespace
	str = s.TrimWhitespace(str)
	// 2. Remove control characters
	str = s.RemoveControlCharacters(str)
	// 3. Normalize whitespace
	str = s.NormalizeWhitespace(str)
	// 4. Escape HTML
	str = s.HTMLEscape(str)
	return str
}

// SanitizeSearchQuery очищает поисковый запрос
func (s *Sanitizer) SanitizeSearchQuery(query string) string {
	// Удаляем теги и управляющие символы
	query = s.StripHTML(query)
	query = s.RemoveControlCharacters(query)
	query = s.TrimWhitespace(query)
	query = s.NormalizeWhitespace(query)
	return query
}

// NewSanitizer создает новый очиститель
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}
