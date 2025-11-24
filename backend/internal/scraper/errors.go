package scraper

import "errors"

var (
	// ErrShopNotFound магазин не найден
	ErrShopNotFound = errors.New("shop not found")
	
	// ErrScrapingFailed ошибка при парсинге
	ErrScrapingFailed = errors.New("scraping failed")
	
	// ErrInvalidURL невалидный URL
	ErrInvalidURL = errors.New("invalid URL")
	
	// ErrRateLimitExceeded превышен лимит запросов
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	
	// ErrShopDisabled магазин отключен
	ErrShopDisabled = errors.New("shop is disabled")
)

