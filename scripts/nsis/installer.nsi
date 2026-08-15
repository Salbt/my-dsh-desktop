Unicode true
!include "WinMessages.nsh"
!ifndef VERSION
  !define VERSION "0.1.0"
!endif
Name "DeepSeek Harness Desktop"
OutFile "${OUT}"
InstallDir "$PROGRAMFILES64\my-dsh-desktop"
InstallDirRegKey HKLM "Software\DeepSeek Harness Desktop" "InstallDir"
RequestExecutionLevel admin
Icon "..\..\assets\app.ico"
SetCompressor /SOLID lzma
Page directory
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

Section
  SetRegView 64
  SetShellVarContext all
  SetOutPath $INSTDIR
  File /r /x portable.marker "${STAGE}\*"
  WriteUninstaller "$INSTDIR\uninstall.exe"

  WriteRegStr HKLM "Software\DeepSeek Harness Desktop" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "DisplayName" "DeepSeek Harness Desktop"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "Publisher" "Salbt"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "DisplayIcon" "$INSTDIR\my-dsh-desktop.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop" "NoRepair" 1

  CreateDirectory "$SMPROGRAMS\DeepSeek Harness Desktop"
  CreateShortcut "$SMPROGRAMS\DeepSeek Harness Desktop\DeepSeek Harness Desktop.lnk" "$INSTDIR\my-dsh-desktop.exe"
  CreateShortcut "$DESKTOP\DeepSeek Harness Desktop.lnk" "$INSTDIR\my-dsh-desktop.exe"

  DetailPrint "Adding $INSTDIR\bin to the system PATH"
  ExecWait '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$INSTDIR\bin\update-path.ps1" -Action Add -Entry "$INSTDIR\bin" -Scope Machine' $0
  StrCmp $0 0 +2
    MessageBox MB_ICONEXCLAMATION|MB_OK "The app was installed, but its dsh command could not be added to PATH. You can still run $INSTDIR\bin\dsh.cmd directly."
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd

Section "Uninstall"
  SetRegView 64
  SetShellVarContext all
  DetailPrint "Removing $INSTDIR\bin from the system PATH"
  ExecWait '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$INSTDIR\bin\update-path.ps1" -Action Remove -Entry "$INSTDIR\bin" -Scope Machine'
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" /TIMEOUT=5000

  Delete "$DESKTOP\DeepSeek Harness Desktop.lnk"
  Delete "$SMPROGRAMS\DeepSeek Harness Desktop\DeepSeek Harness Desktop.lnk"
  RMDir "$SMPROGRAMS\DeepSeek Harness Desktop"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\my-dsh-desktop"
  DeleteRegKey HKLM "Software\DeepSeek Harness Desktop"
  RMDir /r "$INSTDIR"
SectionEnd
