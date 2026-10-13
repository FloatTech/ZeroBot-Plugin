[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
& (Join-Path $PSScriptRoot 'setup-onebot.ps1')
$bundleRoot = Join-Path $repoRoot '.local/onebot/napcat-v4.18.28'
$runtimeRoot = Join-Path $repoRoot '.local/onebot/runtime'
$nodePath = Join-Path $bundleRoot 'node.exe'
$statePath = Join-Path $runtimeRoot 'process.json'
if (Test-Path -LiteralPath $statePath) {
    $state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
    $existing = Get-Process -Id $state.processId -ErrorAction SilentlyContinue
    if ($existing) {
        if ($existing.Path -ne $nodePath) { throw 'Recorded PID belongs to another executable.' }
        Write-Host "Local NapCat is already running (PID $($existing.Id))."
        return
    }
}
$configRoot = Join-Path $runtimeRoot 'config'
$web = Get-Content -LiteralPath (Join-Path $configRoot 'webui.json') -Raw | ConvertFrom-Json
$onebot = Get-Content -LiteralPath (Join-Path $configRoot 'onebot11.json') -Raw | ConvertFrom-Json
$ports = @([int]$web.port) + @($onebot.network.websocketServers | Where-Object enable | ForEach-Object { [int]$_.port })
$listeners = [Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners()
foreach ($port in $ports) {
    if (@($listeners | Where-Object Port -eq $port).Count) { throw "Configured port is already in use: $port" }
}
# Run in the current terminal, which keeps startup and errors visible to the operator.
$previousWorkdir = $env:NAPCAT_WORKDIR
$previousAccount = $env:ACCOUNT
try {
    $env:NAPCAT_WORKDIR = $runtimeRoot
    $env:ACCOUNT = ''
    Push-Location -LiteralPath $bundleRoot
    try {
        $launcher = Start-Process -FilePath $nodePath -ArgumentList 'index.js' -WorkingDirectory $bundleRoot -NoNewWindow -PassThru
        $state = [ordered]@{processId=$launcher.Id;executable=$nodePath;startedAt=(Get-Date).ToString('o')}
        [IO.File]::WriteAllText($statePath, ($state | ConvertTo-Json), [Text.UTF8Encoding]::new($false))
        try {
            Wait-Process -Id $launcher.Id
            $launcher.Refresh()
            if ($launcher.ExitCode -ne 0) { throw "NapCat exited with code $($launcher.ExitCode)." }
        } finally { & (Join-Path $PSScriptRoot 'stop-onebot.ps1') }
    } finally { Pop-Location }
} finally {
    $env:NAPCAT_WORKDIR = $previousWorkdir
    $env:ACCOUNT = $previousAccount
}
