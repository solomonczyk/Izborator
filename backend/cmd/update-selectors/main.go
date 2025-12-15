package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5433/izborator?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	sql := `
		UPDATE shops
		SET selectors = '{
			"name": "h1, .page-title, .page-title-wrapper h1, .product-info-main h1, .product-name, [data-ui-id=\"page-title-wrapper\"]",
			"price": ".price, .price-wrapper .price, [data-price-type], .product-info-price .price, .price-final, span.price",
			"image": "img.product-image, .product-image-gallery img, .gallery-image img, .product-media img, .fotorama__img, .product.media img",
			"description": ".product-description, .product-info-description, .product.attribute.description, [data-ui-id=\"page-title-wrapper\"] + *",
			"brand": ".product-brand, .product-attribute-brand, [data-attribute-code=\"brand\"], .brand",
			"category": ".breadcrumbs, .category-path, nav.breadcrumbs"
		}'::jsonb
		WHERE code = 'tehnomanija' OR id = 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22';
	`

	_, err = pool.Exec(context.Background(), sql)
	if err != nil {
		log.Fatalf("Failed to update selectors: %v", err)
	}

	log.Println("✅ Selectors updated successfully!")
}

