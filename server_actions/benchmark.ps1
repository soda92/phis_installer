$ErrorActionPreference = 'Stop'
$installerPath = "C:\installer_test\installer.exe"
$installDir = "C:\wu-xian-shi-xun"
$uninstallPath = "$installDir\uninstall.exe"

# 1. Uninstall if exists
if (Test-Path $uninstallPath) {
    Write-Host "Found existing installation at $installDir. Uninstalling..."
    $p = Start-Process -FilePath $uninstallPath -ArgumentList '/S' -PassThru -Wait
    Write-Host "Uninstall complete. Exit Code: $($p.ExitCode)"
    
    # Wait a bit for file release and cleanup
    Start-Sleep -Seconds 2
    
    if (Test-Path $installDir) {
        Write-Host "Cleaning up remaining files in $installDir..."
        Remove-Item -Path $installDir -Recurse -Force -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "No existing installation found."
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
