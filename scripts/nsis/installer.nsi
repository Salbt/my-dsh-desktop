Unicode true
!ifndef VERSION
  !define VERSION "0.1.0"
!endif
Name "DeepSeek Harness Desktop"
OutFile "${OUT}"
InstallDir "$PROGRAMFILES64\my-dsh-desktop"
RequestExecutionLevel admin
Icon "..\..\assets\app.ico"
Page directory
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

Section
  SetOutPath $INSTDIR
  File /r /x portable.marker "${STAGE}\*"
  WriteUninstaller "$INSTDIR\uninstall.exe"
  CreateDirectory "$SMPROGRAMS\DeepSeek Harness Desktop"
  CreateShortcut "$SMPROGRAMS\DeepSeek Harness Desktop\DeepSeek Harness Desktop.lnk" "$INSTDIR\my-dsh-desktop.exe"
  CreateShortcut "$DESKTOP\DeepSeek Harness Desktop.lnk" "$INSTDIR\my-dsh-desktop.exe"
SectionEnd

Section "Uninstall"
  Delete "$DESKTOP\DeepSeek Harness Desktop.lnk"
  Delete "$SMPROGRAMS\DeepSeek Harness Desktop\DeepSeek Harness Desktop.lnk"
  RMDir "$SMPROGRAMS\DeepSeek Harness Desktop"
  RMDir /r "$INSTDIR"
SectionEnd
