$ErrorActionPreference = 'Stop'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  Write-Error "Go not found. Install Go 1.25+ first: https://go.dev/dl/"
  exit 1
}
if (-not (Get-Command gcc -ErrorAction SilentlyContinue) -or -not (Get-Command g++ -ErrorAction SilentlyContinue)) {
  Write-Error "webview_go is a CGO project and needs gcc/g++. Install MinGW-w64 and add it to PATH (e.g. choco install mingw -y)."
  exit 1
}

go run github.com/akavel/rsrc@latest -ico assets/app.ico -o rsrc_windows_amd64.syso
if (-not $?) { exit 1 }
go build -trimpath -ldflags "-s -w -H windowsgui" -o my-dsh-desktop.exe .
Write-Host "build ok: my-dsh-desktop.exe"
