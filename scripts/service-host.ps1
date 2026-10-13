[CmdletBinding()]
param([Parameter(Mandatory=$true)][ValidateSet('OneBot','ZeroBot')][string]$Service)

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'windows-service-identity.ps1')
$repoRoot = Resolve-LocalServicePath (Split-Path -Parent $PSScriptRoot)
$runtimeRoot = Join-Path $repoRoot '.local/services'
New-Item -ItemType Directory -Path $runtimeRoot -Force | Out-Null
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss-fff'
if ($Service -eq 'OneBot') {
    $workingDirectory = Join-Path $repoRoot '.local/onebot/napcat-v4.18.28'
    $executable = Join-Path $workingDirectory 'node.exe'
    $arguments = @('index.js')
    $statePath = Join-Path $repoRoot '.local/onebot/runtime/process.json'
    $env:NAPCAT_WORKDIR = Join-Path $repoRoot '.local/onebot/runtime'
    $env:ACCOUNT = ''
} else {
    $workingDirectory = $repoRoot
    $executable = Join-Path $repoRoot '.local/bin/zerobot.exe'
    $arguments = @('-c', ('"' + (Join-Path $repoRoot '.local/config.json') + '"'))
    $adminDirectory=Join-Path $repoRoot '.local/admin'
    $adminOptionsPath=Join-Path $adminDirectory 'options.json'
    if(Test-Path -LiteralPath $adminOptionsPath){
        $adminOptions=Get-Content -LiteralPath $adminOptionsPath -Raw|ConvertFrom-Json
        if($adminOptions.enabled -isnot [bool]){throw 'Local admin enabled must be a boolean.'}
        if($adminOptions.enabled){
            $adminAddress=[string]$adminOptions.address
            if($adminAddress -notmatch '^127\.0\.0\.1:[1-9][0-9]{0,4}$'){throw 'Local admin address must be an IPv4 loopback address and port.'}
            $arguments+=@('-admin','-admin-addr',$adminAddress,'-admin-data',('"'+$adminDirectory+'"'))
        }
    }
    $statePath = Join-Path $repoRoot '.local/zerobot/process.json'
    $env:DEBUG_MODE = '1'
}
if (Test-Path -LiteralPath $statePath) {
    $old = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
    $existing = Get-Process -Id $old.processId -ErrorAction SilentlyContinue
    if ($existing -and (Test-LocalServiceProcess $old $existing $executable)) { return }
    if ($existing -and $existing.Path -and (Resolve-LocalServicePath $existing.Path) -eq $executable) {throw 'Existing service identity changed; refusing to create a duplicate.'}
}
$stdout = Join-Path $runtimeRoot ($Service + '-' + $stamp + '-stdout.log')
$stderr = Join-Path $runtimeRoot ($Service + '-' + $stamp + '-stderr.log')
# The scheduler starts this host with a hidden window. Its child stays attached
# to this persistent host, rather than to the temporary development tool.
$child = Start-Process -FilePath $executable -ArgumentList $arguments -WorkingDirectory $workingDirectory -NoNewWindow -RedirectStandardOutput $stdout -RedirectStandardError $stderr -PassThru
$state = [ordered]@{processId=$child.Id;hostProcessId=$PID;executable=$executable;startedAt=(Get-Date).ToString('o');processStartedAt=$child.StartTime.ToString('o');stdout=$stdout;stderr=$stderr;managedBy='Windows Task Scheduler'}
New-Item -ItemType Directory -Path (Split-Path $statePath) -Force | Out-Null
[IO.File]::WriteAllText($statePath, ($state | ConvertTo-Json), [Text.UTF8Encoding]::new($false))
Wait-Process -Id $child.Id
$child.Refresh()
exit $child.ExitCode
