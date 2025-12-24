package matching

import "errors"

var (
	// ErrMatchNotFound сопоставление не найдено
	ErrMatchNotFound = errors.New("match not found")

	// ErrInvalidSimilarity невалидное значение схожести
	ErrInvalidSimilarity = errors.New("invalid similarity value")

	// ErrInsufficientData недостаточно данных для сопоставления
	ErrInsufficientData = errors.New("insufficient data for matching")
)
