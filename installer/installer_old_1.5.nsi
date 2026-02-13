; NSIS Script for RPA Project Environment Installer (Optimized)
; 请将该文件保存为UTF16 BE 编码

!define PRODUCT_NAME "数字员工平台"
!define OLD_PRODUCT_NAME "公卫RPA运行时"
!define PRODUCT_VERSION "1.5"
!define COMPANY_NAME "无限视讯"
!define INSTALLER_OUTPUT "数字员工平台1.5.exe"

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

Function InstallDependencies
  SetRegView 64
  SetOutPath "$INSTDIR"
  ; $0: 传入 --force-reinstall (用于修复) 或 "" (用于安装)
  Pop $R0
  DetailPrint "正在准备 Python 依赖包..."
  File /r "packages"
  File "requirements.txt"

  DetailPrint "正在升级 Pip..."
  ; 升级 Pip，新版会显示进度条
  ExecWait '"$PYTHON_EXE" -m pip install --upgrade --no-index --find-links="$INSTDIR\packages" "pip>=24.0"'

  DetailPrint "正在安装依赖..."
  ExecWait '"$PYTHON_EXE" -m pip install $R0 --no-index --find-links="$INSTDIR\packages" -r "$INSTDIR\requirements.txt"' $1
  ${If} $1 != 0
    MessageBox MB_OK "依赖安装/修复返回代码: $1"
  ${EndIf}

  DetailPrint "清理临时文件..."
  RMDir /r "$INSTDIR\packages"
  Delete "$INSTDIR\requirements.txt"
FunctionEnd

Function CopyBrowser
  SetOutPath "$INSTDIR"
  DetailPrint "正在复制 Thorium 浏览器..."
  File /r "Thorium107"
FunctionEnd

Function CreateAssociations
  DetailPrint "正在关联文件并创建快捷方式..."
  WriteRegStr HKLM "Software\Classes\.pyz" "" "Python.File"
  WriteRegStr HKLM "Software\Classes\Python.File\shell\open\command" "" '"$PYTHON_EXE" "%1"'
  WriteRegStr HKLM "Software\Classes\Python.File\DefaultIcon" "" "$PYTHON_EXE,0"
  CreateShortCut "$DESKTOP\${PRODUCT_NAME}.lnk" "$PYTHON_EXE" "" "$PYTHON_EXE" 0
FunctionEnd

;--------------------------------
; 安装与修复区段
;--------------------------------

Section "安装/修复运行时" SecInstall
  SetRegView 64
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

  ; 动态检测 Python
  ReadRegStr $0 HKLM "SOFTWARE\Python\PythonCore\3.8\InstallPath"  ""
  ${If} ${Errors}
    ; 如果读取失败 (键不存在)，则说明 Python 未安装
    DetailPrint "未检测到 Python 3.8，正在开始安装..."
    File "python-3.8.10-amd64.exe"
    ExecWait '"$INSTDIR\python-3.8.10-amd64.exe" /quiet InstallAllUsers=1 PrependPath=1 TargetDir=$INSTDIR\python38 Include_test=0' $1

    ; 强制刷新环境变量，确保后续命令能正确找到Python
    SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000

    Delete "$INSTDIR\python-3.8.10-amd64.exe"
    ${If} $1 != 0
      MessageBox MB_ICONEXCLAMATION|MB_TOPMOST "Python 安装失败，返回代码: $1"
      Abort "Python 安装失败，无法继续。"
    ${EndIf}
    ; 再次读取以获取路径
    ReadRegStr $0 HKLM "SOFTWARE\Python\PythonCore\3.8\InstallPath" ""
  ${Else}
    DetailPrint "检测到 Python 3.8 已安装。"
  ${EndIf}

  ; 安装 VC++ 运行库
  DetailPrint "正在安装 Microsoft Visual C++ Redistributable..."
  File "VC_redist2015-2022.x64.exe"
  ExecWait '"$INSTDIR\VC_redist2015-2022.x64.exe" /install /passive /norestart'
  Delete "$INSTDIR\VC_redist2015-2022.x64.exe"
  DetailPrint "VC++ 运行库安装完成。"

  ; 将获取到的 Python 路径存入变量
  StrCpy $PYTHON_EXE "$0\python.exe"

  ; 调用函数执行任务
  Push "" ; 为 InstallDependencies 传入空参数
  Call InstallDependencies
  Call CreateAssociations
  Call CopyBrowser

  WriteUninstaller "$INSTDIR\uninstall.exe"
  DetailPrint "安装完成。"
