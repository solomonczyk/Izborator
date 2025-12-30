#!/bin/bash

# Скрипт для загрузки тестовых товаров на production сервер

set -e

SERVER="152.53.227.37"
KEY="~/.ssh/izborator_key"

echo "📦 Loading test products to production database..."
echo "Server: $SERVER"
echo ""

# SSH коннект и загрузка данных
ssh -i "$KEY" root@"$SERVER" << 'EOFSCRIPT'

cd /root/Izborator

echo "1️⃣  Loading test products..."
docker-compose exec -T postgres psql -U postgres -d izborator < /root/Izborator/backend/scripts/seed_test_products.sql

if [ $? -eq 0 ]; then
    echo "✅ Test products loaded successfully!"
else
    echo "⚠️ Failed to load test products"
    exit 1
fi

echo ""
echo "2️⃣  Verifying products in database..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT COUNT(*) as product_count FROM products;"

echo ""
echo "3️⃣  Re-indexing products in Meilisearch..."
cd /root/Izborator/backend
go run cmd/indexer/main.go -reindex

if [ $? -eq 0 ]; then
    echo "✅ Re-indexing completed!"
else
    echo "⚠️ Re-indexing failed"
    exit 1
fi

echo ""
echo "4️⃣  Checking Meilisearch index..."
curl -s http://meilisearch:7700/indexes/products/stats | jq .

echo ""
echo "5️⃣  Testing search..."
curl -s "http://backend:8080/api/v1/products/search?q=Samsung" | jq .

EOFSCRIPT

echo ""
echo "✅ Done! Test products should now be searchable."
echo "Try searching for: 'Samsung', 'Nike', 'Lenovo', 'Televisor'"
