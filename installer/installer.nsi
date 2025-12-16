; NSIS Script for RPA Project Environment Installer (Optimized)
; 请将该文件保存为UTF16 BE 编码

!define PRODUCT_NAME "数字员工平台"
!define OLD_PRODUCT_NAME "公卫RPA运行时"
!define PRODUCT_VERSION "1.8"
!define COMPANY_NAME "无限视讯"
!define INSTALLER_OUTPUT "数字员工平台1.8.exe"

!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "WordFunc.nsh"

Name "${PRODUCT_NAME}"
OutFile "${INSTALLER_OUTPUT}"
InstallDir "C:\wu-xian-shi-xun"
InstallDirRegKey HKLM "Software\${PRODUCT_NAME}" "InstallDir"
RequestExecutionLevel admin

!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"
!define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\modern-uninstall.ico"
!define MUI_CUSTOMFUNCTION_ABORT CustomAbort

!insertmacro MUI_PAGE_WELCOME
!define MUI_PAGE_CUSTOMFUNCTION_PRE SkipDirPage
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "SimpChinese"

; 全局变量，用于存储 Python 可执行文件的路径
Var PYTHON_EXE

;--------------------------------
; 可重用函数
;--------------------------------

Function CleanupOnFailure
  SetRegView 64
  SetOutPath "$INSTDIR"
  DetailPrint "安装失败，正在执行清理..."
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"
  DeleteRegKey HKLM "Software\${PRODUCT_NAME}"
  ; 兼容性：如果旧键存在，也一并删除
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${OLD_PRODUCT_NAME}"
  DeleteRegKey HKLM "Software\${OLD_PRODUCT_NAME}"
  DetailPrint "清理完成。"
FunctionEnd

Function DeployPythonEmbeded
  SetRegView 64
  SetOutPath "$INSTDIR"

  ; --- 升级时总是重新部署嵌入式 Python, 兼容从 1.5 版本 (系统Python) 的升级 ---
  DetailPrint "正在部署嵌入式 Python 3.8..."
  SetOutPath "$INSTDIR"
  ; 1. 解压 Python 嵌入版
  File "python-3.8.10-embed-amd64.zip"
  CreateDirectory "$INSTDIR\python38-embed"
  nsisunz::UnzipToLog "$INSTDIR\python-3.8.10-embed-amd64.zip" "$INSTDIR\python38-embed"
  Delete "$INSTDIR\python-3.8.10-embed-amd64.zip"
  
  ; 2. 覆盖 ._pth 文件以启用 site-packages
  DetailPrint "配置 Python 环境..."
  SetOutPath "$INSTDIR\python38-embed"
  File "python38._pth"
  
  ; 3. 为离线安装 pip 创建临时目录并复制 wheels
  DetailPrint "正在准备离线安装 pip..."
  CreateDirectory "$INSTDIR\pip_wheels"
  SetOutPath "$INSTDIR\pip_wheels"
  File "pip-*.whl"
  File "setuptools-*.whl"
  File "wheel-*.whl"
  
  ; 4. 离线安装 pip
  SetOutPath "$INSTDIR\python38-embed"
  File "get-pip.py"
  ExecWait '"$INSTDIR\python38-embed\python.exe" "$INSTDIR\python38-embed\get-pip.py" --no-index --find-links="$INSTDIR\pip_wheels"' $1
  Delete "$INSTDIR\python38-embed\get-pip.py"
  RMDir /r "$INSTDIR\pip_wheels" ; 清理临时 wheels
  
  ${If} $1 != 0
    MessageBox MB_ICONEXCLAMATION|MB_TOPMOST "pip 离线安装失败，返回代码: $1"
    Call CleanupOnFailure
    Abort "pip 安装失败，无法继续。"
  ${EndIf}
  
  ; 将 Python 路径存入变量和我们自己的注册表键
  StrCpy $0 "$INSTDIR\python38-embed"
  WriteRegStr HKLM "Software\${PRODUCT_NAME}" "PythonPath" "$0"
  ; --- Python 部署结束 ---

  ; 将获取到的 Python 路径存入变量
  StrCpy $PYTHON_EXE "$0\python.exe"
