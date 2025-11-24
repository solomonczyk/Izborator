package products

import "errors"

var (
	// ErrProductNotFound товар не найден
	ErrProductNotFound = errors.New("product not found")
	
	// ErrInvalidProductID невалидный ID товара
	ErrInvalidProductID = errors.New("invalid product ID")
	
	// ErrProductAlreadyExists товар уже существует
	ErrProductAlreadyExists = errors.New("product already exists")
	
	// ErrInvalidSearchQuery невалидный поисковый запрос
	ErrInvalidSearchQuery = errors.New("invalid search query")
)

