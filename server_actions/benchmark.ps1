$ErrorActionPreference = 'Stop'
$installerPath = "C:\installer_test\installer.exe"
$installDir = "C:\wu-xian-shi-xun"
$uninstallPath = "$installDir\uninstall.exe"

# Pre-cleanup: Kill any potential lingering processes that might lock files
Get-Process -Name "python", "pip", "installer", "uninstall", "数字员工平台*" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

# 1. Uninstall if exists
if (Test-Path $uninstallPath) {
    Write-Host "Found existing installation at $installDir. Uninstalling..."
    
    # Start uninstaller with _?=$installDir
    $p = Start-Process -FilePath $uninstallPath -ArgumentList "/S", "_?=$installDir" -PassThru
    
    # Wait up to 60 seconds for it to finish
    if ($p.WaitForExit(60000)) {
        Write-Host "Uninstall process exited normally. Code: $($p.ExitCode)"
    } else {
        Write-Warning "Uninstall process timed out. Killing..."
        $p | Stop-Process -Force -ErrorAction SilentlyContinue
    }
    
    # Wait a bit for file handles to release
    Start-Sleep -Seconds 2
    
    # Manually clean up
    if (Test-Path $installDir) {
        Write-Host "Forcing cleanup of installation directory..."
        for ($i=0; $i -lt 3; $i++) {
            try {
                Remove-Item -Path $installDir -Recurse -Force -ErrorAction Stop
                break
            } catch {
                Write-Host "Cleanup attempt $($i+1) failed ($($_.Exception.Message)). Retrying..."
                Start-Sleep -Seconds 2
            }
        }
    }
} else {
    Write-Host "No existing installation found."
}

# Double check cleanup
if (Test-Path $installDir) {
    Write-Warning "Failed to fully clean $installDir. Installation might fail."
}

# 2. Install and Measure
if (-not (Test-Path $installerPath)) {
    Write-Error "Installer not found at $installerPath"
}

Write-Host "Starting installation of $installerPath..."
$time = Measure-Command {
    $p = Start-Process -FilePath $installerPath -ArgumentList '/S' -PassThru -Wait
    if ($p.ExitCode -ne 0) {
        Write-Error "Installation failed with exit code $($p.ExitCode)"
    }
}

$minutes = [math]::Round($time.TotalMinutes, 2)
Write-Host "---------------------------------------------------"
Write-Host "Installation Time: $($time.TotalSeconds) seconds"
Write-Host "Installation Time: $minutes minutes"
Write-Host "---------------------------------------------------"
