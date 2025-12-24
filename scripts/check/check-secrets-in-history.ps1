# Скрипт для проверки секретов в истории Git
# Запуск: .\check-secrets-in-history.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "ПРОВЕРКА СЕКРЕТОВ В ИСТОРИИ GIT" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Проверка masterKey123
Write-Host "[1/6] Проверка masterKey123..." -ForegroundColor Yellow
$masterKeyCommits = git log --all -S "masterKey123" --source --full-history --oneline
if ($masterKeyCommits) {
    Write-Host "  ⚠️  НАЙДЕНО в коммитах:" -ForegroundColor Red
    $masterKeyCommits | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
} else {
    Write-Host "  ✅ Не найдено" -ForegroundColor Green
}
Write-Host ""

# 2. Проверка OpenAI API ключей (sk-proj-)
Write-Host "[2/6] Проверка OpenAI API ключей (sk-proj-)..." -ForegroundColor Yellow
$openaiCommits = git log --all -S "sk-proj-" --source --full-history --oneline
if ($openaiCommits) {
    Write-Host "  ⚠️  НАЙДЕНО в коммитах:" -ForegroundColor Red
    $openaiCommits | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
} else {
    Write-Host "  ✅ Не найдено" -ForegroundColor Green
}
Write-Host ""

# 3. Проверка Google API ключей (AIzaSy)
Write-Host "[3/6] Проверка Google API ключей (AIzaSy)..." -ForegroundColor Yellow
$googleCommits = git log --all -S "AIzaSy" --source --full-history --oneline
if ($googleCommits) {
    Write-Host "  ⚠️  НАЙДЕНО в коммитах:" -ForegroundColor Red
    $googleCommits | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
} else {
    Write-Host "  ✅ Не найдено" -ForegroundColor Green
}
Write-Host ""

# 4. Проверка adminpassword
Write-Host "[4/6] Проверка adminpassword..." -ForegroundColor Yellow
$adminPassCommits = git log --all -S "adminpassword" --source --full-history --oneline
if ($adminPassCommits) {
    Write-Host "  ⚠️  НАЙДЕНО в коммитах:" -ForegroundColor Red
    $adminPassCommits | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
} else {
    Write-Host "  ✅ Не найдено" -ForegroundColor Green
}
Write-Host ""

# 5. Проверка .env файлов в истории
Write-Host "[5/6] Проверка .env файлов в истории..." -ForegroundColor Yellow
$envFiles = git log --all --diff-filter=A --source --full-history --name-only -- .env backend/.env frontend/.env 2>&1 | Where-Object { $_ -match "\.env" } | Select-Object -Unique
if ($envFiles) {
    Write-Host "  ⚠️  НАЙДЕНО .env файлов в истории:" -ForegroundColor Red
    $envFiles | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
} else {
    Write-Host "  ✅ .env файлы не найдены в истории" -ForegroundColor Green
}
Write-Host ""

# 6. Проверка .env файлов в текущем индексе
Write-Host "[6/6] Проверка .env файлов в текущем индексе..." -ForegroundColor Yellow
$trackedEnvFiles = git ls-files | Where-Object { $_ -match "\.env$" -and $_ -notmatch "\.env\.example" }
if ($trackedEnvFiles) {
    Write-Host "  ⚠️  НАЙДЕНО .env файлов в индексе:" -ForegroundColor Red
    $trackedEnvFiles | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
} else {
    Write-Host "  ✅ .env файлы не отслеживаются Git" -ForegroundColor Green
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "РЕЗУЛЬТАТЫ ПРОВЕРКИ" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$hasIssues = $false
if ($masterKeyCommits -or $openaiCommits -or $googleCommits -or $adminPassCommits -or $envFiles -or $trackedEnvFiles) {
    $hasIssues = $true
    Write-Host "⚠️  ОБНАРУЖЕНЫ СЕКРЕТЫ В ИСТОРИИ GIT!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Следующие шаги:" -ForegroundColor Yellow
    Write-Host "1. Запустить очистку истории: clean-secrets-history.bat" -ForegroundColor Yellow
    Write-Host "2. После очистки - принудительный push: git push --force --all" -ForegroundColor Yellow
    Write-Host "3. Ротировать все API ключи (OpenAI, Google)" -ForegroundColor Yellow
    Write-Host "4. Обновить .env файлы на сервере" -ForegroundColor Yellow
} else {
    Write-Host "✅ Секреты в истории Git не обнаружены" -ForegroundColor Green
    Write-Host ""
    Write-Host "Рекомендации:" -ForegroundColor Yellow
    Write-Host "- Продолжать следить за тем, чтобы .env файлы не попадали в Git" -ForegroundColor Yellow
    Write-Host "- Использовать .env.example для шаблонов" -ForegroundColor Yellow
}

Write-Host ""

