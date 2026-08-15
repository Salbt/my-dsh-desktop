@echo off
setlocal

for %%I in ("%~dp0..") do set "DSH_ROOT=%%~fI"

if exist "%DSH_ROOT%\portable.marker" (
  set "DSH_DATA=%DSH_ROOT%"
) else if defined LOCALAPPDATA (
  set "DSH_DATA=%LOCALAPPDATA%\my-dsh-desktop"
) else (
  set "DSH_DATA=%TEMP%\my-dsh-desktop"
)

set "NODE_HOME="
for /d %%D in ("%DSH_ROOT%\runtime\node\node-*-win-x64") do if not defined NODE_HOME set "NODE_HOME=%%~fD"

if not defined NODE_HOME (
  >&2 echo dsh: bundled Node.js was not found under "%DSH_ROOT%\runtime\node"
  exit /b 1
)

set "DSH_ENTRY=%DSH_ROOT%\runtime\dsh\node_modules\@deepseek-ai\dsh\lib\bin.js"
if not exist "%DSH_ENTRY%" (
  >&2 echo dsh: DeepSeek Harness was not found under "%DSH_ROOT%\runtime\dsh"
  exit /b 1
)

set "DSH_HOME=%DSH_DATA%\home"
set "NPM_CONFIG_CACHE=%DSH_DATA%\npm-cache"
set "NPM_CONFIG_STORE_DIR=%DSH_DATA%\npm-cache\pnpm-store"
set "PATH=%NODE_HOME%;%DSH_ROOT%\bin;%DSH_ROOT%\runtime\pnpm\node_modules\.bin;%PATH%"

if not exist "%DSH_HOME%" mkdir "%DSH_HOME%" >nul 2>&1
if not exist "%NPM_CONFIG_CACHE%" mkdir "%NPM_CONFIG_CACHE%" >nul 2>&1

"%NODE_HOME%\node.exe" "%DSH_ENTRY%" %*
exit /b %ERRORLEVEL%
