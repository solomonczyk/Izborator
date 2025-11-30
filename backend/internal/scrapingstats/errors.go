package scrapingstats

import "errors"

var (
	ErrInvalidShopID = errors.New("invalid shop ID")
	ErrInvalidStatus = errors.New("invalid status (must be success, error, or partial)")
)

