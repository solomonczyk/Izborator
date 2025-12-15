-- Обновление селекторов для Tehnomanija
-- На основе анализа структуры страницы (Magento-based)
-- Используем более универсальные селекторы

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

