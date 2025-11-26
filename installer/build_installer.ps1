# This script automates the build process for the 'check_only' NSIS installer.

# Strict mode and error handling
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

try {
    # Ensure the script runs from its own directory to resolve relative paths correctly
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    Set-Location $scriptDir
    Write-Host "当前工作目录: $(Get-Location)"

    # 1. Run prerequisite scripts to download dependencies
    Write-Host "--- 正在下载 pip, setuptools, wheel ---"
    & .\download_pip.ps1
    Write-Host "--- pip 下载完成 ---"

    Write-Host "--- 正在下载依赖包 ---"
    & .\download_packages.ps1
    Write-Host "--- 依赖包下载完成 ---"

    # 2. Find NSIS installation directory, handling scoop shims
    Write-Host "--- 正在查找 NSIS 安装目录 ---"
    $makensisCommand = Get-Command makensis.exe -ErrorAction SilentlyContinue
    if (-not $makensisCommand) {
        Write-Error "错误: 'makensis.exe' 未找到。请确保 NSIS 已安装并已添加到系统的 PATH 环境变量中。"
        exit 1
    }
    
    $initialPath = $makensisCommand.Source
    $realMakensisPath = Join-Path -Path (Split-Path $initialPath -Parent) -ChildPath "makensis.shim"

    # Handle scoop shims by reading the real path from the shim file
    if ($initialPath -like "*\scoop\shims*") {
        Write-Host "Scoop shim 检测到。正在读取真实路径..."
        $shimContent = Get-Content $realMakensisPath -Raw
        $pathLine = $shimContent | Select-String -Pattern 'path\s*=\s*"(.*)"'
        if ($pathLine) {
            $realMakensisPath = $pathLine.Matches[0].Groups[1].Value
            Write-Host "真实 makensis 路径: $realMakensisPath"
        } else {
            Write-Error "无法从 scoop shim 文件中解析真实的 makensis 路径。"
            exit 1
        }
    }

    # Determine the NSIS root directory from the real executable path
    $nsisDir = Split-Path -Parent $realMakensisPath
    # For scoop installs, the structure is often '.../nsis/current/bin/makensis.exe',
    # so we need to go up one more level if the parent is 'bin'.
    if ((Split-Path $nsisDir -Leaf).ToLower() -eq 'bin') {
        $nsisDir = Split-Path -Parent $nsisDir
    }
    Write-Host "NSIS 安装目录: $nsisDir"

    # 3. Copy the nsisunz.dll plugin to the NSIS Plugins directory
    Write-Host "--- 正在安装 nsisunz.dll 插件 ---"
    $pluginSource = ".\nsisunz.dll"
    $pluginDestDir = Join-Path -Path $nsisDir -ChildPath "Plugins" -AdditionalChildPath "x86-unicode"
    
    if (-not (Test-Path $pluginSource)) {
        Write-Error "错误: 未找到 '.\nsisunz.dll'。"
        exit 1
    }
    if (-not (Test-Path $pluginDestDir)) {
        Write-Warning "警告: 找不到 NSIS 的 'Plugins' 目录 ($pluginDestDir)。跳过复制 nsisunz.dll。"
    } else {
        Write-Host "正在复制 nsisunz.dll 到 $pluginDestDir"
        Copy-Item -Path $pluginSource -Destination $pluginDestDir -Force
    }

    # 4. Convert the source NSI script to UTF-16BE for the compiler
    $nsiSource = ".\installer.nsi"
    $nsiConverted = ".\installer.utf16be.nsi"
    Write-Host "--- 正在转换 $nsiSource 为 UTF-16 BE ---"
    Get-Content $nsiSource -Raw | Out-File -FilePath $nsiConverted -Encoding BigEndianUnicode -Force
    Write-Host "转换完成: $nsiConverted"

    # 5. Run the makensis compiler and stream its output in real-time
    Write-Host "--- 正在运行 makensis ---"
    & $realMakensisPath /V2 $nsiConverted
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "❌ makensis 编译失败，返回代码: $LASTEXITCODE"
        exit $LASTEXITCODE
    } else {
        Write-Host "✅ makensis 编译成功。"
    }

} catch {
    Write-Error "脚本执行过程中发生错误: $_"
    exit 1
}

Write-Host "--- 构建脚本执行完毕 ---"
