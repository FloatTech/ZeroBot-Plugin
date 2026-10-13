[CmdletBinding()]
param()
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'live-observation-core.ps1')
$repoRoot=Split-Path -Parent $PSScriptRoot
$fixture=Join-Path $repoRoot ('.local/live-observation-tests-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $fixture|Out-Null
$utf8=[Text.UTF8Encoding]::new($false)
$cases=[Collections.Generic.List[object]]::new()
function Assert-Observation([bool]$Condition,[string]$Name){if(-not $Condition){throw ('FAILED: '+$Name)}}
function Record-Observation([string]$Name){$cases.Add(@{name=$Name;passed=$true})}
function Fixture-Log([string]$Name,[string]$Text){$path=Join-Path $fixture $Name;[IO.File]::WriteAllText($path,$Text,$utf8);return $path}
function Append-Fixture([string]$Path,[string]$Text){[IO.File]::AppendAllText($Path,$Text,$utf8)}
function Drain-Fixture([hashtable]$Cursor,[int]$Limit=262144){do{Read-LiveLogDelta $Cursor $Limit}while($Cursor.UnreadBytes -gt 0)}
function Sample-Fixture([double]$At,[double]$CPU,[long]$Private=67108864){return [pscustomobject]@{atSeconds=$At;cpuSeconds=$CPU;privateBytes=$Private;workingSetBytes=33554432;threadCount=8;handleCount=100}}

$log=Fixture-Log 'append.log' "time=`"old`" level=error msg=`"old error`"`n"
$cursor=New-LiveLogCursor $log
Append-Fixture $log "time=`"new`" level=warning msg=`"synthetic warning`"`ntime=`"new`" level=error msg=`"synthetic error`"`ntime=`"new`" level=fatal msg=`"synthetic fatal`"`nlevel=panic msg=`"synthetic panic`"`npanic: native fixture`n"
Drain-Fixture $cursor;$summary=Get-LiveLogSummary @($cursor)
Assert-Observation ($summary.appendedLines -eq 5 -and $summary.warningLines -eq 1 -and $summary.errorLines -eq 1 -and $summary.fatalLines -eq 1 -and $summary.panicLines -eq 2 -and $summary.coverageComplete) 'Only new log records count, with separate severities'
Record-Observation 'ignores historical data; appended warning/error/fatal/panic counters are exact'

$log=Fixture-Log 'spoof.log' '';$cursor=New-LiveLogCursor $log
Append-Fixture $log "time=`"new`" level=info msg=`"synthetic level=error password=synthetic-only-secret`"`n"
Drain-Fixture $cursor 3;$summary=Get-LiveLogSummary @($cursor)
Assert-Observation ($summary.errorLines -eq 0 -and $summary.otherLines -eq 1 -and $summary.coverageComplete) 'Message body must not be classified as metadata'
Assert-Observation (($summary|ConvertTo-Json -Depth 5) -notmatch 'synthetic|password|msg=') 'Result must not contain body or secret fragments'
Record-Observation 'metadata split across reads is reconstructed; body severity spoof and secrets stay out of results'

$log=Fixture-Log 'partial.log' '';$cursor=New-LiveLogCursor $log
Append-Fixture $log 'time="new" level=error msg="synthetic未完成'
Read-LiveLogDelta $cursor 10;$summary=Get-LiveLogSummary @($cursor)
Assert-Observation ($summary.appendedLines -eq 0 -and $summary.unreadBytes -gt 0 -and $summary.partialLines -eq 1 -and -not $summary.coverageComplete) 'Backlog and partial lines must remain explicit'
Drain-Fixture $cursor;Append-Fixture $log "结尾`"`r`n";Drain-Fixture $cursor 2;$summary=Get-LiveLogSummary @($cursor)
Assert-Observation ($summary.errorLines -eq 1 -and $summary.appendedLines -eq 1 -and $summary.coverageComplete) 'Split Unicode payload and CRLF must count once'
Record-Observation 'bounded reads flag backlog and partial records, then count a completed CRLF record once'

$log=Fixture-Log 'long.log' '';$cursor=New-LiveLogCursor $log
Append-Fixture $log ('level=warning msg="'+('synthetic-only-secret'*10000))
Read-LiveLogDelta $cursor
Assert-Observation ($cursor.Prefix.Length -le 4096 -and $cursor.LongLine) 'Long record prefix must stay bounded'
Append-Fixture $log "`"`nlevel=error msg=`"next`"`n";Drain-Fixture $cursor 3000;$summary=Get-LiveLogSummary @($cursor)
Assert-Observation ($summary.warningLines -eq 1 -and $summary.errorLines -eq 1 -and $summary.appendedLines -eq 2 -and $summary.coverageComplete) 'Long record and following record must not duplicate'
Record-Observation 'very long records stay bounded in memory and do not duplicate the following severity'

$log=Fixture-Log 'truncate.log' ('old fixture text'*100);$cursor=New-LiveLogCursor $log
[IO.File]::WriteAllText($log,"level=error msg=`"new`"`n",$utf8);Drain-Fixture $cursor;$summary=Get-LiveLogSummary @($cursor)
Assert-Observation ($summary.logTruncated -and -not $summary.coverageComplete -and $summary.errorLines -eq 1) 'Truncation must invalidate complete log coverage'
Record-Observation 'log truncation remains marked, even after counters resume'

$missing=Join-Path $fixture 'missing.log';$rejected=$false;try{New-LiveLogCursor $missing|Out-Null}catch{$rejected=$true}
Assert-Observation ($rejected -and -not (Test-Path -LiteralPath $missing)) 'Opening a missing log must not create it'
Record-Observation 'missing log read fails without creating or modifying a log'

$samples=@();foreach($at in 0..180){$samples+=Sample-Fixture ($at*5) ($at*0.2) (64MB+($at*1KB))}
$summary=Get-LiveResourceSummary $samples 900 5 4
Assert-Observation ($summary.fullTiming -and $summary.completeWindow -and $summary.cadencePassed -and $summary.resourceThresholdsPassed -and [Math]::Abs($summary.cpuAveragePercent-1) -lt 0.00001 -and $summary.sampleCount -eq 181) 'Full monotonic CPU capacity and timing summary'
Record-Observation '900-second window and 181 samples yield the correct CPU share of machine capacity'

$samples=@((Sample-Fixture 0 0 64MB),(Sample-Fixture 5 1 130MB),(Sample-Fixture 10 2 65MB))
$summary=Get-LiveResourceSummary $samples 10 5 4
Assert-Observation ($summary.privateGrowthMiB -eq 66 -and -not $summary.resourceThresholdsPassed -and $summary.privateLastBytes -eq 65MB) 'Peak memory growth must not use only the last sample'
Record-Observation 'peak memory growth fails its threshold even when final memory falls'

$samples=@((Sample-Fixture 0 0 64MB),(Sample-Fixture 5 1 (64MB+8KB)),(Sample-Fixture 10 2 (64MB+4KB)))
$summary=Get-LiveResourceSummary $samples 10 5 4
Assert-Observation ($summary.privateGrowthMiB -eq 0.0078125) 'Sub-MiB memory growth must not round to an integer'
$summary=Get-LiveResourceSummary $samples 10 5 4 20 0.005
Assert-Observation (-not $summary.resourceThresholdsPassed) 'Fractional memory limit must not falsely pass after rounding'
Record-Observation '8 KiB peak growth retains fractional MiB precision and can fail a smaller configured threshold'

$samples=@((Sample-Fixture 0 0),(Sample-Fixture 5 10),(Sample-Fixture 10 20))
$summary=Get-LiveResourceSummary $samples 10 5 4
Assert-Observation ($summary.cpuAveragePercent -eq 50 -and -not $summary.resourceThresholdsPassed) 'High CPU must fail the default resource threshold'
$summary=Get-LiveResourceSummary $samples 10 5 4 60 64
Assert-Observation $summary.resourceThresholdsPassed 'Explicit CPU threshold must be applied'
$samples=@((Sample-Fixture 0 0 64MB),(Sample-Fixture 5 1 67MB),(Sample-Fixture 10 2 65MB))
$summary=Get-LiveResourceSummary $samples 10 5 4 20 2
Assert-Observation (-not $summary.resourceThresholdsPassed -and $summary.privateGrowthMiB -eq 3) 'Explicit memory threshold must be applied to peak growth'
Record-Observation 'high CPU fails the default limit; explicit CPU and memory thresholds affect the correct metrics'

$samples=@((Sample-Fixture 0 0),(Sample-Fixture 5 1),(Sample-Fixture 10 2))
$summary=Get-LiveResourceSummary $samples 10 5 4
Assert-Observation ($summary.completeWindow -and $summary.cadencePassed -and -not $summary.fullTiming) 'Short diagnostic must not become formal observation'
$summary=Get-LiveResourceSummary $samples 900 5 4
Assert-Observation ($summary.fullTiming -and -not $summary.completeWindow -and -not $summary.cadencePassed) 'Early stop must not become a full window'
Record-Observation 'short diagnostics and early stops cannot be counted as formal completed observation'

$samples=@((Sample-Fixture 0 0),(Sample-Fixture 450 1),(Sample-Fixture 900 2))
$summary=Get-LiveResourceSummary $samples 900 5 4
Assert-Observation ($summary.completeWindow -and -not $summary.cadencePassed -and $summary.maxSampleGapSeconds -eq 450) 'Sparse endpoints must not count as continuous sampling'
Record-Observation 'full span with sparse samples or large gaps fails cadence'

foreach($samples in @(@((Sample-Fixture 0 5),(Sample-Fixture 5 4)),@((Sample-Fixture 5 1),(Sample-Fixture 5 2)))){
 $summary=Get-LiveResourceSummary $samples 10 1 4
 Assert-Observation (-not $summary.metricsValid -and -not $summary.resourceThresholdsPassed) 'Counter reversal or zero span invalidates metrics'
}
Record-Observation 'CPU counter reversal and non-increasing timestamps invalidate metrics'

$summary=Get-LiveResourceSummary @((Sample-Fixture 0 0)) 900 5 4
Assert-Observation ($summary.sampleCount -eq 1 -and -not $summary.metricsValid -and -not $summary.completeWindow -and $null -eq $summary.cpuAveragePercent) 'Insufficient samples must not invent zero CPU'
Record-Observation 'insufficient samples return unmeasured CPU rather than a fabricated zero'

$rejected=0
try{Get-LiveResourceSummary @() 900 5 0|Out-Null}catch{$rejected++}
try{Get-LiveResourceSummary @() 9 5 4|Out-Null}catch{$rejected++}
try{Read-LiveLogDelta @{Path=$missing} 0}catch{$rejected++}
Assert-Observation ($rejected -eq 3 -and -not (Test-Path -LiteralPath $missing)) 'Invalid statistics and log limits must fail before reading or fabricating metrics'
Record-Observation 'invalid CPU count, duration and read limit fail before reads or invented statistics'

$parsed=0
foreach($file in @('observe-live-services.ps1','live-observation-core.ps1','test-live-observation.ps1')){
 $tokens=$null;$errors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot $file),[ref]$tokens,[ref]$errors)
 Assert-Observation (-not $errors) ('PowerShell syntax: '+$file)
 if($file -eq 'observe-live-services.ps1'){
  $unsafe=$ast.FindAll({param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -in @('Start-Process','Stop-Process','Invoke-WebRequest','Invoke-RestMethod','Start-ScheduledTask','Stop-ScheduledTask','Register-ScheduledTask')},$true)
  Assert-Observation ($unsafe.Count -eq 0) 'Observer must not control services or invoke network APIs'
 }
 $parsed++
}
Record-Observation 'three scripts parse, and observer contains no service-control or network commands'
$result=@{completedAt=(Get-Date).ToString('o');passed=$true;tests=$cases.Count;cases=$cases;scriptsParsed=$parsed;syntheticLogFixtures=$true;realQQMessagesSent=$false;servicesChanged=$false}
[IO.File]::WriteAllText((Join-Path $fixture 'result.json'),($result|ConvertTo-Json -Depth 8),$utf8)
$result|ConvertTo-Json -Depth 8
Write-Host ('ARTIFACTS='+$fixture)
