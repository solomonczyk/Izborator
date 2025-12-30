#!/bin/bash

# Диагностический скрипт для проверки почему не ищутся товары

set -e

SERVER="152.53.227.37"
KEY="~/.ssh/izborator_key"

echo "🔍 Diagnosing product search issue..."
echo "Server: $SERVER"
echo ""

# SSH коннект и диагностика
ssh -i "$KEY" root@"$SERVER" << 'EOFSCRIPT'

echo "1️⃣  Checking PostgreSQL..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT COUNT(*) as product_count FROM products;" || echo "❌ PostgreSQL error"

echo ""
echo "2️⃣  Checking Meilisearch..."
curl -s http://meilisearch:7700/indexes/products/stats | jq . || echo "❌ Meilisearch error (try: curl http://localhost:7700/indexes/products/stats)"

echo ""
echo "3️⃣  Checking if API is accessible..."
curl -s http://localhost:8080/api/health | jq . || echo "❌ API error"

echo ""
echo "4️⃣  Testing search endpoint..."
curl -s "http://localhost:8080/api/v1/products/search?q=test" | jq . || echo "❌ Search error"

echo ""
echo "5️⃣  Checking backend logs..."
docker-compose logs --tail=20 backend | grep -i "search\|error\|index" || echo "No relevant logs"

EOFSCRIPT

echo ""
echo "✅ Diagnostics complete!"
