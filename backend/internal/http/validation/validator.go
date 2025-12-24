package validation

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)


// ValidateUUID ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃ‘â€š, Ã‘â€¡Ã‘â€šÃÂ¾ Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ¾ÃÂºÃÂ° Ã‘ÂÃÂ²ÃÂ»Ã‘ÂÃÂµÃ‘â€šÃ‘ÂÃ‘Â ÃÂ²ÃÂ°ÃÂ»ÃÂ¸ÃÂ´ÃÂ½Ã‘â€¹ÃÂ¼ UUID
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("UUID cannot be empty")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}
	return nil
}

// ValidateURL ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃ‘â€š, Ã‘â€¡Ã‘â€šÃÂ¾ Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ¾ÃÂºÃÂ° Ã‘ÂÃÂ²ÃÂ»Ã‘ÂÃÂµÃ‘â€šÃ‘ÂÃ‘Â ÃÂ²ÃÂ°ÃÂ»ÃÂ¸ÃÂ´ÃÂ½Ã‘â€¹ÃÂ¼ URL
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

// ValidatePagination ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃ‘â€š ÃÂ¿ÃÂ°Ã‘â‚¬ÃÂ°ÃÂ¼ÃÂµÃ‘â€šÃ‘â‚¬Ã‘â€¹ ÃÂ¿ÃÂ°ÃÂ³ÃÂ¸ÃÂ½ÃÂ°Ã‘â€ ÃÂ¸ÃÂ¸
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

// ValidatePrice ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃ‘â€š, Ã‘â€¡Ã‘â€šÃÂ¾ Ã‘â€ ÃÂµÃÂ½ÃÂ° ÃÂ²ÃÂ°ÃÂ»ÃÂ¸ÃÂ´ÃÂ½ÃÂ°
func ValidatePrice(price float64) error {
	if price < 0 {
		return fmt.Errorf("price cannot be negative, got %.2f", price)
	}
	if price > 1000000000 { // 1 ÃÂ¼ÃÂ¸ÃÂ»ÃÂ»ÃÂ¸ÃÂ°Ã‘â‚¬ÃÂ´
		return fmt.Errorf("price seems unreasonably high: %.2f", price)
	}
	return nil
}

// ParseIntParam ÃÂ¿ÃÂ°Ã‘â‚¬Ã‘ÂÃÂ¸Ã‘â€š Ã‘â€ ÃÂµÃÂ»ÃÂ¾Ã‘â€¡ÃÂ¸Ã‘ÂÃÂ»ÃÂµÃÂ½ÃÂ½Ã‘â€¹ÃÂ¹ ÃÂ¿ÃÂ°Ã‘â‚¬ÃÂ°ÃÂ¼ÃÂµÃ‘â€šÃ‘â‚¬ ÃÂ¸ÃÂ· query string
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

// ParseIntParamWithBounds ÃÂ¿ÃÂ°Ã‘â‚¬Ã‘ÂÃÂ¸Ã‘â€š Ã‘â€ ÃÂµÃÂ»ÃÂ¾Ã‘â€¡ÃÂ¸Ã‘ÂÃÂ»ÃÂµÃÂ½ÃÂ½Ã‘â€¹ÃÂ¹ ÃÂ¿ÃÂ°Ã‘â‚¬ÃÂ°ÃÂ¼ÃÂµÃ‘â€šÃ‘â‚¬ Ã‘Â ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬ÃÂºÃÂ¾ÃÂ¹ ÃÂ³Ã‘â‚¬ÃÂ°ÃÂ½ÃÂ¸Ã‘â€ 
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

// SanitizeString ÃÂ¾Ã‘â€¡ÃÂ¸Ã‘â€°ÃÂ°ÃÂµÃ‘â€š Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ¾ÃÂºÃ‘Æ’ ÃÂ¾Ã‘â€š ÃÂ¿ÃÂ¾Ã‘â€šÃÂµÃÂ½Ã‘â€ ÃÂ¸ÃÂ°ÃÂ»Ã‘Å’ÃÂ½ÃÂ¾ ÃÂ¾ÃÂ¿ÃÂ°Ã‘ÂÃÂ½Ã‘â€¹Ã‘â€¦ Ã‘ÂÃÂ¸ÃÂ¼ÃÂ²ÃÂ¾ÃÂ»ÃÂ¾ÃÂ²
func SanitizeString(s string) string {
	// ÃÂ£ÃÂ´ÃÂ°ÃÂ»Ã‘ÂÃÂµÃÂ¼ Ã‘Æ’ÃÂ¿Ã‘â‚¬ÃÂ°ÃÂ²ÃÂ»Ã‘ÂÃ‘Å½Ã‘â€°ÃÂ¸ÃÂµ Ã‘ÂÃÂ¸ÃÂ¼ÃÂ²ÃÂ¾ÃÂ»Ã‘â€¹
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)
	return s
}

// ValidateSearchQuery ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃ‘â€š ÃÂ¿ÃÂ¾ÃÂ¸Ã‘ÂÃÂºÃÂ¾ÃÂ²Ã‘â€¹ÃÂ¹ ÃÂ·ÃÂ°ÃÂ¿Ã‘â‚¬ÃÂ¾Ã‘Â
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
