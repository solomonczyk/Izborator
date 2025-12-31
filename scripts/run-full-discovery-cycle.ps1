# Полный цикл Discovery: Discovery -> Classifier -> AutoConfig
# Использование: .\scripts\run-full-discovery-cycle.ps1 [-LimitAutoConfig <number>]

param(
    [int]$LimitAutoConfig = 5
)

Write-Host "[START] Запуск полного цикла Discovery" -ForegroundColor Cyan
Write-Host ""

# Переходим в директорию backend
$backendPath = Join-Path $PSScriptRoot ".." "backend"
Push-Location $backendPath

try {
    # Шаг 1: Discovery
    Write-Host "[STEP 1] Запуск Discovery (поиск кандидатов)..." -ForegroundColor Yellow
    Write-Host ""
    
    if (Test-Path "discovery.exe") {
        & .\discovery.exe
    } elseif (Test-Path "discovery") {
        & .\discovery
    } else {
        Write-Host "[ERROR] discovery не найден!" -ForegroundColor Red
        exit 1
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARNING] Discovery завершился с предупреждениями (код: $LASTEXITCODE)" -ForegroundColor Yellow
    } else {
        Write-Host "[OK] Discovery завершен успешно" -ForegroundColor Green
    }
    Write-Host ""
    
    # Шаг 2: Classifier
    Write-Host "[STEP 2] Запуск Classifier (классификация)..." -ForegroundColor Yellow
    Write-Host ""
    
    if (Test-Path "classifier.exe") {
        & .\classifier.exe -classify-all
    } elseif (Test-Path "classifier") {
        & .\classifier -classify-all
    } else {
        Write-Host "[ERROR] classifier не найден!" -ForegroundColor Red
        exit 1
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARNING] Classifier завершился с предупреждениями (код: $LASTEXITCODE)" -ForegroundColor Yellow
    } else {
        Write-Host "[OK] Classifier завершен успешно" -ForegroundColor Green
    }
    Write-Host ""
    
    # Шаг 3: AutoConfig
    Write-Host "[STEP 3] Запуск AutoConfig (AI генерация селекторов)..." -ForegroundColor Yellow
    Write-Host "Обрабатываем $LimitAutoConfig кандидатов..." -ForegroundColor Gray
    Write-Host ""
    
    if (Test-Path "autoconfig.exe") {
        & .\autoconfig.exe -limit $LimitAutoConfig
    } elseif (Test-Path "autoconfig") {
        & .\autoconfig -limit $LimitAutoConfig
    } else {
        Write-Host "[ERROR] autoconfig не найден!" -ForegroundColor Red
        exit 1
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARNING] AutoConfig завершился с предупреждениями (код: $LASTEXITCODE)" -ForegroundColor Yellow
    } else {
        Write-Host "[OK] AutoConfig завершен успешно" -ForegroundColor Green
    }
    Write-Host ""
    
    Write-Host "[COMPLETE] Полный цикл Discovery завершен!" -ForegroundColor Green
    Write-Host ""
    
} finally {
    Pop-Location
}

