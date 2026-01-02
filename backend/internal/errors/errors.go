package errors

import (
	"fmt"
	"net/http"
)

// AppError представляет ошибку приложения с кодом и сообщением
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"` // Не сериализуется в JSON
	Err        error  `json:"-"` // Оригинальная ошибка для логирования
	Details    map[string]interface{} `json:"-"`
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap возвращает оригинальную ошибку
func (e *AppError) Unwrap() error {
	return e.Err
}

// Предопределенные коды ошибок
const (
	// Общие ошибки
	CodeInternalError    = "INTERNAL_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeNotFound         = "NOT_FOUND"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeValidationFailed = "VALIDATION_FAILED"

	// Ошибки товаров
	CodeProductNotFound   = "PRODUCT_NOT_FOUND"
	CodeInvalidProductID  = "INVALID_PRODUCT_ID"
	CodeProductLoadFailed = "PRODUCT_LOAD_FAILED"

	// Ошибки поиска
	CodeSearchFailed = "SEARCH_FAILED"
	CodeInvalidQuery = "INVALID_QUERY"
	CodeBrowseFailed = "BROWSE_FAILED"
	CodeRateLimited  = "RATE_LIMITED"

	// Ошибки цен
	CodePriceHistoryFailed = "PRICE_HISTORY_FAILED"
	CodeInvalidPagination  = "INVALID_PAGINATION"

	// Ошибки категорий
	CodeCategoryNotFound = "CATEGORY_NOT_FOUND"

	// Ошибки городов
	CodeCityNotFound = "CITY_NOT_FOUND"
)

// NewAppError создает новую ошибку приложения
func NewAppError(code, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// NewAppErrorWithDetails creates an AppError with optional details payload.
func NewAppErrorWithDetails(code, message string, httpStatus int, err error, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
		Details:    details,
	}
}


// NewInternalError создает ошибку внутренней ошибки сервера
func NewInternalError(message string, err error) *AppError {
	return NewAppError(CodeInternalError, message, http.StatusInternalServerError, err)
}

// NewBadRequest создает ошибку неверного запроса
func NewBadRequest(message string, err error) *AppError {
	return NewAppError(CodeBadRequest, message, http.StatusBadRequest, err)
}

// NewNotFound создает ошибку "не найдено"
func NewNotFound(message string) *AppError {
	return NewAppError(CodeNotFound, message, http.StatusNotFound, nil)
}

// NewValidationError создает ошибку валидации
func NewValidationError(message string, err error) *AppError {
	return NewAppError(CodeValidationFailed, message, http.StatusBadRequest, err)
}

// WrapError оборачивает существующую ошибку в AppError
func WrapError(err error, code, message string, httpStatus int) *AppError {
	if err == nil {
		return nil
	}

	// Если уже AppError, возвращаем как есть
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewAppError(code, message, httpStatus, err)
}

// ToHTTPError преобразует ошибку в HTTP статус и сообщение
func ToHTTPError(err error) (int, string, string) {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus, appErr.Code, appErr.Message
	}

	// По умолчанию - внутренняя ошибка
	return http.StatusInternalServerError, CodeInternalError, "Internal server error"
}
