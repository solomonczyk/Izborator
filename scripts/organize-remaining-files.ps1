# Скрипт для перемещения оставшихся файлов в соответствующие директории

Write-Host "📁 Организация оставшихся файлов проекта..." -ForegroundColor Cyan
Write-Host ""

# Перемещение документации в docs/
Write-Host "📄 Перемещение документации..." -ForegroundColor Yellow

# Архитектурная документация
$archDocs = @(
    "PROJECT_STRUCTURE.md",
    "PROJECT_RULES.md",
    "ARCHITECTURE_RULES.md"
)

foreach ($doc in $archDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/architecture/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $doc → docs/architecture/" -ForegroundColor Gray
    }
}

# Документация разработки
$devDocs = @(
    "IMPROVEMENTS.md",
    "FIXES_REPORT.md",
    "SEED_DATA_RESULTS.md",
    "TEST_DRIVE.md",
    "TEST_SERVER_API.md",
    "ADD_REAL_URLS.md",
    "MULTI_SHOP_CATALOG_SETUP.md",
    "ROADMAP_CURRENT_STEP.md"
)

foreach ($doc in $devDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/development/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $doc → docs/development/" -ForegroundColor Gray
    }
}

# Документация деплоя
$deployDocs = @(
    "docker-compose.README.md",
    "START_COMMANDS.md",
    "STATUS.md"
)

foreach ($doc in $deployDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/deployment/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $doc → docs/deployment/" -ForegroundColor Gray
    }
}

# Гайды и стратегия
$guides = @(
    "PROJECT_HORIZON.md",
    "STRATEGY.md",
    "PLAN.md",
    "SUMMARY.md"
)

foreach ($guide in $guides) {
    if (Test-Path $guide) {
        Move-Item -Path $guide -Destination "docs/guides/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $guide → docs/guides/" -ForegroundColor Gray
    }
}

# Безопасность
$securityDocs = @(
    "SECURITY_CLEANUP.md",
    "SECURITY_CLEANUP_OPENAI.md",
    "SECURITY_FIXES.md",
    "SECURITY_GUIDELINES.md",
    "GIT_SECRETS_AUDIT.md"
)

foreach ($doc in $securityDocs) {
    if (Test-Path $doc) {
        Move-Item -Path $doc -Destination "docs/deployment/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $doc → docs/deployment/" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "🔧 Перемещение скриптов..." -ForegroundColor Yellow

# Скрипты для БД
$dbScripts = @(
    "check-migration-status.sh",
    "fix-dirty-migration.sh",
    "fix-shop-config-attempts-table.sh",
    "create-shop-config-attempts-fixed.sql"
)

foreach ($script in $dbScripts) {
    if (Test-Path $script) {
        Move-Item -Path $script -Destination "scripts/database/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $script → scripts/database/" -ForegroundColor Gray
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
        Move-Item -Path $script -Destination "scripts/deployment/" -Force -ErrorAction SilentlyContinue
        Write-Host "  ✅ $script → scripts/deployment/" -ForegroundColor Gray
    }
}

# Тестовые скрипты
Get-ChildItem -Path . -Filter "test-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/testing/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/testing/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "test-*.ps1" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/testing/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/testing/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "check-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/testing/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/testing/" -ForegroundColor Gray
}

# Скрипты обслуживания
Get-ChildItem -Path . -Filter "run-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "do-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "fix-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "rebuild-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "update-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "remove-*.sh" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "clean-*.bat" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "remove-*.bat" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "check-and-remove-*.bat" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Get-ChildItem -Path . -Filter "*.ps1" | Where-Object { $_.Name -ne "organize-project.ps1" -and $_.Name -ne "organize-remaining-files.ps1" } | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

# Python скрипты
Get-ChildItem -Path . -Filter "*.py" | ForEach-Object {
    Move-Item -Path $_.FullName -Destination "scripts/maintenance/" -Force -ErrorAction SilentlyContinue
    Write-Host "  ✅ $($_.Name) → scripts/maintenance/" -ForegroundColor Gray
}

Write-Host ""
Write-Host "✅ Организация завершена!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Оставшиеся файлы в корне:" -ForegroundColor Cyan
Get-ChildItem -Path . -File | Where-Object { $_.Extension -in @(".md", ".sh", ".bat", ".ps1", ".py") } | Select-Object Name | Format-Table -AutoSize

