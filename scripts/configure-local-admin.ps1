[CmdletBinding()]
param([ValidateRange(1,65535)][int]$Port=3010,[switch]$Disable)

$ErrorActionPreference='Stop'
$repoRoot=Split-Path -Parent $PSScriptRoot
$directory=Join-Path $repoRoot '.local/admin'
New-Item -ItemType Directory -Path $directory -Force|Out-Null
$settings=@{enabled=(-not $Disable);address=('127.0.0.1:'+$Port)}
[IO.File]::WriteAllText((Join-Path $directory 'options.json'),($settings|ConvertTo-Json),[Text.UTF8Encoding]::new($false))
Write-Host ('Local admin '+$(if($Disable){'disabled'}else{'enabled at http://127.0.0.1:'+$Port})+'. Restart ZeroBot to apply.')