FunctionEnd

Function InstallDependencies
  SetRegView 64
  SetOutPath "$INSTDIR"
  ; $0: 传入 --force-reinstall (用于修复) 或 "" (用于安装)
  Pop $R0
  DetailPrint "正在准备 Python 依赖包..."
  File /r "packages"
  File "requirements.txt"

  DetailPrint "正在安装依赖..."
  ExecWait '"$PYTHON_EXE" -m pip install $R0 --no-index --no-warn-script-location --find-links="$INSTDIR\packages" -r "$INSTDIR\requirements.txt"' $1
  ${If} $1 != 0
    MessageBox MB_OK "依赖安装/修复返回代码: $1。安装失败。"
    ; 在函数中返回错误状态
    SetErrors
  ${EndIf}

  DetailPrint "清理临时文件..."
  RMDir /r "$INSTDIR\packages"
  Delete "$INSTDIR\requirements.txt"
FunctionEnd



Function CreateAssociations
  DetailPrint "正在关联文件并创建快捷方式..."
  WriteRegStr HKLM "Software\Classes\.pyz" "" "Python.File"
  WriteRegStr HKLM "Software\Classes\Python.File\shell\open\command" "" '"$PYTHON_EXE" "%1"'
  WriteRegStr HKLM "Software\Classes\Python.File\DefaultIcon" "" "$PYTHON_EXE,0"
  CreateShortCut "$DESKTOP\${PRODUCT_NAME}.lnk" "$PYTHON_EXE" "" "$PYTHON_EXE" 0
FunctionEnd

Function SetEnvironmentVariable
  DetailPrint "设置 PYTHONUTF8=1 环境变量..."
  WriteRegStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PYTHONUTF8" "1"
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000
FunctionEnd

;--------------------------------
; 安装与修复区段
;--------------------------------

Section "安装/修复运行时" SecInstall
  SetRegView 64
  SetOutPath "$INSTDIR"
  ; 注册表信息（写入新名称）
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "Publisher" "${COMPANY_NAME}"
  WriteRegStr HKLM "Software\${PRODUCT_NAME}" "InstallDir" "$INSTDIR"

  ; 如果检测到旧产品注册表，迁移 InstallDir 并删除旧键（兼容）
  ReadRegStr $R3 HKLM "Software\${OLD_PRODUCT_NAME}" "InstallDir"
  ${If} ${Errors}
    ; 无旧键，跳过
  ${Else}
    ; 将旧安装目录写入新键（防止丢失），然后删除旧注册表信息
    WriteRegStr HKLM "Software\${PRODUCT_NAME}" "InstallDir" "$R3"
    DeleteRegKey HKLM "Software\${OLD_PRODUCT_NAME}"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${OLD_PRODUCT_NAME}"
  ${EndIf}

  ; 检查并安装 VC++ 运行库
  DetailPrint "检查 Microsoft Visual C++ Redistributable..."
  SetRegView 64
  ReadRegDWORD $R0 HKLM "SOFTWARE\Microsoft\VisualStudio\14.0\VC\Runtimes\x64" "Installed"
  ${If} $R0 == 1
    DetailPrint "VC++ Redistributable 已安装，跳过。"
  ${Else}
    DetailPrint "未检测到 VC++ Redistributable，正在安装..."
    File "VC_redist2015-2022.x64.exe"
    ExecWait '"$INSTDIR\VC_redist2015-2022.x64.exe" /install /passive /norestart' $0
    Delete "$INSTDIR\VC_redist2015-2022.x64.exe"
    
    ; VC++ 安装包在成功时可能返回特定代码 (如 3010 表示需要重启)，这里我们只检查严格的失败
    ${If} $0 != 0
    ${AndIf} $0 != 3010
      Call CleanupOnFailure
      Abort "VC++ 运行库安装失败。返回代码: $0"
    ${EndIf}
    DetailPrint "VC++ 运行库安装完成。"
  ${EndIf}

  Call SetEnvironmentVariable
  Call DeployPythonEmbeded

  ; 调用函数执行任务
  Push "" ; 为 InstallDependencies 传入空参数
  Call InstallDependencies
  ; 检查 InstallDependencies 是否设置了错误标志
  ${If} ${Errors}
    Call CleanupOnFailure
    Abort "依赖包安装失败。"
  ${EndIf}

  Call CreateAssociations


  WriteUninstaller "$INSTDIR\uninstall.exe"
  DetailPrint "安装完成。"
