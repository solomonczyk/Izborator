# Start dev environment for Izborator
# Usage: .\start-dev.ps1

Write-Host "Starting Izborator dev environment" -ForegroundColor Cyan
Write-Host ""

# Get paths
$rootPath = $PSScriptRoot
$backendPath = Join-Path $rootPath "backend"
$frontendPath = Join-Path $rootPath "frontend"

# Check if folders exist
if (-not (Test-Path $backendPath)) {
    Write-Host "ERROR: Backend folder not found: $backendPath" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $frontendPath)) {
    Write-Host "ERROR: Frontend folder not found: $frontendPath" -ForegroundColor Red
    exit 1
}

# Function to stop processes on ports
function Stop-PortProcess {
    param([int]$Port)
    
    $connections = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
    if ($connections) {
        foreach ($conn in $connections) {
            $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
            if ($proc) {
                Write-Host "Stopping process on port $Port (PID: $($proc.Id), Name: $($proc.ProcessName))" -ForegroundColor Yellow
                Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
            }
        }
        Start-Sleep -Seconds 1
    }
}

# Stop old processes
Write-Host "Checking running processes..." -ForegroundColor Cyan
Stop-PortProcess -Port 8081
Stop-PortProcess -Port 3000
Stop-PortProcess -Port 3001
Stop-PortProcess -Port 3002

Write-Host ""

# Clear frontend cache
Write-Host "Clearing frontend cache..." -ForegroundColor Cyan
$nextCachePath = Join-Path $frontendPath ".next"
if (Test-Path $nextCachePath) {
    Remove-Item -Recurse -Force $nextCachePath
    Write-Host "Cache .next cleared" -ForegroundColor Green
} else {
    Write-Host "Cache .next not found" -ForegroundColor Gray
}

Write-Host ""

# Start backend
Write-Host "Starting backend API (port 8081)..." -ForegroundColor Cyan
$backendScript = @"
cd '$backendPath'
`$env:DB_HOST = 'localhost'
`$env:DB_PORT = '5433'
`$env:DB_USER = 'postgres'
`$env:DB_PASSWORD = 'postgres'
`$env:DB_NAME = 'izborator'
`$env:SERVER_PORT = '8081'
go run cmd/api/main.go
"@

Start-Process powershell -ArgumentList "-NoExit", "-Command", $backendScript -WindowStyle Normal
Write-Host "Backend started in new PowerShell window" -ForegroundColor Green

# Wait a bit for backend to start
Start-Sleep -Seconds 3

# Start frontend
Write-Host ""
Write-Host "Starting frontend (Next.js)..." -ForegroundColor Cyan
$frontendScript = @"
cd '$frontendPath'
npm run dev
"@

Start-Process powershell -ArgumentList "-NoExit", "-Command", $frontendScript -WindowStyle Normal
Write-Host "Frontend started in new PowerShell window" -ForegroundColor Green

Write-Host ""
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host "Dev environment started!" -ForegroundColor Green
Write-Host ""
Write-Host "Backend API:  http://localhost:8081" -ForegroundColor Yellow
Write-Host "Frontend:     http://localhost:3000 (or 3001, 3002)" -ForegroundColor Yellow
Write-Host ""
Write-Host "Check PowerShell windows for logs" -ForegroundColor Gray
Write-Host "To stop: close PowerShell windows or press Ctrl+C" -ForegroundColor Gray
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host ""