SectionEnd

Section "修复环境" SecRepair
  SetRegView 64
  ; 修复模式下，Python 必须存在
  ReadRegStr $0 HKLM "SOFTWARE\Python\PythonCore\3.8\InstallPath" ""
  ${If} ${Errors}
    MessageBox MB_ICONSTOP "找不到 Python 3.8，无法修复！"
    Abort
  ${EndIf}
  StrCpy $PYTHON_EXE "$0\python.exe"

  DetailPrint "正在修复环境..."
  Call CopyBrowser
  Push "--force-reinstall" ; 为 InstallDependencies 传入修复参数
  Call InstallDependencies
  Call CreateAssociations
  DetailPrint "修复完成。"
SectionEnd

Section "升级" SecUpgrade
  SetRegView 64
  ReadRegStr $0 HKLM "SOFTWARE\Python\PythonCore\3.8\InstallPath" ""
  ${If} ${Errors}
    MessageBox MB_ICONSTOP "找不到 Python 3.8，无法升级！"
    Abort
  ${EndIf}
  StrCpy $PYTHON_EXE "$0\python.exe"

  DetailPrint "正在升级环境..."
  Call CopyBrowser
  Push "-U" ; 传递 pip install -U 参数
  Call InstallDependencies
  Call CreateAssociations

  ; 更新注册表版本号
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion" "${PRODUCT_VERSION}"
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
; 运行期初始化：决定执行哪一个区段
; 必须放在 Section 定义之后，才能引用 ${SecInstall} / ${SecRepair}
;--------------------------------
Function .onInit
  SetRegView 64
  ; 检查 D 盘是否存在，如果存在则修改默认安装目录
  IfFileExists "D:\" 0 +2
    StrCpy $INSTDIR "D:\wu-xian-shi-xun"

  ; 先检查新产品名的卸载注册表
  ReadRegStr $R1 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion"

  ${If} $R1 == ""
    ; 再检查旧产品名（兼容迁移）
    ReadRegStr $R1 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${OLD_PRODUCT_NAME}" "DisplayVersion"
    ${If} ${Errors}
      ; 未安装（两个键都不存在）
      StrCpy $R1 ""
    ${Else}
      ; 标记为来自旧产品名的安装（后续会迁移）
      StrCpy $R2 "legacy"
    ${EndIf}
  ${EndIf}

  ${If} $R1 == ""
    ; 未安装，执行全新安装
    Goto NotInstalled
  ${Else}
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
  ${EndIf}

DoRepair:
  ; 修复时也锁定目录，优先从新键读取，若没有则从旧键读取
  ReadRegStr $INSTDIR HKLM "Software\${PRODUCT_NAME}" "InstallDir"
  ${If} $INSTDIR == ""
    ReadRegStr $INSTDIR HKLM "Software\${OLD_PRODUCT_NAME}" "InstallDir"
  ${EndIf}
  SectionSetFlags ${SecRepair} ${SF_SELECTED}
  SectionSetFlags ${SecInstall} 0
  SectionSetFlags ${SecUpgrade} 0
  Return

DoUpgrade:
  ; 升级时锁定目录，优先从新键读取，若没有则从旧键读取
  ReadRegStr $INSTDIR HKLM "Software\${PRODUCT_NAME}" "InstallDir"
  ${If} $INSTDIR == ""
    ReadRegStr $INSTDIR HKLM "Software\${OLD_PRODUCT_NAME}" "InstallDir"
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
  ${If} $0 != ${SF_SELECTED}
    Abort
  ${EndIf}
FunctionEnd