SectionEnd

Section "修复环境" SecRepair
  SetRegView 64
  ; 修复模式下，从我们自己的注册表键读取 Python 路径
  ReadRegStr $0 HKLM "Software\${PRODUCT_NAME}" "PythonPath"
  ${If} ${Errors}
      Call DeployPythonEmbeded
  ${EndIf}
  StrCpy $PYTHON_EXE "$0\python.exe"

  DetailPrint "正在修复环境..."

  
  Call SetEnvironmentVariable
  Push "--force-reinstall" ; 为 InstallDependencies 传入修复参数
  Call InstallDependencies
  Call CreateAssociations
  DetailPrint "修复完成。"
SectionEnd

Section "升级" SecUpgrade
  SetRegView 64
  SetOutPath "$INSTDIR"

  Call DeployPythonEmbeded

  ; 将获取到的 Python 路径存入变量
  StrCpy $PYTHON_EXE "$0\python.exe"

  DetailPrint "正在升级环境..."

  
  Call SetEnvironmentVariable
  Push "-U" ; 传递 pip install -U 参数
  Call InstallDependencies
  ; 检查 InstallDependencies 是否设置了错误标志
  ${If} ${Errors}
    Call CleanupOnFailure
    Abort "依赖包升级失败。"
  ${EndIf}
  Call CreateAssociations

  ; --- 新增：写入完整的卸载信息并清理旧版本 ---
  DetailPrint "正在更新应用程序注册表..."
  ; 写入新产品名的完整卸载信息
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "Publisher" "${COMPANY_NAME}"
  WriteRegStr HKLM "Software\${PRODUCT_NAME}" "InstallDir" "$INSTDIR"

  ; 创建新的卸载程序
  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; 删除旧产品名的注册表项
  DeleteRegKey HKLM "Software\${OLD_PRODUCT_NAME}"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${OLD_PRODUCT_NAME}"
  Delete "$DESKTOP\${OLD_PRODUCT_NAME}.lnk"
  ; --- 新增结束 ---

  DetailPrint "升级完成。"
SectionEnd

;--------------------------------
; 卸载区段
;--------------------------------
Section "Uninstall"
  DetailPrint "正在准备卸载..."
  SetRegView 64

  ; --- 清理本应用 ---
  DetailPrint "正在清理 ${PRODUCT_NAME} 的文件和注册表..."
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"
  DeleteRegKey HKLM "Software\${PRODUCT_NAME}"
  ; 兼容性：同时删除旧产品名的键
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${OLD_PRODUCT_NAME}"
  DeleteRegKey HKLM "Software\${OLD_PRODUCT_NAME}"
  Delete "$DESKTOP\${PRODUCT_NAME}.lnk"

  ; 删除安装目录，/r 表示递归删除
  RMDir /r "$INSTDIR"

  MessageBox MB_OK|MB_ICONINFORMATION "卸载完成。"
SectionEnd

;--------------------------------
; 运行期初始化与回调
;--------------------------------

; 当用户在安装过程中点击“取消”时调用
Function CustomAbort
  Call CleanupOnFailure
FunctionEnd

