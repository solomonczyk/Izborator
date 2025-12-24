package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var (
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// ValidateUUID проверяет, что строка является валидным UUID
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("UUID cannot be empty")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}
	return nil
}

// ValidateURL проверяет, что строка является валидным URL
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}

// ValidatePagination проверяет параметры пагинации
func ValidatePagination(page, perPage int) error {
	if page < 1 {
		return fmt.Errorf("page must be at least 1, got %d", page)
	}
	if perPage < 1 {
		return fmt.Errorf("per_page must be at least 1, got %d", perPage)
	}
	if perPage > 100 {
		return fmt.Errorf("per_page cannot exceed 100, got %d", perPage)
	}
	return nil
}

// ValidatePrice проверяет, что цена валидна
func ValidatePrice(price float64) error {
	if price < 0 {
		return fmt.Errorf("price cannot be negative, got %.2f", price)
	}
	if price > 1000000000 { // 1 миллиард
		return fmt.Errorf("price seems unreasonably high: %.2f", price)
	}
	return nil
}

// ParseIntParam парсит целочисленный параметр из query string
func ParseIntParam(query url.Values, key string, defaultValue int) (int, error) {
	valueStr := query.Get(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %q is not a number", key, valueStr)
	}
	return value, nil
}

// ParseIntParamWithBounds парсит целочисленный параметр с проверкой границ
func ParseIntParamWithBounds(query url.Values, key string, defaultValue, min, max int) (int, error) {
	value, err := ParseIntParam(query, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value < min {
		return 0, fmt.Errorf("%s must be at least %d, got %d", key, min, value)
	}
	if value > max {
		return 0, fmt.Errorf("%s cannot exceed %d, got %d", key, max, value)
	}
	return value, nil
}

// SanitizeString очищает строку от потенциально опасных символов
func SanitizeString(s string) string {
	// Удаляем управляющие символы
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)
	return s
}

// ValidateSearchQuery проверяет поисковый запрос
func ValidateSearchQuery(query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}
	if len(query) < 2 {
		return fmt.Errorf("search query must be at least 2 characters")
	}
	if len(query) > 200 {
		return fmt.Errorf("search query cannot exceed 200 characters")
	}
	return nil
}
