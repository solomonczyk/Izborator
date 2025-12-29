package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/solomonczyk/izborator/internal/http/response"
)

var validate = validator.New()

// ValidationErrorDetail детали ошибки валидации
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
}

// ValidateStruct валидирует структуру
func ValidateStruct(data interface{}) *response.AppError {
	if err := validate.Struct(data); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		details := make([]ValidationErrorDetail, len(validationErrors))

		for i, fe := range validationErrors {
			details[i] = ValidationErrorDetail{
				Field:   fe.Field(),
				Message: getValidationMessage(fe),
				Tag:     fe.Tag(),
			}
		}

		appErr := response.NewAppError(response.ErrorValidationFailed, "Validation failed")
		return appErr.WithDetails(map[string]interface{}{
			"errors": details,
		})
	}

	return nil
}

// ValidateQuery валидирует query параметры
func ValidateQuery(r *http.Request, params map[string]string) *response.AppError {
	query := r.URL.Query()
	errors := []ValidationErrorDetail{}

	for param, rule := range params {
		value := query.Get(param)
		if detail := validateQueryParam(param, value, rule); detail != nil {
			errors = append(errors, *detail)
		}
	}

	if len(errors) > 0 {
		appErr := response.NewAppError(response.ErrorInvalidInput, "Invalid query parameters")
		return appErr.WithDetails(map[string]interface{}{
			"errors": errors,
		})
	}

	return nil
}

// validateQueryParam валидирует один параметр
func validateQueryParam(name, value, rule string) *ValidationErrorDetail {
	rules := strings.Split(rule, ",")

	for _, r := range rules {
		r = strings.TrimSpace(r)

		if r == "required" && value == "" {
			return &ValidationErrorDetail{
				Field:   name,
				Message: fmt.Sprintf("%s is required", name),
				Tag:     "required",
			}
		}

		if r == "number" && value != "" {
			if _, err := strconv.Atoi(value); err != nil {
				return &ValidationErrorDetail{
					Field:   name,
					Message: fmt.Sprintf("%s must be a number", name),
					Tag:     "number",
				}
			}
		}

		if r == "email" && value != "" {
			if err := validate.Var(value, "email"); err != nil {
				return &ValidationErrorDetail{
					Field:   name,
					Message: fmt.Sprintf("%s must be a valid email", name),
					Tag:     "email",
				}
			}
		}

		if r == "uuid" && value != "" {
			if err := validate.Var(value, "uuid"); err != nil {
				return &ValidationErrorDetail{
					Field:   name,
					Message: fmt.Sprintf("%s must be a valid UUID", name),
					Tag:     "uuid",
				}
			}
		}

		if r == "url" && value != "" {
			if err := validate.Var(value, "url"); err != nil {
				return &ValidationErrorDetail{
					Field:   name,
					Message: fmt.Sprintf("%s must be a valid URL", name),
					Tag:     "url",
				}
			}
		}

		if strings.HasPrefix(r, "min=") {
			minStr := strings.TrimPrefix(r, "min=")
			minLen, err := strconv.Atoi(minStr)
			if err != nil {
				continue
			}
			if len(value) < minLen {
				return &ValidationErrorDetail{
					Field:   name,
					Message: fmt.Sprintf("%s must be at least %d characters", name, minLen),
					Tag:     "min",
				}
			}
		}

		if strings.HasPrefix(r, "max=") {
			maxStr := strings.TrimPrefix(r, "max=")
			maxLen, err := strconv.Atoi(maxStr)
			if err != nil {
				continue
			}
			if len(value) > maxLen {
				return &ValidationErrorDetail{
					Field:   name,
					Message: fmt.Sprintf("%s must not exceed %d characters", name, maxLen),
					Tag:     "max",
				}
			}
		}
	}

	return nil
}

// getValidationMessage возвращает пользовательское сообщение для ошибки валидации
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "numeric":
		return fmt.Sprintf("%s must be numeric", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	default:
		return fmt.Sprintf("%s validation failed: %s", fe.Field(), fe.Tag())
	}
}
