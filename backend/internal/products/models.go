package products

import "time"

// ProductType тип продукта
type ProductType string

const (
	ProductTypeGood    ProductType = "good"    // Товар
	ProductTypeService ProductType = "service" // Услуга
)

// ServiceMetadata метаданные для услуг
type ServiceMetadata struct {
	Duration    string `json:"duration,omitempty"`     // Длительность услуги (например, "30 мин", "1 час")
	MasterName  string `json:"master_name,omitempty"`  // Имя мастера/специалиста
	ServiceArea string `json:"service_area,omitempty"` // Район обслуживания (например, "Нови-Сад", "Белград")
}

// Product каноническая карточка товара или услуги (универсальная модель Offer)
type Product struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Brand           string            `json:"brand"`
	Category        string            `json:"category"`              // Старое текстовое поле (для обратной совместимости)
	CategoryID      *string           `json:"category_id,omitempty"` // FK к categories.id
	ImageURL        string            `json:"image_url"`
	Specs           map[string]string `json:"specs"`
	Type            ProductType       `json:"type"`                    // "good" | "service"
	ServiceMetadata *ServiceMetadata  `json:"service_metadata,omitempty"` // Метаданные для услуг (JSONB)
	IsDeliverable   bool              `json:"is_deliverable"`        // Товар можно доставить (для товаров)
	IsOnsite         bool              `json:"is_onsite"`            // Услуга с выездом мастера (для услуг)
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// ProductPrice цена товара в конкретном магазине
type ProductPrice struct {
	ProductID string    `json:"product_id"`
	ShopID    string    `json:"shop_id"`
	ShopName  string    `json:"shop_name"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	URL       string    `json:"url"`
	InStock   bool      `json:"in_stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchResult результат поиска товаров
type SearchResult struct {
	Items  []*Product `json:"items"`
	Total  int        `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// BrowseProduct товар или услуга для каталога (с агрегированными ценами)
type BrowseProduct struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Brand           string            `json:"brand,omitempty"`
	Category        string            `json:"category,omitempty"`    // Старое текстовое поле
	CategoryID      *string           `json:"category_id,omitempty"` // FK к categories.id
	ImageURL        string            `json:"image_url,omitempty"`
	MinPrice         float64          `json:"min_price,omitempty"`
	MaxPrice         float64          `json:"max_price,omitempty"`
	Currency        string            `json:"currency,omitempty"`
	ShopsCount      int               `json:"shops_count,omitempty"`
	ShopNames       []string          `json:"shop_names"` // Список названий магазинов (без omitempty, чтобы всегда возвращался массив)
	Specs           map[string]string `json:"specs,omitempty"`
	Type            ProductType       `json:"type"`                    // "good" | "service"
	ServiceMetadata *ServiceMetadata  `json:"service_metadata,omitempty"` // Метаданные для услуг
	IsDeliverable   bool              `json:"is_deliverable"`        // Товар можно доставить
	IsOnsite         bool              `json:"is_onsite"`            // Услуга с выездом мастера
}

// BrowseParams параметры для каталога
type BrowseParams struct {
	Query       string
	Category    string   // slug категории (будет преобразован в category_id)
	CategoryID  *string  // ID категории (используется внутри)
	CategoryIDs []string // Список ID категорий (родитель + дочерние, для фильтрации)
	City        string   // slug города (будет преобразован в city_id)
	CityID      *string  // ID города (используется внутри)
	ShopID      string
	MinPrice    *float64
	MaxPrice    *float64
	Page        int
	PerPage     int
	Sort        string
}

// BrowseResult результат каталога
type BrowseResult struct {
	Items      []BrowseProduct `json:"items"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"total_pages"`
}
