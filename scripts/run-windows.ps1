[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$ConfigPath,
    [switch]$Headless
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$resolvedConfig = (Resolve-Path -LiteralPath $ConfigPath).Path
$binaryPath = Join-Path $repoRoot '.local/bin/zerobot.exe'
if (-not (Test-Path -LiteralPath $binaryPath)) {
    throw 'Build first with scripts/build-windows.ps1.'
}
if (-not (Test-Path -LiteralPath (Join-Path $repoRoot 'data/Control/kanban.png'))) {
    throw 'Runtime resources are missing. Initialize the data submodule first.'
}
$previousDebugMode = $env:DEBUG_MODE
Push-Location $repoRoot
try {
    # Upstream uses this variable to tolerate redirected Windows console handles.
    if ($Headless) { $env:DEBUG_MODE = '1' }
    & $binaryPath -c $resolvedConfig
    if ($LASTEXITCODE -ne 0) { throw "Bot exited with code $LASTEXITCODE." }
} finally {
    $env:DEBUG_MODE = $previousDebugMode
    Pop-Location
}
