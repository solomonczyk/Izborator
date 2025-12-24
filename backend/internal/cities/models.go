package cities

// City город Сербии
type City struct {
	ID        string
	Slug      string
	NameSr    string
	RegionSr  *string
	SortOrder int
	IsActive  bool
}
