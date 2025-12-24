package attributes

// Attribute атрибут товара (RAM, Storage, Color, Size...)
type Attribute struct {
	ID           string
	Code         string
	NameSr       string
	DataType     string // "int", "float", "string", "bool", "enum"
	UnitSr       *string
	IsFilterable bool
	IsSortable   bool
}

// ProductTypeAttribute связь типа товара с атрибутом
type ProductTypeAttribute struct {
	ProductTypeID string
	AttributeID   string
	IsRequired    bool
	SortOrder     int
}
