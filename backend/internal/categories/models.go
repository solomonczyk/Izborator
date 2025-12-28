package categories

// Category категория товара (универсальная таксономия)
type Category struct {
	ID        string
	ParentID  *string
	Slug      string
	Code      string // Внутренний код: "ELEKTRONIKA", "PHONES", "LAPTOPOVI"
	NameSr    string // "Elektronika", "Mobilni telefoni"
	NameSrLc  string // "elektronika", "mobilni telefoni" (для поиска/сортировки)
	NameRu    *string // "Электроника", "Мобильные телефоны"
	NameEn    *string // "Electronics", "Mobile Phones"
	NameHu    *string // "Elektronika", "Mobiltelefonok"
	NameZh    *string // "电子产品", "手机"
	Level     int    // 1 = раздел, 2 = категория, 3 = подкатегория
	IsActive  bool
	SortOrder int
}

// GetName возвращает название категории на указанном языке
// Если перевода нет, возвращает сербское название
func (c *Category) GetName(locale string) string {
	switch locale {
	case "ru":
		if c.NameRu != nil && *c.NameRu != "" {
			return *c.NameRu
		}
	case "en":
		if c.NameEn != nil && *c.NameEn != "" {
			return *c.NameEn
		}
	case "hu":
		if c.NameHu != nil && *c.NameHu != "" {
			return *c.NameHu
		}
	case "zh":
		if c.NameZh != nil && *c.NameZh != "" {
			return *c.NameZh
		}
	}
	// По умолчанию возвращаем сербское название
	return c.NameSr
}
