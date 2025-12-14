@echo off
REM Start Indexer with correct environment variables
cd /d %~dp0
set DB_HOST=localhost
set DB_PORT=5433
set DB_USER=postgres
set DB_PASSWORD=postgres
set DB_NAME=izborator
set MEILISEARCH_API_KEY=masterKey123
set MEILISEARCH_HOST=localhost
set MEILISEARCH_PORT=7700

if "%1"=="setup" (
    go run cmd/indexer/main.go -setup
) else if "%1"=="sync" (
    go run cmd/indexer/main.go -sync
) else if "%1"=="reindex" (
    go run cmd/indexer/main.go -reindex
) else (
    echo Usage: start-indexer.bat [setup^|sync^|reindex]
    echo   setup   - Setup Meilisearch index
    echo   sync    - Sync products from PostgreSQL to Meilisearch
    echo   reindex - Reindex all products (clear and rebuild)
)

pause

