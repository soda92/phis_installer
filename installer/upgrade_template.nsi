; NSIS Upgrade Script Template

!define PRODUCT_NAME "数字员工平台"
!define FROM_VERSION "%%FROM_VERSION%%"
!define TO_VERSION "%%TO_VERSION%%"

!define UPGRADE_NAME "${PRODUCT_NAME} ${TO_VERSION} (从 ${FROM_VERSION} 升级)"
!define INSTALLER_OUTPUT "数字员工平台_${FROM_VERSION}_to_${TO_VERSION}_upgrade.exe"

!include "MUI2.nsh"
!include "LogicLib.nsh"

Name "${UPGRADE_NAME}"
OutFile "${INSTALLER_OUTPUT}"
RequestExecutionLevel admin
InstallDir "$PROGRAMFILES\wu-xian-shi-xun" ; This default is overwritten by the detected path

Var PYTHON_EXE

!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_LANGUAGE "SimpChinese"

Function .onInit
  SetRegView 64
  
  ReadRegStr $INSTDIR HKLM "Software\${PRODUCT_NAME}" "InstallDir"
  IfErrors NoInstallFound
  
  ReadRegStr $PYTHON_EXE HKLM "Software\${PRODUCT_NAME}" "PythonPath"
  IfErrors NoPythonPath

  ; Check if the installed version is the one we are upgrading FROM
  ReadRegStr $R1 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion"
  IfErrors NoInstallFound ; If version key is missing, treat as not found
  
  ${If} $R1 != "${FROM_VERSION}"
    MessageBox MB_OK|MB_ICONSTOP "此升级包适用于从版本 ${FROM_VERSION} 升级。$\n您当前安装的版本是 $R1。请使用对应的升级包或完整安装包。"
    Abort
  ${EndIf} 
  
  StrCpy $PYTHON_EXE "$PYTHON_EXE\python.exe"
  Return

NoInstallFound:
  MessageBox MB_OK|MB_ICONSTOP "未找到 ${PRODUCT_NAME} 的现有安装。无法进行升级。\n请先运行完整安装包。"
  Abort
NoPythonPath:
  MessageBox MB_OK|MB_ICONSTOP "在注册表中找不到 Python 路径。安装可能已损坏，请重新进行完整安装。"
  Abort
FunctionEnd

Function SetEnvironmentVariable
  DetailPrint "设置 PYTHONUTF8=1 环境变量..."
  WriteRegStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PYTHONUTF8" "1"
FunctionEnd

Section "升级依赖包"
  SetOutPath "$INSTDIR"
  Call SetEnvironmentVariable
  
  DetailPrint "正在准备升级依赖包..."
  File /r "packages_upgrade_${FROM_VERSION}_to_${TO_VERSION}"
  File "requirements_upgrade_${FROM_VERSION}_to_${TO_VERSION}.txt"

  DetailPrint "正在升级依赖..."
  ExecWait '"$PYTHON_EXE" -m pip install --upgrade --no-index --find-links="$INSTDIR\packages_upgrade_${FROM_VERSION}_to_${TO_VERSION}" -r "$INSTDIR\requirements_upgrade_${FROM_VERSION}_to_${TO_VERSION}.txt"' $0
  
  ${If} $0 != 0
    MessageBox MB_OK|MB_ICONSTOP "依赖包升级失败，返回代码: $0。"
  ${Else}
    DetailPrint "依赖包升级成功。"
  ${EndIf}

  DetailPrint "清理临时文件..."
  RMDir /r "$INSTDIR\packages_upgrade_${FROM_VERSION}_to_${TO_VERSION}"
  Delete "$INSTDIR\requirements_upgrade_${FROM_VERSION}_to_${TO_VERSION}.txt"
  
  DetailPrint "清理不再需要的浏览器组件..."
  RMDir /r "$INSTDIR\Thorium107"

  ; Update the version in the registry
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" "DisplayVersion" "${TO_VERSION}"

  DetailPrint "升级完成。"
SectionEnd
