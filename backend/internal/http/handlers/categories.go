package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/solomonczyk/izborator/internal/categories"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
)

// CategoriesHandler обработчик для работы с категориями
type CategoriesHandler struct {
	service    *categories.Service
	logger     *logger.Logger
	translator *i18n.Translator
}

// NewCategoriesHandler создаёт новый обработчик категорий
func NewCategoriesHandler(service *categories.Service, log *logger.Logger, translator *i18n.Translator) *CategoriesHandler {
	return &CategoriesHandler{
		service:    service,
		logger:     log,
		translator: translator,
	}
}

// GetTree обрабатывает получение дерева категорий
// GET /api/v1/categories/tree
func (h *CategoriesHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetTree()
	if err != nil {
		h.logger.Error("GetTree failed", map[string]interface{}{
			"error": err.Error(),
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.categories_load_failed")
		return
	}

	// Преобразуем в JSON структуру с children
	result := h.buildTree(tree)

	h.respondJSON(w, http.StatusOK, result)
}

// buildTree строит иерархическое дерево из плоского списка
func (h *CategoriesHandler) buildTree(cats []*categories.Category) []CategoryNode {
	// Создаём map для быстрого доступа
	categoryMap := make(map[string]*CategoryNode)
	var roots []CategoryNode

	// Сначала создаём все узлы
	for _, cat := range cats {
		node := &CategoryNode{
			ID:        cat.ID,
			Slug:      cat.Slug,
			Code:      cat.Code,
			NameSr:    cat.NameSr,
			NameSrLc:  cat.NameSrLc,
			Level:     cat.Level,
			IsActive:  cat.IsActive,
			SortOrder: cat.SortOrder,
			Children:  []CategoryNode{},
		}
		categoryMap[cat.ID] = node
	}

	// Затем связываем родители и дети
	for _, cat := range cats {
		node := categoryMap[cat.ID]
		if cat.ParentID == nil {
			// Корневая категория
			roots = append(roots, *node)
		} else {
			// Подкатегория - добавляем к родителю
			if parent, ok := categoryMap[*cat.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	// Сортируем корни по sort_order
	for i := 0; i < len(roots); i++ {
		for j := i + 1; j < len(roots); j++ {
			if roots[i].SortOrder > roots[j].SortOrder {
				roots[i], roots[j] = roots[j], roots[i]
			}
		}
	}

	return roots
}

// CategoryNode узел дерева категорий для JSON ответа
type CategoryNode struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug"`
	Code      string         `json:"code"`
	NameSr    string         `json:"name_sr"`
	NameSrLc  string         `json:"name_sr_lc"`
	Level     int            `json:"level"`
	IsActive  bool           `json:"is_active"`
	SortOrder int            `json:"sort_order"`
	Children  []CategoryNode `json:"children,omitempty"`
}

// respondJSON отправляет JSON ответ
func (h *CategoriesHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// respondError отправляет JSON ошибку
func (h *CategoriesHandler) respondError(w http.ResponseWriter, r *http.Request, status int, key string) {
	lang := httpMiddleware.GetLangFromContext(r.Context())
	message := h.translator.T(lang, key)
	if message == key || message == "" {
		// fallback на английский
		message = h.translator.T("en", key)
		if message == "" {
			message = key
		}
	}
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
