package response

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse стандартная структура успешного ответа
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta *MetaInfo   `json:"meta,omitempty"`
}

// MetaInfo информация о ответе (пагинация, статистика и т.д.)
type MetaInfo struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// WriteJSON пишет JSON ответ с правильными заголовками
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteError пишет ошибку в стандартном формате
func WriteError(w http.ResponseWriter, err *AppError) error {
	return WriteJSON(w, err.StatusCode, err.ToResponse())
}

// WriteSuccess пишет успешный ответ
func WriteSuccess(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// WriteCreated пишет ответ с кодом 201 Created
func WriteCreated(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusCreated, SuccessResponse{Data: data})
}

// WriteNoContent пишет пустой ответ с кодом 204
func WriteNoContent(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	return nil
}
