package pricehistory

import "errors"

var (
	// ErrPriceNotFound цена не найдена
	ErrPriceNotFound = errors.New("price not found")

	// ErrInvalidPeriod невалидный период
	ErrInvalidPeriod = errors.New("invalid period")

	// ErrInvalidTimeRange невалидный временной диапазон
	ErrInvalidTimeRange = errors.New("invalid time range")
)
