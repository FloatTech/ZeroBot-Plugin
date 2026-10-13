[CmdletBinding()]
param([ValidateSet('OneBot','ZeroBot','All')][string]$Service = 'All')

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'windows-service-identity.ps1')
$repoRoot = Resolve-LocalServicePath (Split-Path -Parent $PSScriptRoot)
& (Join-Path $PSScriptRoot 'setup-onebot.ps1')
$accountPath = Join-Path $repoRoot '.local/onebot/account.json'
if (-not (Test-Path -LiteralPath $accountPath)) { throw 'Set the expected bot QQ in .local/onebot/account.json first.' }
$account = Get-Content -LiteralPath $accountPath -Raw | ConvertFrom-Json
if ($account.expectedQQ -notmatch '^[1-9][0-9]{4,11}$') { throw 'Expected bot QQ is invalid.' }
$sha = [Security.Cryptography.SHA256]::Create()
try { $id = [BitConverter]::ToString($sha.ComputeHash([Text.Encoding]::UTF8.GetBytes($repoRoot.ToLowerInvariant()))).Replace('-','').Substring(0,8) } finally { $sha.Dispose() }
$runtimeRoot = Join-Path $repoRoot '.local/services'
New-Item -ItemType Directory -Path $runtimeRoot -Force | Out-Null
$shellPath = (Get-Command pwsh.exe -ErrorAction SilentlyContinue).Source
if (-not $shellPath) { $shellPath = Join-Path $env:WINDIR 'System32/WindowsPowerShell/v1.0/powershell.exe' }
$hostScript = Join-Path $repoRoot 'scripts/service-host.ps1'
$userName = [Security.Principal.WindowsIdentity]::GetCurrent().Name
$taskNames = [ordered]@{}
foreach ($currentService in @('OneBot','ZeroBot')) {
    if ($Service -ne 'All' -and $Service -ne $currentService) { continue }
    if ($currentService -eq 'ZeroBot') {
        $health = & (Join-Path $PSScriptRoot 'test-onebot.ps1') | ConvertFrom-Json
        if (-not $health.qqLoggedIn -or -not $health.onebotReadOnlyAPIsPassed) {
            throw 'OneBot is not logged in to the expected account. Complete login before starting ZeroBot.'
        }
    }
    $taskName = 'ZeroBot-Local-' + $id + '-' + $currentService
    $arguments = '-NoProfile -WindowStyle Hidden -File "' + $hostScript + '" -Service ' + $currentService
    $task = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    $active=@(Get-ScheduledTask -TaskName ('ZeroBot-Local-*-'+$currentService) -ErrorAction SilentlyContinue | Where-Object {
        $_.State -eq 'Running' -and (Test-LocalServiceTask $_ $repoRoot $currentService)
    })
    if($active.Count -gt 1){throw 'Multiple matching service tasks are running; inspect before starting another.'}
    if($active.Count -eq 1){$task=$active[0];$taskName=$task.TaskName}
    if ($task) {
        if (-not (Test-LocalServiceTask $task $repoRoot $currentService)) { throw 'Existing task definition does not match this local project.' }
    } else {
        $action = New-ScheduledTaskAction -Execute $shellPath -Argument $arguments -WorkingDirectory $repoRoot
        $principal = New-ScheduledTaskPrincipal -UserId $userName -LogonType Interactive -RunLevel Limited
        $settings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit ([TimeSpan]::Zero) -MultipleInstances IgnoreNew -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
        Register-ScheduledTask -TaskName $taskName -Action $action -Principal $principal -Settings $settings -Description ('On-demand local '+$currentService+' host for '+$repoRoot) | Out-Null
    }
    if($task -and $task.State -eq 'Running'){Write-Host ('Reusing local task: '+$taskName)}
    else {Start-ScheduledTask -TaskName $taskName; Write-Host ('Started local task: ' + $taskName)}
    $taskNames[$currentService] = $taskName
    if ($currentService -eq 'OneBot') {
        $web = Get-Content -LiteralPath (Join-Path $repoRoot '.local/onebot/runtime/config/webui.json') -Raw | ConvertFrom-Json
        $ready = $false
        for ($attempt=0; $attempt -lt 15; $attempt++) {
            Start-Sleep -Seconds 2
            try { $page=Invoke-WebRequest -Uri ('http://127.0.0.1:'+$web.port+'/webui/') -UseBasicParsing -TimeoutSec 2; if ($page.StatusCode -eq 200) { $ready=$true; break } } catch { }
        }
        if (-not $ready) { throw 'NapCat page did not become ready. Inspect .local/services logs and task status.' }
    }
}
$manifestPath = Join-Path $runtimeRoot 'tasks.json'
$manifest = [ordered]@{}
if (Test-Path -LiteralPath $manifestPath) { $previous=Get-Content -LiteralPath $manifestPath -Raw|ConvertFrom-Json; foreach($property in $previous.psobject.Properties){$manifest[$property.Name]=$property.Value} }
foreach($key in $taskNames.Keys){$manifest[$key]=$taskNames[$key]}
[IO.File]::WriteAllText($manifestPath,($manifest|ConvertTo-Json),[Text.UTF8Encoding]::new($false))
