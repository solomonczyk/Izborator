# Stop dev environment for Izborator
# Usage: .\stop-dev.ps1

Write-Host "Stopping Izborator dev environment" -ForegroundColor Yellow
Write-Host ""

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
        return $true
    }
    return $false
}

# Stop processes
$stopped = $false

if (Stop-PortProcess -Port 8081) { $stopped = $true }
if (Stop-PortProcess -Port 3000) { $stopped = $true }
if (Stop-PortProcess -Port 3001) { $stopped = $true }
if (Stop-PortProcess -Port 3002) { $stopped = $true }

if ($stopped) {
    Write-Host ""
    Write-Host "All processes stopped" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "No processes found" -ForegroundColor Gray
}

Write-Host ""
