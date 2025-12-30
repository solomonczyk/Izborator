#!/bin/bash

# Скрипт для переиндексирования товаров в Meilisearch на production сервере

set -e

echo "🔄 Re-indexing products in Meilisearch..."

# SSH на сервер и выполняем переиндексирование
ssh -i ~/.ssh/izborator_key root@152.53.227.37 << 'EOF'

cd /root/Izborator

# Остановим worker если запущен
docker-compose exec -T backend pkill -f "go run cmd/worker" || true

# Запустим переиндексирование
echo "📋 Running indexer command..."
docker-compose exec -T backend go run cmd/indexer/main.go

echo "✅ Re-indexing completed!"
echo ""
echo "Testing search..."
curl -s http://localhost:8080/api/v1/products/search?q=тест | jq . || echo "(No results or API not available)"

EOF

echo "✅ Done!"
