package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appErrors "github.com/solomonczyk/izborator/internal/errors"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
)

// BaseHandler базовый обработчик с общими методами для всех handlers
type BaseHandler struct {
	logger     *logger.Logger
	translator *i18n.Translator
}

// NewBaseHandler создает новый базовый обработчик
func NewBaseHandler(logger *logger.Logger, translator *i18n.Translator) *BaseHandler {
	return &BaseHandler{
		logger:     logger,
		translator: translator,
	}
}

// RespondJSON отправляет JSON ответ
func (h *BaseHandler) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// RespondAppError отправляет JSON ошибку из AppError с поддержкой i18n
func (h *BaseHandler) RespondAppError(w http.ResponseWriter, r *http.Request, err *appErrors.AppError) {
	message := h.resolveErrorMessage(r, err)

	h.logger.Error("API error response", map[string]interface{}{
		"code":    err.Code,
		"status":  err.HTTPStatus,
		"message": message,
	})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(err.HTTPStatus)
	if err := json.NewEncoder(w).Encode(appErrors.NewErrorResponse(err.Code, message)); err != nil {
		h.logger.Error("Failed to encode error response", map[string]interface{}{
			"error": err,
		})
	}
}

func (h *BaseHandler) resolveErrorMessage(r *http.Request, err *appErrors.AppError) string {
	message := err.Message
	if h.translator == nil {
		if message == "" {
			return "Internal server error"
		}
		return message
	}

	lang := httpMiddleware.GetLangFromContext(r.Context())
	messageKey := "api.errors." + strings.ToLower(err.Code)
	translated := h.translator.T(lang, messageKey)
	if translated != messageKey && translated != "" {
		return translated
	}

	if err.Code == appErrors.CodeInternalError {
		internalKey := "api.errors.internal"
		translated = h.translator.T(lang, internalKey)
		if translated != internalKey && translated != "" {
			return translated
		}
	}

	if message == "" {
		return "Internal server error"
	}

	return message
}

// ParseIntParam парсит целое число из query параметра с значением по умолчанию
func (h *BaseHandler) ParseIntParam(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultValue
	}
	return v
}

// ParseIntParam_Unsigned парсит беззнаковое целое число с проверкой >= 0
func (h *BaseHandler) ParseIntParamUnsigned(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultValue
	}
	return v
}
