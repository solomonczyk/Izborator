package handlers

import (
	"net/http"

	"github.com/solomonczyk/izborator/internal/categories"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
)

// CategoriesHandler обработчик для работы с категориями
type CategoriesHandler struct {
	*BaseHandler
	service *categories.Service
}

// NewCategoriesHandler создаёт новый обработчик категорий
func NewCategoriesHandler(service *categories.Service, log *logger.Logger, translator *i18n.Translator) *CategoriesHandler {
	return &CategoriesHandler{
		BaseHandler: NewBaseHandler(log, translator),
		service:     service,
	}
}

// GetTree обрабатывает получение дерева категорий
// GET /api/v1/categories/tree
func (h *CategoriesHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetTree()
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to load categories tree", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Если нет категорий, сразу возвращаем пустой массив
	if len(tree) == 0 {
		emptyArray := []CategoryNode{}
		h.RespondJSON(w, http.StatusOK, emptyArray)
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

	h.RespondJSON(w, http.StatusOK, finalResult)
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
