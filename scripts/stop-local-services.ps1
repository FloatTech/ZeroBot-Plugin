[CmdletBinding()]
param([ValidateSet('OneBot','ZeroBot','All')][string]$Service='All', [switch]$Unregister)

$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'windows-service-identity.ps1')
$repoRoot=Resolve-LocalServicePath (Split-Path -Parent $PSScriptRoot)
$sha=[Security.Cryptography.SHA256]::Create()
try { $id=[BitConverter]::ToString($sha.ComputeHash([Text.Encoding]::UTF8.GetBytes($repoRoot.ToLowerInvariant()))).Replace('-','').Substring(0,8) } finally { $sha.Dispose() }
$hostScript=Join-Path $PSScriptRoot 'service-host.ps1'
foreach($currentService in @('ZeroBot','OneBot')){
    if($Service -ne 'All' -and $Service -ne $currentService){continue}
    $taskName='ZeroBot-Local-'+$id+'-'+$currentService
    $exact=Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if($exact -and -not (Test-LocalServiceTask $exact $repoRoot $currentService)){throw 'Task definition does not match this project; refusing to stop it.'}
    $tasks=@(Get-ScheduledTask -TaskName ('ZeroBot-Local-*-'+$currentService) -ErrorAction SilentlyContinue | Where-Object {Test-LocalServiceTask $_ $repoRoot $currentService})
    if(-not $tasks.Count){continue}
    if($currentService -eq 'ZeroBot'){
        $statePath=Join-Path $repoRoot '.local/zerobot/process.json'
        $expectedExecutable=Join-Path $repoRoot '.local/bin/zerobot.exe'
    } else {
        $statePath=Join-Path $repoRoot '.local/onebot/runtime/process.json'
        $expectedExecutable=Join-Path $repoRoot '.local/onebot/napcat-v4.18.28/node.exe'
    }
    if(Test-Path -LiteralPath $statePath){
        $state=Get-Content -LiteralPath $statePath -Raw|ConvertFrom-Json
        $child=Get-Process -Id $state.processId -ErrorAction SilentlyContinue
        if($child){
            if(-not (Test-LocalServiceProcess $state $child $expectedExecutable)){throw 'Runtime process identity does not match this project; refusing to stop it.'}
            # Stopping the scheduler host alone leaves its child running. Terminate
            # the verified service tree as well (NapCat can spawn Node workers).
            & (Join-Path $env:WINDIR 'System32/taskkill.exe') /PID $child.Id /T /F
            if($LASTEXITCODE -ne 0){throw 'Failed to stop the verified local service process tree.'}
        }
    }
    foreach($task in $tasks){
        if($task.State -eq 'Running'){Stop-ScheduledTask -TaskName $task.TaskName}
        if($Unregister){Unregister-ScheduledTask -TaskName $task.TaskName -Confirm:$false}
        Write-Host ('Stopped local task: '+$task.TaskName)
    }
}
