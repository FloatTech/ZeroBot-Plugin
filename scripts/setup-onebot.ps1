[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$runtimeRoot = Join-Path $repoRoot '.local/onebot/runtime'
$configRoot = Join-Path $runtimeRoot 'config'
$bundleRoot = Join-Path $repoRoot '.local/onebot/napcat-v4.18.28'
if (-not (Test-Path -LiteralPath (Join-Path $bundleRoot 'node.exe'))) {
    throw 'NapCat v4.18.28 Windows Node bundle is not installed in .local/onebot.'
}
New-Item -ItemType Directory -Path $configRoot -Force | Out-Null

function New-LocalToken {
    $bytes = New-Object byte[] 32
    $rng = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return [BitConverter]::ToString($bytes).Replace('-', '').ToLowerInvariant()
}
function Write-LocalJson($Path, $Value) {
    [IO.File]::WriteAllText($Path, ($Value | ConvertTo-Json -Depth 10), [Text.UTF8Encoding]::new($false))
}

$webConfigPath = Join-Path $configRoot 'webui.json'
if (-not (Test-Path -LiteralPath $webConfigPath)) {
    Write-LocalJson $webConfigPath ([ordered]@{
        host = '127.0.0.1'; port = 6099; token = (New-LocalToken)
        loginRate = 5; autoLoginAccount = ''; disableWebUI = $false
        accessControlMode = 'whitelist'; ipWhitelist = @('127.0.0.1', '::1')
        ipBlacklist = @(); enableXForwardedFor = $false
    })
}
$onebotConfigPath = Join-Path $configRoot 'onebot11.json'
if (-not (Test-Path -LiteralPath $onebotConfigPath)) {
    Write-LocalJson $onebotConfigPath ([ordered]@{
        network = [ordered]@{
            httpServers = @(); httpSseServers = @(); httpClients = @()
            websocketServers = @([ordered]@{
                name = 'zerobot-local'; enable = $true; host = '127.0.0.1'
                port = 6700; messagePostFormat = 'array'; reportSelfMessage = $false
                token = (New-LocalToken); enableForcePushEvent = $true
                debug = $false; heartInterval = 30000
            })
            websocketClients = @(); plugins = @()
        }
        musicSignUrl = ''; enableLocalFile2Url = $false; parseMultMsg = $false
    })
}
$napcatConfigPath = Join-Path $configRoot 'napcat.json'
if (-not (Test-Path -LiteralPath $napcatConfigPath)) {
    Write-LocalJson $napcatConfigPath ([ordered]@{
        fileLog = $true; consoleLog = $true
        fileLogLevel = 'info'; consoleLogLevel = 'info'
        packetBackend = 'auto'; packetServer = ''; autoTimeSync = $false
    })
}

# Preserve existing user configuration and check that both tokens match.
$botConfigPath = Join-Path $repoRoot '.local/config.json'
$onebotConfig = Get-Content -LiteralPath $onebotConfigPath -Raw | ConvertFrom-Json
$wsServer = @($onebotConfig.network.websocketServers | Where-Object { $_.name -eq 'zerobot-local' })
if ($wsServer.Count -ne 1 -or $wsServer[0].host -ne '127.0.0.1') {
    throw 'Expected one loopback zerobot-local WebSocket server.'
}
if (-not (Test-Path -LiteralPath $botConfigPath)) {
    $botConfig = Get-Content -LiteralPath (Join-Path $repoRoot 'docs/local-admin/config.windows.example.json') -Raw | ConvertFrom-Json
    $botConfig.ws[0].Url = 'ws://127.0.0.1:' + $wsServer[0].port
    $botConfig.ws[0].AccessToken = $wsServer[0].token
    Write-LocalJson $botConfigPath $botConfig
}
$botConfig = Get-Content -LiteralPath $botConfigPath -Raw | ConvertFrom-Json
$paired = @($botConfig.ws | Where-Object { $_.Url -eq ('ws://127.0.0.1:' + $wsServer[0].port) -and $_.AccessToken -eq $wsServer[0].token })
if ($paired.Count -eq 0) { throw 'Existing bot config does not match NapCat. Review both local configs before proceeding.' }
$webConfig = Get-Content -LiteralPath $webConfigPath -Raw | ConvertFrom-Json
if ($webConfig.host -ne '127.0.0.1') { throw 'WebUI must bind only to 127.0.0.1.' }
Write-Host "Local OneBot configuration ready. WebUI: http://127.0.0.1:$($webConfig.port)/webui"
Write-Host "Bot WebSocket: ws://127.0.0.1:$($wsServer[0].port)"
Write-Host 'Tokens are stored in local config files and were not printed.'
