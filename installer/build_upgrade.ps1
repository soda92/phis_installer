# This script automates the build process for creating differential upgrade installers.

param(
    [Parameter(Mandatory=$true)]
    [string]$FromVersion,

    [Parameter(Mandatory=$true)]
    [string]$ToVersion
)

# Strict mode and error handling
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

try {
    # Ensure the script runs from its own directory
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    Set-Location $scriptDir
    Write-Host "当前工作目录: $(Get-Location)"

    # 1. Run the python script to create the differential package content
    Write-Host "--- 正在为 $FromVersion -> $ToVersion 创建差异升级包内容 ---"
    python .\create_upgrade_package.py $FromVersion $ToVersion
    Write-Host "--- 差异包内容创建完成 ---"

    # 2. Generate the version-specific NSIS script from the template
    Write-Host "--- 正在从模板生成 NSIS 脚本 ---"
    $template = Get-Content ".\upgrade_template.nsi" -Raw
    $template = $template -replace "%%FROM_VERSION%%", $FromVersion
    $template = $template -replace "%%TO_VERSION%%", $ToVersion
    
    $nsiSource = ".\upgrade_${FromVersion}_to_${ToVersion}.nsi"
    $template | Out-File $nsiSource -Encoding utf8
    Write-Host "已生成 NSIS 脚本: $nsiSource"

    # 3. Find NSIS installation directory, handling scoop shims
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


    $nsisDir = Split-Path -Parent $realMakensisPath
    if ((Split-Path $nsisDir -Leaf).ToLower() -eq 'bin') {
        $nsisDir = Split-Path -Parent $nsisDir
    }
    Write-Host "NSIS 安装目录: $nsisDir"

    # 4. Convert the generated NSI script to UTF-16BE
    $nsiConverted = ".\upgrade_${FromVersion}_to_${ToVersion}.utf16be.nsi"
    Write-Host "--- 正在转换 $nsiSource 为 UTF-16 BE ---"
    Get-Content $nsiSource -Raw | Out-File -FilePath $nsiConverted -Encoding BigEndianUnicode -Force
    Write-Host "转换完成: $nsiConverted"

    # 5. Run the makensis compiler
    Write-Host "--- 正在运行 makensis 编译升级包 ---"
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

Write-Host "--- 升级包构建脚本执行完毕 ---"