Function .onInit
  SetRegView 64
  SetOutPath "$INSTDIR"
  ; 检查 D 盘是否存在，如果存在则修改默认安装目录
  IfFileExists "D:\" 0 +2
    StrCpy $INSTDIR "D:\wu-xian-shi-xun"

  ; 检查新产品名的卸载注册表
  ReadRegStr $R1 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion"
  ${If} ${Errors}
    ; 新产品名不存在，检查旧产品名（兼容迁移）
    ReadRegStr $R1 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${OLD_PRODUCT_NAME}" "DisplayVersion"
    ${If} ${Errors}
      ; 新旧产品都不存在 -> 全新安装
      Goto NotInstalled
    ${Else}
      ; 找到了旧产品 -> 走升级/修复逻辑
      Goto ExistingInstallationFound
    ${EndIf}
  ${Else}
    ; 找到了新产品 -> 走升级/修复逻辑
    Goto ExistingInstallationFound
  ${EndIf}

ExistingInstallationFound:
  ; 已安装，比较版本
  ${VersionCompare} "${PRODUCT_VERSION}" "$R1" $R0
  ; $R0=0: 版本相同; $R0=1: 当前包版本新; $R0=2: 已安装版本新

  ${If} $R0 == 0
    ; 版本相同，提示修复
    MessageBox MB_YESNO|MB_ICONQUESTION \
      "检测到已安装相同版本 ($R1)。$\n是否要修复现有安装？" IDYES DoRepair
    Abort ; 用户选择否，则退出
  ${ElseIf} $R0 == 1
    ; 当前包版本新，提示升级
    MessageBox MB_YESNO|MB_ICONQUESTION \
      "检测到已安装版本 $R1，当前版本为 ${PRODUCT_VERSION}。$\n是否升级？" IDYES DoUpgrade
    Abort ; 用户选择否，则退出
  ${Else}
    ; 已安装版本更新，提示降级风险
    MessageBox MB_OK|MB_ICONEXCLAMATION \
      "检测到已安装了更新的版本 ($R1)。不支持降级安装。"
    Abort
  ${EndIf}

DoRepair:
  ; 修复时也锁定目录，优先从新键读取，若没有则从旧键读取
  ReadRegStr $INSTDIR HKLM "Software\${PRODUCT_NAME}" "InstallDir"
  ${If} ${Errors}
    ReadRegStr $INSTDIR HKLM "Software\${OLD_PRODUCT_NAME}" "InstallDir"
  ${EndIf}

  ; 如果安装目录不存在，视为未安装，允许用户重新选择目录
  ${IfNot} ${FileExists} "$INSTDIR"
    Goto NotInstalled
  ${EndIf}

  SectionSetFlags ${SecRepair} ${SF_SELECTED}
  SectionSetFlags ${SecInstall} 0
  SectionSetFlags ${SecUpgrade} 0
  Return

DoUpgrade:
  ; 升级时锁定目录，优先从新键读取，若没有则从旧键读取
  ReadRegStr $INSTDIR HKLM "Software\${PRODUCT_NAME}" "InstallDir"
  ${If} ${Errors}
    ReadRegStr $INSTDIR HKLM "Software\${OLD_PRODUCT_NAME}" "InstallDir"
  ${EndIf}

  ; 如果安装目录不存在，视为未安装，允许用户重新选择目录
  ${IfNot} ${FileExists} "$INSTDIR"
    Goto NotInstalled
  ${EndIf}

  SectionSetFlags ${SecUpgrade} ${SF_SELECTED}
  SectionSetFlags ${SecInstall} 0
  SectionSetFlags ${SecRepair} 0
  Return

NotInstalled:
  SectionSetFlags ${SecInstall} ${SF_SELECTED}
  SectionSetFlags ${SecRepair} 0
  SectionSetFlags ${SecUpgrade} 0
FunctionEnd

Function SkipDirPage
  ; 如果是修复或升级模式，则跳过目录选择页
  SectionGetFlags ${SecInstall} $0
  IntOp $0 $0 & ${SF_SELECTED}
  ${If} $0 == ${SF_SELECTED}
    ; 这是全新安装，不要跳过
  ${Else}
    ; 这是修复或升级，跳过目录页
    Abort
  ${EndIf}
FunctionEnd
