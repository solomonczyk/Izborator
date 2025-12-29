package response

import "net/http"

// ErrorCode коды ошибок приложения
type ErrorCode string

const (
	// Validation errors
	ErrorInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorMissingField     ErrorCode = "MISSING_FIELD"
	ErrorInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// Resource errors
	ErrorNotFound      ErrorCode = "NOT_FOUND"
	ErrorAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrorConflict      ErrorCode = "CONFLICT"

	// Database errors
	ErrorDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrorQueryFailed   ErrorCode = "QUERY_FAILED"

	// External service errors
	ErrorExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorTimeout         ErrorCode = "TIMEOUT"

	// Server errors
	ErrorInternal       ErrorCode = "INTERNAL_ERROR"
	ErrorNotImplemented ErrorCode = "NOT_IMPLEMENTED"
	ErrorUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrorForbidden      ErrorCode = "FORBIDDEN"
)

// ErrorResponse стандартная структура ответа об ошибке
type ErrorResponse struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	TraceID string                 `json:"trace_id,omitempty"`
}

// ErrorMap маппинг кодов ошибок на HTTP статус коды
var ErrorMap = map[ErrorCode]int{
	ErrorInvalidInput:     http.StatusBadRequest,
	ErrorValidationFailed: http.StatusBadRequest,
	ErrorMissingField:     http.StatusBadRequest,
	ErrorInvalidFormat:    http.StatusBadRequest,

	ErrorNotFound:      http.StatusNotFound,
	ErrorAlreadyExists: http.StatusConflict,
	ErrorConflict:      http.StatusConflict,

	ErrorDatabaseError: http.StatusInternalServerError,
	ErrorQueryFailed:   http.StatusInternalServerError,

	ErrorExternalService: http.StatusBadGateway,
	ErrorTimeout:         http.StatusGatewayTimeout,

	ErrorInternal:       http.StatusInternalServerError,
	ErrorNotImplemented: http.StatusNotImplemented,
	ErrorUnauthorized:   http.StatusUnauthorized,
	ErrorForbidden:      http.StatusForbidden,
}

// AppError представляет ошибку приложения
type AppError struct {
	Code        ErrorCode
	Message     string
	StatusCode  int
	Details     map[string]interface{}
	OriginalErr error
}

// NewAppError создает новую ошибку приложения
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: ErrorMap[code],
	}
}

// NewAppErrorWithStatus создает ошибку с кастомным статус кодом
func NewAppErrorWithStatus(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails добавляет детали к ошибке
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithOriginalError добавляет оригинальную ошибку
func (e *AppError) WithOriginalError(err error) *AppError {
	e.OriginalErr = err
	return e
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	return e.Message
}

// ToResponse преобразует AppError в ErrorResponse для клиента
func (e *AppError) ToResponse() *ErrorResponse {
	return &ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}
