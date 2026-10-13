[CmdletBinding()]
param(
    [ValidateRange(10,86400)][int]$DurationSeconds=900,
    [ValidateRange(1,30)][int]$IntervalSeconds=5,
    [ValidateRange(0.01,100)][double]$MaxCpuPercent=20,
    [ValidateRange(0.01,4096)][double]$MaxPrivateGrowthMiB=64
)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'windows-service-identity.ps1')
. (Join-Path $PSScriptRoot 'live-observation-core.ps1')
$repoRoot=Resolve-LocalServicePath (Split-Path -Parent $PSScriptRoot)
$expected=Join-Path $repoRoot '.local/bin/zerobot.exe'
$statePath=Join-Path $repoRoot '.local/zerobot/process.json'
$state=Get-Content -LiteralPath $statePath -Raw|ConvertFrom-Json
$target=Get-Process -Id $state.processId
if(-not (Test-LocalServiceProcess $state $target $expected)){throw 'Managed daily ZeroBot identity could not be verified.'}
$logsRoot=(Join-Path $repoRoot '.local/services')+[IO.Path]::DirectorySeparatorChar
$cursors=@()
foreach($field in @('stdout','stderr')){
    $path=Resolve-LocalServicePath $state.$field
    if(-not $path.StartsWith($logsRoot,[StringComparison]::OrdinalIgnoreCase) -or [IO.Path]::GetFileName($path) -notmatch '^ZeroBot-[0-9-]+-(stdout|stderr)\.log$'){throw 'Managed service log path could not be verified.'}
    $cursors+=New-LiveLogCursor $path
}
$programHash=(Get-FileHash -LiteralPath $expected -Algorithm SHA256).Hash
$out=Join-Path $repoRoot ('.local/live-observation/run-'+(Get-Date -Format 'yyyyMMdd-HHmmss')+'-'+[guid]::NewGuid().ToString('N').Substring(0,8))
New-Item -ItemType Directory -Path $out|Out-Null
$utf8=[Text.UTF8Encoding]::new($false)
function Write-ObservationJSON([string]$Name,$Value){
    $path=Join-Path $out $Name
    [IO.File]::WriteAllText(($path+'.tmp'),($Value|ConvertTo-Json -Depth 8),$utf8)
    Move-Item -LiteralPath ($path+'.tmp') -Destination $path -Force
}
$samples=[Collections.Generic.List[object]]::new()
$result=[ordered]@{startedAt=(Get-Date).ToString('o');processId=$target.Id;processStartedAt=$target.StartTime.ToString('o');programSHA256=$programHash;requestedDurationSeconds=$DurationSeconds;intervalSeconds=$IntervalSeconds;logicalCPUs=[Environment]::ProcessorCount;thresholds=@{cpuAverageCapacityPercent=$MaxCpuPercent;privateGrowthMiB=$MaxPrivateGrowthMiB};completed=$false;passed=$false;identityStable=$true;stopReason='not-started';realQQMessagesSent=$false;modelCalls=0;businessLoadMeasured=$false;businessAcceptancePassed=$false;messageContentsSaved=$false;servicesStartedOrStopped=$false}
$timer=[Diagnostics.Stopwatch]::StartNew();$nextProgress=0.0;$stage='sampling';$failed=$false
Write-Host ('ARTIFACTS='+$out)
try {
    while($true){
        if(Test-Path -LiteralPath (Join-Path $out 'stop.request')){$result.stopReason='stop-request';break}
        $stage='process-identity'
        $fresh=Get-Content -LiteralPath $statePath -Raw|ConvertFrom-Json
        $target.Refresh()
        if($target.HasExited -or $fresh.processId -ne $state.processId -or -not (Test-LocalServiceProcess $fresh $target $expected)){$result.identityStable=$false;$result.stopReason='process-identity-changed';break}
        $stage='resource-sample'
        $sample=[ordered]@{at=(Get-Date).ToString('o');atSeconds=$timer.Elapsed.TotalSeconds;cpuSeconds=$target.TotalProcessorTime.TotalSeconds;privateBytes=$target.PrivateMemorySize64;workingSetBytes=$target.WorkingSet64;threadCount=$target.Threads.Count;handleCount=$target.HandleCount}
        $samples.Add([pscustomobject]$sample)
        [IO.File]::AppendAllText((Join-Path $out 'samples.jsonl'),(($sample|ConvertTo-Json -Compress)+[Environment]::NewLine),$utf8)
        $stage='log-counters'
        foreach($cursor in $cursors){Read-LiveLogDelta $cursor}
        $elapsed=$timer.Elapsed.TotalSeconds
        if($elapsed -ge $nextProgress){
            $progress=@{checkedAt=(Get-Date).ToString('o');state='observing';elapsedSeconds=$elapsed;sampleCount=$samples.Count;privateBytes=$sample.privateBytes;logCounts=(Get-LiveLogSummary $cursors)}
            Write-ObservationJSON 'progress.json' $progress
            Write-Host ('OBSERVING seconds={0:N1} samples={1}' -f $elapsed,$samples.Count)
            $nextProgress=$elapsed+60
        }
        if($sample.atSeconds-$samples[0].atSeconds -ge $DurationSeconds){$result.completed=$true;$result.stopReason='duration-complete';break}
        $remaining=$DurationSeconds-($sample.atSeconds-$samples[0].atSeconds)
        Start-Sleep -Milliseconds ([int][Math]::Max(1,[Math]::Min($IntervalSeconds*1000,$remaining*1000)))
    }
} catch {
    $failed=$true;$result.stopReason='observation-error';$result.failureStage=$stage
    if($stage -eq 'process-identity'){$result.identityStable=$false}
} finally {
    $timer.Stop()
    try {
        foreach($cursor in $cursors){for($attempt=0;$attempt -lt 4;$attempt++){Read-LiveLogDelta $cursor;if($cursor.UnreadBytes -eq 0){break}}}
        $logs=Get-LiveLogSummary $cursors
        $result.logCounts=$logs;$result.noNewErrorLines=($logs.errorLines+$logs.fatalLines+$logs.panicLines) -eq 0
    } catch {$result.logReadFailed=$true;$result.noNewErrorLines=$false;$failed=$true}
    $result.resources=Get-LiveResourceSummary $samples.ToArray() $DurationSeconds $IntervalSeconds $result.logicalCPUs $MaxCpuPercent $MaxPrivateGrowthMiB
    try {$result.programUnchanged=(Get-FileHash -LiteralPath $expected -Algorithm SHA256).Hash -eq $programHash} catch {$result.programUnchanged=$false}
    $result.passed=$result.completed -and $result.identityStable -and $result.programUnchanged -and $result.resources.fullTiming -and $result.resources.completeWindow -and $result.resources.cadencePassed -and $result.resources.resourceThresholdsPassed -and $result.noNewErrorLines -and $result.logCounts.coverageComplete -and -not $failed
    $result.completedAt=(Get-Date).ToString('o')
    Write-ObservationJSON 'result.json' $result
    Write-ObservationJSON 'progress.json' @{state='finished';completedAt=$result.completedAt;stopReason=$result.stopReason;sampleCount=$samples.Count;completed=$result.completed;passed=$result.passed}
    foreach($cursor in $cursors){$cursor.Prefix=''}
    $target.Dispose()
}
Write-Host ('OBSERVATION_COMPLETE completed={0} fullTiming={1} passed={2}' -f $result.completed,$result.resources.fullTiming,$result.passed)
if($failed){throw 'Live observation failed; the recorded failure stage identifies the check. Services were not changed.'}
