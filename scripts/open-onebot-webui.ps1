[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$webConfig = Get-Content -LiteralPath (Join-Path $repoRoot '.local/onebot/runtime/config/webui.json') -Raw | ConvertFrom-Json
if ($webConfig.host -ne '127.0.0.1') { throw 'WebUI must bind only to 127.0.0.1.' }
$url = 'http://127.0.0.1:' + $webConfig.port + '/webui/?token=' + [Uri]::EscapeDataString($webConfig.token)
# This is an interactive browser the user needs for QQ login. Keep the token out of console output.
Start-Process $url
