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
		// Возвращаем пустой массив вместо ошибки, чтобы фронтенд не ломался
		emptyArray := []CategoryNode{}
		h.respondJSON(w, http.StatusOK, emptyArray)
		return
	}

	// Если нет категорий, сразу возвращаем пустой массив
	if len(tree) == 0 {
		emptyArray := []CategoryNode{}
		h.respondJSON(w, http.StatusOK, emptyArray)
		return
	}

	// Преобразуем в JSON структуру с children
	result := h.buildTree(tree)

	// Убеждаемся, что возвращаем массив, даже если пустой
	// Всегда создаем новый слайс явно
	var finalResult []CategoryNode
	if len(result) == 0 {
		finalResult = []CategoryNode{}
	} else {
		finalResult = append([]CategoryNode{}, result...)
	}

	h.respondJSON(w, http.StatusOK, finalResult)
}

// buildTree строит иерархическое дерево из плоского списка
func (h *CategoriesHandler) buildTree(cats []*categories.Category) []CategoryNode {
	// Если нет категорий, возвращаем пустой массив (не nil)
	if len(cats) == 0 {
		return []CategoryNode{}
	}

	// Создаём map для быстрого доступа (храним указатели)
	categoryMap := make(map[string]*CategoryNode)
	var rootNodes []*CategoryNode

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
			// Корневая категория - сохраняем указатель
			rootNodes = append(rootNodes, node)
		} else {
			// Подкатегория - добавляем к родителю
			if parent, ok := categoryMap[*cat.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	// Сортируем корни по sort_order
	for i := 0; i < len(rootNodes); i++ {
		for j := i + 1; j < len(rootNodes); j++ {
			if rootNodes[i].SortOrder > rootNodes[j].SortOrder {
				rootNodes[i], rootNodes[j] = rootNodes[j], rootNodes[i]
			}
		}
	}

	// Преобразуем указатели в значения для возврата
	roots := make([]CategoryNode, 0, len(rootNodes))
	for _, node := range rootNodes {
		roots = append(roots, *node)
	}

	// Убеждаемся, что возвращаем не-nil слайс
	if roots == nil {
		return []CategoryNode{}
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// Специальная обработка для []CategoryNode - всегда возвращаем массив
	if slice, ok := data.([]CategoryNode); ok {
		// Если это слайс категорий, всегда сериализуем как массив
		if len(slice) == 0 {
			// Явно пишем пустой массив
			_, _ = w.Write([]byte("[]\n"))
			return
		}
		// Сериализуем непустой слайс
		jsonBytes, err := json.Marshal(slice)
		if err != nil {
			h.logger.Error("Failed to marshal JSON response", map[string]interface{}{
				"error": err,
			})
			_, _ = w.Write([]byte("[]\n"))
			return
		}
		_, _ = w.Write(jsonBytes)
		return
	}

	// Для других типов используем стандартную сериализацию
	if data == nil {
		_, _ = w.Write([]byte("[]\n"))
		return
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Failed to marshal JSON response", map[string]interface{}{
			"error": err,
		})
		_, _ = w.Write([]byte("[]\n"))
		return
	}

	_, _ = w.Write(jsonBytes)
}

// respondError отправляет JSON ошибку
// Используется для обработки ошибок в будущем
func (h *CategoriesHandler) respondError(w http.ResponseWriter, r *http.Request, status int, key string) { //nolint:unused
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
