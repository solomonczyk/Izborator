package categories

// Category категория товара (универсальная таксономия)
type Category struct {
	ID        string
	ParentID  *string
	Slug      string
	Code      string // Внутренний код: "ELEKTRONIKA", "PHONES", "LAPTOPOVI"
	NameSr    string // "Elektronika", "Mobilni telefoni"
	NameSrLc  string // "elektronika", "mobilni telefoni" (для поиска/сортировки)
	Level     int    // 1 = раздел, 2 = категория, 3 = подкатегория
	IsActive  bool
	SortOrder int
}
