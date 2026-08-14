param(
  [string]$NodeVersion = "v24.19.0",
  [string]$DshVersion = "0.1.0-rc.6",
  [string]$AppVersion = "0.1.0",
  [string]$Registry = "",
  [string]$OutDir = "dist"
)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

& "$PSScriptRoot\..\build.ps1"
if (-not $?) { throw "build failed" }

$stage = Join-Path $env:TEMP "dsh-stage"
if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
New-Item -ItemType Directory -Path "$stage\runtime" -Force | Out-Null
Copy-Item "my-dsh-desktop.exe" $stage
New-Item -ItemType File -Path "$stage\portable.marker" -Force | Out-Null
Write-Host "stage: $stage"

$nodeDir = "$stage\runtime\node"
New-Item -ItemType Directory -Path $nodeDir -Force | Out-Null
$localNode = "$env:LOCALAPPDATA\my-dsh-desktop\runtime\node\node-$NodeVersion-win-x64"
if (Test-Path $localNode) {
  Write-Host "reuse local node: $localNode"
  $nodeVerDir = "$nodeDir\node-$NodeVersion-win-x64"
  robocopy $localNode $nodeVerDir /E /NFL /NDL /NJH /NJS /NC /NS | Out-Null
  if ($LASTEXITCODE -ge 8) { throw "robocopy node failed" }
} else {
  $zip = Join-Path $env:TEMP "node-$NodeVersion-win-x64.zip"
  Invoke-WebRequest -Uri "https://nodejs.org/dist/$NodeVersion/node-$NodeVersion-win-x64.zip" -OutFile $zip -UseBasicParsing
  Expand-Archive -Path $zip -DestinationPath $nodeDir -Force
  Remove-Item $zip
}
$nodeExe = Join-Path $nodeDir "node-$NodeVersion-win-x64\node.exe"
$npmCli = Join-Path $nodeDir "node-$NodeVersion-win-x64\node_modules\npm\bin\npm-cli.js"
if (-not (Test-Path $nodeExe)) { throw "node.exe not found" }

$dshDir = "$stage\runtime\dsh"
$localDsh = "$env:LOCALAPPDATA\my-dsh-desktop\runtime\dsh"
$localDshMeta = "$localDsh\node_modules\@deepseek-ai\dsh\package.json"
$localDshVersion = $null
if (Test-Path $localDshMeta) {
  $localDshVersion = (Get-Content $localDshMeta -Raw | ConvertFrom-Json).version
}
if ($localDshVersion -eq $DshVersion -and (Test-Path "$localDsh\node_modules\@deepseek-ai\dsh\lib\bin.js")) {
  Write-Host "reuse local dsh: $localDsh ($localDshVersion)"
  robocopy $localDsh $dshDir /E /NFL /NDL /NJH /NJS /NC /NS | Out-Null
  if ($LASTEXITCODE -ge 8) { throw "robocopy dsh failed" }
} else {
  Write-Host "local dsh version mismatch (local=$localDshVersion want=$DshVersion), install via npm..."
  $cache = Join-Path $env:TEMP "dsh-npm-cache"
  $npmArgs = @($npmCli, "install", "--prefix", $dshDir, "--no-audit", "--no-fund", "--cache", $cache)
  if ($Registry) { $npmArgs += @("--registry", $Registry) }
  $npmArgs += "@deepseek-ai/dsh@$DshVersion"
  & $nodeExe $npmArgs
  if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
}

$zipName = "my-dsh-desktop-portable-v$AppVersion.zip"
$zipPath = Join-Path $OutDir $zipName
if (Test-Path $zipPath) { Remove-Item $zipPath }
$tar = Get-Command tar.exe -ErrorAction SilentlyContinue
if ($tar) {
  Write-Host "compressing with bsdtar..."
  & $tar.Source -a -cf $zipPath -C $stage .
  if ($LASTEXITCODE -ne 0) { throw "tar failed" }
} else {
  Compress-Archive -Path "$stage\*" -DestinationPath $zipPath -CompressionLevel Fastest
}
Write-Host "wrote $zipPath ($([math]::Round((Get-Item $zipPath).Length/1MB)) MB)"

$hashPath = Join-Path $OutDir "sha256.txt"
$hash = Get-FileHash $zipPath -Algorithm SHA256
Set-Content -Path $hashPath -Value "$($hash.Hash)  $zipName"

$makensis = Get-Command makensis -ErrorAction SilentlyContinue
if ($makensis) {
  $nsi = Join-Path $PSScriptRoot "nsis\installer.nsi"
  $out = Join-Path $OutDir "my-dsh-desktop-setup-v$AppVersion.exe"
  & $makensis.Source /DVERSION=$AppVersion /DSTAGE=$stage /DOUT=$out $nsi
  if ($LASTEXITCODE -ne 0) { throw "makensis failed" }
  $hash = Get-FileHash $out -Algorithm SHA256
  Add-Content -Path $hashPath -Value "$($hash.Hash)  $(Split-Path $out -Leaf)"
  Write-Host "wrote $out ($([math]::Round((Get-Item $out).Length/1MB)) MB)"
} else {
  Write-Warning "makensis not found, skip installer"
}

Write-Host "package done"
