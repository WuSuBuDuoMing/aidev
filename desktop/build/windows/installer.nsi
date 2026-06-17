!include MUI2.nsh

Name "NeoCode"
OutFile "NeoCode-${VERSION}-windows-${ARCH}-installer.exe"
InstallDir "$LOCALAPPDATA\Programs\NeoCode"
InstallDirRegKey HKCU "Software\NeoCode" "InstallDir"

RequestExecutionLevel user
ManifestDPIAware true

!define MUI_ABORTWARNING
!define MUI_ICON "icon.ico"

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "SimpChinese"
!insertmacro MUI_LANGUAGE "English"

Section "NeoCode" SecNeoCode
  SetOutPath "$INSTDIR"
  File /r "app\*.*"
  WriteUninstaller "$INSTDIR\uninstall.exe"
  CreateDirectory "$SMPROGRAMS\NeoCode"
  CreateShortCut "$SMPROGRAMS\NeoCode\NeoCode.lnk" "$INSTDIR\NeoCode.exe"
  CreateShortCut "$DESKTOP\NeoCode.lnk" "$INSTDIR\NeoCode.exe"
  WriteRegStr HKCU "Software\NeoCode" "InstallDir" "$INSTDIR"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\NeoCode" "DisplayName" "NeoCode"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\NeoCode" "UninstallString" "$INSTDIR\uninstall.exe"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\NeoCode" "DisplayVersion" "${VERSION}"
SectionEnd

Section "Uninstall"
  RMDir /r "$INSTDIR"
  RMDir /r "$SMPROGRAMS\NeoCode"
  Delete "$DESKTOP\NeoCode.lnk"
  DeleteRegKey HKCU "Software\NeoCode"
  DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\NeoCode"
SectionEnd
