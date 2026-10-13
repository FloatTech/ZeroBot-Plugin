[CmdletBinding()]
param([switch]$RunTests,[ValidatePattern('^[A-Za-z0-9][A-Za-z0-9_.-]*\.exe$')][string]$OutputName='zerobot.exe')

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$previousCgo = $env:CGO_ENABLED
Push-Location $repoRoot
try {
    $env:CGO_ENABLED = '0'
    & go version
    if ($LASTEXITCODE -ne 0) { throw 'Go toolchain is unavailable.' }
    & go mod download
    if ($LASTEXITCODE -ne 0) { throw 'Dependency download failed.' }
    & go mod verify
    if ($LASTEXITCODE -ne 0) { throw 'Dependency verification failed.' }
    & go generate main.go
    if ($LASTEXITCODE -ne 0) { throw 'Code generation failed.' }
    New-Item -ItemType Directory -Path '.local/bin' -Force | Out-Null
    & go build -trimpath -o (Join-Path '.local/bin' $OutputName) .
    if ($LASTEXITCODE -ne 0) { throw 'Windows build failed.' }
    if ($RunTests) {
        # Existing tests in other packages include live services and deliberate failures.
        # Even these packages may download fonts during package initialization.
        & go test ./internal/basiccommands ./internal/localadmin ./kanban ./plugin/chess ./plugin/tarot ./plugin/llm -count=1 -timeout=60s
        if ($LASTEXITCODE -ne 0) { throw 'Selected baseline tests failed.' }
        Push-Location (Join-Path $repoRoot 'third_party/go-registry')
        try {
            & go test ./... -count=1 -timeout=30s
            if ($LASTEXITCODE -ne 0) { throw 'Registry timeout tests failed.' }
        } finally { Pop-Location }
    }
    Write-Host "Build ready: $repoRoot/.local/bin/$OutputName"
} finally {
    $env:CGO_ENABLED = $previousCgo
    Pop-Location
}
