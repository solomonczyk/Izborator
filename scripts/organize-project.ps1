# Скрипт для организации структуры проекта
# Перемещает файлы в соответствующие директории

Write-Host "📁 Организация структуры проекта..." -ForegroundColor Cyan
Write-Host ""

# Создаем структуру директорий
$directories = @(
    "docs/architecture",
    "docs/development", 
    "docs/deployment",
    "docs/guides",
    "scripts/database",
    "scripts/deployment",
    "scripts/testing",
    "scripts/maintenance"
)

foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        Write-Host "✅ Создана директория: $dir" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "📄 Перемещение документации..." -ForegroundColor Yellow

# Архитектурная документация
$archDocs = @(
    "ARCHITECTURE_DATA_FLOW.md",
    "ARCHITECTURE_RULES.md",
    "PROJECT_STRUCTURE.md",
    "MODULE_ARCHITECTURE.md"
)

foreach ($doc in $archDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/architecture/" -Force
        Write-Host "  ✅ Перемещен: $doc → docs/architecture/" -ForegroundColor Gray
    }
}

# Документация разработки
$devDocs = @(
    "DEVELOPMENT_LOG.md",
    "DEVELOPMENT_SETUP.md",
    "DEVELOPMENT_FLOW.md",
    "TESTING_GUIDE.md",
    "E2E_TESTING_GUIDE.md",
    "E2E_TEST_CHECKLIST.md",
    "E2E_TEST_RESULTS.md"
)

foreach ($doc in $devDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/development/" -Force
        Write-Host "  ✅ Перемещен: $doc → docs/development/" -ForegroundColor Gray
    }
}

# Документация деплоя
$deployDocs = @(
    "DEPLOY.md",
    "DEPLOY_SERVER.md",
    "DEPLOY_FIX.md",
    "CICD_SETUP.md",
    "CICD_TROUBLESHOOTING.md",
    "CI_CD_STATUS.md",
    "QUICK_CI_SETUP.md",
    "NGINX_SETUP.md",
    "NGINX_PROXY_SETUP.md",
    "HTTPS_SETUP.md",
    "VERIFY_HTTPS.md"
)

foreach ($doc in $deployDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/deployment/" -Force
        Write-Host "  ✅ Перемещен: $doc → docs/deployment/" -ForegroundColor Gray
    }
}

# Гайды
$guides = @(
    "AUTOCONFIG_RUN.md",
    "AUTOCONFIG_CHECK_STATUS.md",
    "CHECK_AUTOCONFIG.md",
    "CLASSIFIER_RUN.md",
    "DEBUG_CLASSIFIER.md",
    "DISCOVERY_SETUP.md",
    "CATALOG_PARSER_SETUP.md",
    "PARSE_INSTRUCTIONS.md",
    "QUICK_PARSE_GUIDE.md",
    "QUICK_FIX.md",
    "QUICK_E2E_TEST.md",
    "HARVEST.md",
    "FIX_HARVEST.md",
    "WORKER_CHECK.md",
    "TEST_API_SERVER.md",
    "RUN_API_AND_TEST.md"
)

foreach ($guide in $guides) {
    if (Test-Path $guide) {
        Move-Item -Path $guide -Destination "docs/guides/" -Force
        Write-Host "  ✅ Перемещен: $guide → docs/guides/" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "🔧 Перемещение скриптов..." -ForegroundColor Yellow

# Скрипты для БД
$dbScripts = @(
    "check-migration-status.sh",
    "fix-dirty-migration.sh",
    "fix-shop-config-attempts-table.sh"
)

foreach ($script in $dbScripts) {
    if (Test-Path $script) {
        Move-Item -Path $script -Destination "scripts/database/" -Force
        Write-Host "  ✅ Перемещен: $script → scripts/database/" -ForegroundColor Gray
    }
}

# Скрипты деплоя
$deployScripts = @(
    "deploy.sh",
    "fix-on-server.sh",
    "run-fix.sh"
)

foreach ($script in $deployScripts) {
    if (Test-Path $script) {
        Move-Item -Path $script -Destination "scripts/deployment/" -Force
        Write-Host "  ✅ Перемещен: $script → scripts/deployment/" -ForegroundColor Gray
    }
}

# Тестовые скрипты
$testScripts = @(
    "test-*.sh",
    "test-*.ps1",
    "check-*.sh",
    "quick-test-api.sh"
)

Get-ChildItem -Path . -Filter "test-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/testing/" -Force
    Write-Host "  ✅ Перемещен: $($_.Name) → scripts/testing/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "test-*.ps1" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/testing/" -Force
    Write-Host "  ✅ Перемещен: $($_.Name) → scripts/testing/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "check-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/testing/" -Force
    Write-Host "  ✅ Перемещен: $($_.Name) → scripts/testing/" -ForegroundColor Gray
}

# Скрипты обслуживания
$maintenanceScripts = @(
    "run-*.sh",
    "do-*.sh",
    "fix-and-run*.sh",
    "rebuild-*.sh",
    "update-*.sh",
    "remove-*.sh",
    "clean-*.sh",
    "check-and-remove-*.bat"
)

Get-ChildItem -Path . -Filter "run-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force
    Write-Host "  ✅ Перемещен: $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "do-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force
    Write-Host "  ✅ Перемещен: $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Write-Host ""
Write-Host "✅ Организация завершена!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Следующие шаги:" -ForegroundColor Cyan
Write-Host "  1. Проверьте перемещенные файлы"
Write-Host "  2. Обновите ссылки в документации"
Write-Host "  3. Обновите .gitignore при необходимости"
Write-Host "  4. Закоммитьте изменения"

