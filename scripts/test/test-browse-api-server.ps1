# Скрипт для тестирования /browse API endpoints на сервере (PowerShell)

# Определяем URL API
# Если скрипт запускается локально, используем внешний адрес сервера
$API_BASE = if ($env:API_BASE) { $env:API_BASE } else { "https://152.53.227.37" }

# Если переменная SERVER_IP задана, используем её
if ($env:SERVER_IP) {
    $API_BASE = "https://$env:SERVER_IP"
}

Write-Host "🔍 Тестирование /browse API endpoints на сервере" -ForegroundColor Cyan
Write-Host "API Base: $API_BASE"
Write-Host ""

# Функция для тестирования endpoint
function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Url,
        [int]$ExpectedStatus = 200
    )
    
    Write-Host -NoNewline "Тест: $Name... "
    
    try {
        # Игнорируем SSL ошибки для самоподписанных сертификатов
        # Для старых версий PowerShell используем [System.Net.ServicePointManager]::ServerCertificateValidationCallback
        if ($PSVersionTable.PSVersion.Major -lt 6) {
            [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
        }
        $response = if ($PSVersionTable.PSVersion.Major -ge 6) {
            Invoke-WebRequest -Uri $Url -Method GET -UseBasicParsing -SkipCertificateCheck -ErrorAction Stop
        } else {
            Invoke-WebRequest -Uri $Url -Method GET -UseBasicParsing -ErrorAction Stop
        }
        $httpCode = $response.StatusCode
        $body = $response.Content
        
        if ($httpCode -eq $ExpectedStatus) {
            Write-Host "✅ OK" -ForegroundColor Green -NoNewline
            Write-Host " (HTTP $httpCode)"
            
            # Парсим JSON
            try {
                $json = $body | ConvertFrom-Json
                
                $itemsCount = if ($json.items) { $json.items.Count } else { 0 }
                $total = if ($json.total) { $json.total } else { 0 }
                $page = if ($json.page) { $json.page } else { 0 }
                $perPage = if ($json.per_page) { $json.per_page } else { 0 }
                
                Write-Host "   📊 Результаты: items=$itemsCount, total=$total, page=$page, per_page=$perPage" -ForegroundColor Gray
                
                # Показываем первый товар, если есть
                if ($itemsCount -gt 0 -and $json.items[0]) {
                    $firstItem = $json.items[0] | Select-Object -Property id, name, category_id, shops_count
                    Write-Host "   📦 Первый товар: $($firstItem | ConvertTo-Json -Compress)" -ForegroundColor Gray
                }
            } catch {
                Write-Host "   ⚠️  Ответ не является валидным JSON" -ForegroundColor Yellow
                Write-Host "   Ответ: $($body.Substring(0, [Math]::Min(200, $body.Length)))..." -ForegroundColor Gray
            }
        } else {
            Write-Host "❌ FAILED" -ForegroundColor Red -NoNewline
            Write-Host " (HTTP $httpCode, ожидался $ExpectedStatus)"
            Write-Host "   Ответ: $($body.Substring(0, [Math]::Min(200, $body.Length)))..." -ForegroundColor Gray
            return $false
        }
    } catch {
        Write-Host "❌ ERROR" -ForegroundColor Red
        Write-Host "   Ошибка: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    Write-Host ""
    return $true
}

# Тест 1: Health check
Write-Host "🔍 Проверка здоровья API..." -ForegroundColor Cyan
try {
    if ($PSVersionTable.PSVersion.Major -lt 6) {
        [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
    }
    $healthResponse = if ($PSVersionTable.PSVersion.Major -ge 6) {
        Invoke-WebRequest -Uri "$API_BASE/api/health" -Method GET -UseBasicParsing -SkipCertificateCheck -ErrorAction Stop
    } else {
        Invoke-WebRequest -Uri "$API_BASE/api/health" -Method GET -UseBasicParsing -ErrorAction Stop
    }
    if ($healthResponse.Content -match "ok|status") {
        Write-Host "✅ API работает" -ForegroundColor Green
    } else {
        Write-Host "⚠️  API отвечает, но ответ неожиданный" -ForegroundColor Yellow
        Write-Host "   Ответ: $($healthResponse.Content)" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ API не отвечает" -ForegroundColor Red
    Write-Host "   Ошибка: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Тест 2: Browse без фильтров
Test-Endpoint `
    -Name "GET /api/v1/products/browse (без фильтра)" `
    -Url "$API_BASE/api/v1/products/browse?page=1&per_page=5"

# Тест 3: Browse с категорией mobilni-telefoni
Test-Endpoint `
    -Name "GET /api/v1/products/browse?category=mobilni-telefoni" `
    -Url "$API_BASE/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5"

# Тест 4: Browse с категорией laptopovi
Test-Endpoint `
    -Name "GET /api/v1/products/browse?category=laptopovi" `
    -Url "$API_BASE/api/v1/products/browse?category=laptopovi&page=1&per_page=5"

# Тест 5: Browse с несуществующей категорией (fallback)
Test-Endpoint `
    -Name "GET /api/v1/products/browse?category=neexistujuca-kategorija (fallback)" `
    -Url "$API_BASE/api/v1/products/browse?category=neexistujuca-kategorija&page=1&per_page=5" `
    -ExpectedStatus 200

# Тест 6: Проверка структуры BrowseResult
Write-Host "🔍 Проверка структуры BrowseResult..." -ForegroundColor Cyan
try {
    if ($PSVersionTable.PSVersion.Major -lt 6) {
        [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
    }
    $response = if ($PSVersionTable.PSVersion.Major -ge 6) {
        Invoke-WebRequest -Uri "$API_BASE/api/v1/products/browse?page=1&per_page=1" -Method GET -UseBasicParsing -SkipCertificateCheck
    } else {
        Invoke-WebRequest -Uri "$API_BASE/api/v1/products/browse?page=1&per_page=1" -Method GET -UseBasicParsing
    }
    $json = $response.Content | ConvertFrom-Json
    
    $hasItems = $json.PSObject.Properties.Name -contains "items"
    $hasTotal = $json.PSObject.Properties.Name -contains "total"
    $hasPage = $json.PSObject.Properties.Name -contains "page"
    $hasPerPage = $json.PSObject.Properties.Name -contains "per_page"
    
    if ($hasItems -and $hasTotal -and $hasPage -and $hasPerPage) {
        Write-Host "✅ Структура BrowseResult корректна" -ForegroundColor Green
        Write-Host "   Поля: items, total, page, per_page" -ForegroundColor Gray
    } else {
        Write-Host "❌ Структура BrowseResult некорректна" -ForegroundColor Red
        Write-Host "   Найдены поля: $($json.PSObject.Properties.Name -join ', ')" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ Ошибка при проверке структуры: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "✅ Все тесты завершены!" -ForegroundColor Green

