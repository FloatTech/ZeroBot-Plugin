[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$configRoot = Join-Path $repoRoot '.local/onebot/runtime/config'
$web = Get-Content -LiteralPath (Join-Path $configRoot 'webui.json') -Raw | ConvertFrom-Json
$onebot = Get-Content -LiteralPath (Join-Path $configRoot 'onebot11.json') -Raw | ConvertFrom-Json
$bot = Get-Content -LiteralPath (Join-Path $repoRoot '.local/config.json') -Raw | ConvertFrom-Json
$accountPath = Join-Path $repoRoot '.local/onebot/account.json'
$expectedQQ = $null
if (Test-Path -LiteralPath $accountPath) { $expectedQQ = (Get-Content -LiteralPath $accountPath -Raw | ConvertFrom-Json).expectedQQ }
if ($web.host -ne '127.0.0.1') { throw 'WebUI is not restricted to loopback.' }
$server = @($onebot.network.websocketServers | Where-Object { $_.name -eq 'zerobot-local' })
if ($server.Count -ne 1 -or $server[0].host -ne '127.0.0.1' -or -not $server[0].enable) {
    throw 'Expected one enabled loopback WebSocket server.'
}
$endpoint = 'ws://127.0.0.1:' + $server[0].port
if (-not @($bot.ws | Where-Object { $_.Url -eq $endpoint -and $_.AccessToken -eq $server[0].token }).Count) {
    throw 'ZeroBot and NapCat credentials do not match.'
}
$baseUrl = 'http://127.0.0.1:' + $web.port
$page = Invoke-WebRequest -Uri ($baseUrl + '/webui/') -UseBasicParsing -TimeoutSec 5
if ([int]$page.StatusCode -ne 200) { throw 'NapCat login page is unavailable.' }
$denied = Invoke-RestMethod -Method Post -Uri ($baseUrl + '/api/QQLogin/CheckLoginStatus') -ContentType 'application/json' -Body '{}' -TimeoutSec 5
if ($denied.code -eq 0) { throw 'NapCat allowed an unauthenticated status request.' }
$sha = [Security.Cryptography.SHA256]::Create()
try {
    $hash = [BitConverter]::ToString($sha.ComputeHash([Text.Encoding]::UTF8.GetBytes($web.token + '.napcat'))).Replace('-', '').ToLowerInvariant()
} finally { $sha.Dispose() }
$auth = Invoke-RestMethod -Method Post -Uri ($baseUrl + '/api/auth/login') -ContentType 'application/json' -Body (@{hash = $hash} | ConvertTo-Json) -TimeoutSec 5
if ($auth.code -ne 0 -or -not $auth.data.Credential) { throw 'NapCat WebUI authentication failed.' }
$headers = @{Authorization = 'Bearer ' + $auth.data.Credential}
$ws = $null
try {
    $status = Invoke-RestMethod -Method Post -Uri ($baseUrl + '/api/QQLogin/CheckLoginStatus') -Headers $headers -ContentType 'application/json' -Body '{}' -TimeoutSec 5
    if ($status.code -ne 0) { throw 'Authenticated QQ login status request failed.' }
    $result = [ordered]@{
        webUI = $baseUrl + '/webui/'; webUIReady = $true
        unauthenticatedRequestRejected = $true; credentialsMatch = $true
        qqLoggedIn = [bool]$status.data.isLogin; loginPhase = $status.data.loginPhase
        onebotEndpoint = $endpoint; onebotReadOnlyAPIsPassed = $false
    }
    if ($status.data.isLogin) {
        $ws = [Net.WebSockets.ClientWebSocket]::new()
        $ws.Options.SetRequestHeader('Authorization', 'Bearer ' + $server[0].token)
        $ws.Options.SetRequestHeader('X-Client-Role', 'Universal')
        $cts = [Threading.CancellationTokenSource]::new(15000)
        try {
            $ws.ConnectAsync([uri]$endpoint, $cts.Token).GetAwaiter().GetResult() | Out-Null
            foreach ($action in @('get_login_info','get_status','get_version_info')) {
                $echo = 'local-check-' + $action
                $payload = [Text.Encoding]::UTF8.GetBytes((@{action=$action;params=@{};echo=$echo} | ConvertTo-Json -Compress))
                $ws.SendAsync([ArraySegment[byte]]::new($payload), [Net.WebSockets.WebSocketMessageType]::Text, $true, $cts.Token).GetAwaiter().GetResult() | Out-Null
                do {
                    $buffer = New-Object byte[] 8192
                    $message = [IO.MemoryStream]::new()
                    try {
                        do {
                            $received = $ws.ReceiveAsync([ArraySegment[byte]]::new($buffer), $cts.Token).GetAwaiter().GetResult()
                            if ($received.MessageType -eq [Net.WebSockets.WebSocketMessageType]::Close) { throw 'OneBot closed the connection.' }
                            $message.Write($buffer,0,$received.Count)
                        } while (-not $received.EndOfMessage)
                        $reply = [Text.Encoding]::UTF8.GetString($message.ToArray()) | ConvertFrom-Json
                    } finally { $message.Dispose() }
                } while ($reply.echo -ne $echo)
                if ($reply.status -ne 'ok' -or $reply.retcode -ne 0) { throw "OneBot read-only API failed: $action" }
                if ($action -eq 'get_login_info' -and $expectedQQ -and [string]$reply.data.user_id -ne [string]$expectedQQ) {
                    throw 'Logged-in QQ does not match .local/onebot/account.json; refusing to use this account.'
                }
                if ($action -eq 'get_status' -and -not $reply.data.online) { throw 'OneBot reports QQ offline.' }
            }
            $result.onebotReadOnlyAPIsPassed = $true
        } finally { $cts.Dispose() }
    }
    $result | ConvertTo-Json
} finally {
    if ($ws) { $ws.Dispose() }
    Invoke-RestMethod -Method Post -Uri ($baseUrl + '/api/auth/logout') -Headers $headers -ContentType 'application/json' -Body '{}' -TimeoutSec 5 | Out-Null
}
