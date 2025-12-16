# Start Izborator dev environment (one-click)
# Double-click this file to start backend and frontend

Write-Host "Starting Izborator..." -ForegroundColor Cyan
Write-Host ""

# Get paths
$rootPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendPath = Join-Path $rootPath "backend"
$frontendPath = Join-Path $rootPath "frontend"

# Check if folders exist
if (-not (Test-Path $backendPath)) {
    Write-Host "ERROR: Backend folder not found!" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

if (-not (Test-Path $frontendPath)) {
    Write-Host "ERROR: Frontend folder not found!" -ForegroundColor Red
    Read-Host "Press Enter to exit"
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
                Write-Host "Stopping process on port $Port..." -ForegroundColor Yellow
                Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
            }
        }
        Start-Sleep -Seconds 1
    }
}

# Stop old processes
Write-Host "Checking ports..." -ForegroundColor Cyan
Stop-PortProcess -Port 8081
Stop-PortProcess -Port 3000
Stop-PortProcess -Port 3001
Stop-PortProcess -Port 3002

# Clear frontend cache
Write-Host "Clearing cache..." -ForegroundColor Cyan
$nextCachePath = Join-Path $frontendPath ".next"
if (Test-Path $nextCachePath) {
    Remove-Item -Recurse -Force $nextCachePath -ErrorAction SilentlyContinue
}

# Set environment variables for backend
$env:DB_HOST = "localhost"
$env:DB_PORT = "5433"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "postgres"
$env:DB_NAME = "izborator"
$env:SERVER_PORT = "8081"

# Start backend in background
Write-Host ""
Write-Host "Starting Backend API (port 8081)..." -ForegroundColor Green
$backendJob = Start-Job -ScriptBlock {
    param($path, $envVars)
    Set-Location $path
        $env:DB_HOST = $envVars.DB_HOST
        $env:DB_PORT = $envVars.DB_PORT
        $env:DB_USER = $envVars.DB_USER
        $env:DB_PASSWORD = $envVars.DB_PASSWORD
        $env:DB_NAME = $envVars.DB_NAME
        $env:SERVER_PORT = $envVars.SERVER_PORT
        go run cmd/api/main.go 2>&1
} -ArgumentList $backendPath, @{
    DB_HOST = "localhost"
    DB_PORT = "5433"
    DB_USER = "postgres"
    DB_PASSWORD = "postgres"
    DB_NAME = "izborator"
    SERVER_PORT = "8081"
}

# Wait a bit
Start-Sleep -Seconds 3

# Start frontend in background
Write-Host "Starting Frontend (Next.js)..." -ForegroundColor Green
$frontendJob = Start-Job -ScriptBlock {
    param($path)
    Set-Location $path
    npm run dev 2>&1
} -ArgumentList $frontendPath

Write-Host ""
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host "Izborator is starting!" -ForegroundColor Green
Write-Host ""
Write-Host "Backend API:  http://localhost:8081" -ForegroundColor Yellow
Write-Host "Frontend:     http://localhost:3000 (or 3001, 3002)" -ForegroundColor Yellow
Write-Host ""
Write-Host "Waiting for services to start..." -ForegroundColor Gray
Write-Host "Press Ctrl+C to stop all services" -ForegroundColor Gray
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host ""

# Monitor jobs and show output
try {
    while ($true) {
        # Show backend output
        $backendOutput = Receive-Job -Job $backendJob -ErrorAction SilentlyContinue
        if ($backendOutput) {
            Write-Host "[Backend] $backendOutput" -ForegroundColor Blue
        }
        
        # Show frontend output
        $frontendOutput = Receive-Job -Job $frontendJob -ErrorAction SilentlyContinue
        if ($frontendOutput) {
            Write-Host "[Frontend] $frontendOutput" -ForegroundColor Magenta
        }
        
        Start-Sleep -Milliseconds 500
    }
} finally {
    Write-Host ""
    Write-Host "Stopping services..." -ForegroundColor Yellow
    Stop-Job -Job $backendJob, $frontendJob -ErrorAction SilentlyContinue
    Remove-Job -Job $backendJob, $frontendJob -Force -ErrorAction SilentlyContinue
    Write-Host "Done!" -ForegroundColor Green
}





