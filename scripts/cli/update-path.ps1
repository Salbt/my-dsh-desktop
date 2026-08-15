param(
  [Parameter(Mandatory = $true)]
  [ValidateSet('Add', 'Remove')]
  [string]$Action,

  [Parameter(Mandatory = $true)]
  [string]$Entry,

  [ValidateSet('User', 'Machine', 'Process')]
  [string]$Scope = 'Machine'
)

$ErrorActionPreference = 'Stop'
$target = [EnvironmentVariableTarget]::$Scope

function Normalize-PathEntry([string]$Value) {
  if ([string]::IsNullOrWhiteSpace($Value)) { return '' }
  return $Value.Trim().Trim('"').TrimEnd('\')
}

$wanted = Normalize-PathEntry $Entry
$current = [Environment]::GetEnvironmentVariable('Path', $target)
$entries = @(
  $current -split ';' |
    Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
)

$exists = $false
$updated = foreach ($item in $entries) {
  if ([StringComparer]::OrdinalIgnoreCase.Equals((Normalize-PathEntry $item), $wanted)) {
    $exists = $true
    if ($Action -eq 'Remove') { continue }
  }
  $item.Trim()
}

if ($Action -eq 'Add' -and -not $exists) {
  $updated = @($updated) + $Entry.Trim()
}

[Environment]::SetEnvironmentVariable('Path', (@($updated) -join ';'), $target)
