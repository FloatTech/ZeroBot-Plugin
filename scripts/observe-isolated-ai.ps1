[CmdletBinding()]
param(
 [ValidatePattern('^[A-Za-z0-9][A-Za-z0-9_.-]*\.exe$')][string]$BinaryName='zerobot.exe',
 [ValidateRange(10,1200)][int]$DurationSeconds=900,
 [ValidateRange(1,30)][int]$IntervalSeconds=5,
 [ValidateRange(60,120)][int]$BatchIntervalSeconds=60,
 [ValidateRange(0,2000)][int]$ModelDelayMilliseconds=150
)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'windows-service-identity.ps1')
. (Join-Path $PSScriptRoot 'live-observation-core.ps1')
$repoRoot=Resolve-LocalServicePath (Split-Path -Parent $PSScriptRoot)
$binaryPath=Join-Path $repoRoot ('.local/bin/'+$BinaryName)
if(-not (Test-Path -LiteralPath $binaryPath -PathType Leaf)){throw 'Selected test binary does not exist.'}
$programHash=(Get-FileHash -LiteralPath $binaryPath -Algorithm SHA256).Hash
$goPath=(Get-Command go -CommandType Application|Select-Object -First 1).Source
$nodePath=(Get-Command node -CommandType Application|Select-Object -First 1).Source
if(-not (Get-Command ConvertFrom-Json).Parameters.ContainsKey('DateKind')){throw 'PowerShell with JSON DateKind support is required.'}
$observationRoot=Join-Path $repoRoot ('.local/ai-load-observation/run-'+(Get-Date -Format 'yyyyMMdd-HHmmss')+'-'+[guid]::NewGuid().ToString('N').Substring(0,8))
New-Item -ItemType Directory -Path $observationRoot|Out-Null
$resourceRoot=Join-Path $observationRoot 'resource'
New-Item -ItemType Directory -Path $resourceRoot|Out-Null
$utf8=[Text.UTF8Encoding]::new($false)
function Save-IsolatedJSON([string]$Name,$Value){[IO.File]::WriteAllText((Join-Path $observationRoot $Name),($Value|ConvertTo-Json -Depth 12),$utf8)}
$result=[ordered]@{startedAt=(Get-Date).ToString('o');programSHA256=$programHash;isolated=$true;requestedDurationSeconds=$DurationSeconds;intervalSeconds=$IntervalSeconds;batchIntervalSeconds=$BatchIntervalSeconds;syntheticModelDelayMilliseconds=$ModelDelayMilliseconds;completed=$false;passed=$false;fullTiming=$DurationSeconds -ge 900;diagnosticChecksPassed=$false;syntheticRequestBatches=0;syntheticModelCalls=0;realQQMessagesSent=$false;realModelCalls=0;dailyServicesChanged=$false;businessAcceptancePassed=$false;stopReason='not-started';instanceStopped=$false}
$helperProcess=$null;$sampledBot=$null;$readLine=$null;$helperErrors=$null;$fixtureRoot=$null;$afterStatus=$null;$beforeStatus=$null;$resourceResult=$null;$cursors=@();$samples=[Collections.Generic.List[object]]::new();$stage='start-fixture';$failed=$false;$identityStable=$true
function Read-IsolatedStatus {
 $limit=[Diagnostics.Stopwatch]::StartNew()
 while($limit.Elapsed.TotalSeconds -lt 60){
  if($readLine.IsCompleted){
   $line=$readLine.GetAwaiter().GetResult()
   if($null -eq $line){throw 'Fixture output ended.'}
   $script:readLine=$helperProcess.StandardOutput.ReadLineAsync()
   if($line.StartsWith('{')){
    if($line.Length -gt 1048576){throw 'Fixture status limit exceeded.'}
    $status=$line|ConvertFrom-Json -DateKind String
    if($status.state -eq 'ready'){return $status}
   }
  }
  Start-Sleep -Milliseconds 20
 }
 throw 'Fixture status timed out.'
}
function Send-IsolatedCommand([string]$Command){$helperProcess.StandardInput.WriteLine($Command);$helperProcess.StandardInput.Flush();Read-IsolatedStatus}
function Test-IsolatedIdentity($Status){
 if($Status.realQQMessagesSent -ne $false -or $Status.realModelCalls -ne 0 -or $Status.processStarts -ne 1){$script:identityStable=$false;throw 'Fixture scope or process count changed.'}
 $identity=$Status.replyTasks.process
 $sampledBot.Refresh()
 if($sampledBot.HasExited -or $identity.processId -ne $sampledBot.Id -or ([datetimeoffset]$identity.processStartedAt).UtcTicks -ne $sampledBot.StartTime.ToUniversalTime().Ticks -or $identity.programSHA256 -ne $programHash -or (Resolve-LocalServicePath $sampledBot.Path) -ne (Join-Path $fixtureRoot 'zerobot-model-observation-acceptance.exe')){$script:identityStable=$false;throw 'Fixture process identity changed.'}
}
function Save-ReplySnapshot([string]$Name,$Stats){
 Save-IsolatedJSON $Name @{format='zerobot-reply-statistics';version=1;exportedAt=[datetime]::UtcNow.ToString('o');scope=@{kind='ai-reply-tasks';latencyScope='process-cumulative';timingBoundary='bot-matcher-entry-to-onebot-send-receipt';businessAcceptancePassed=$false};snapshot=$Stats}
}
Write-Host ('ARTIFACTS='+$observationRoot)
try{
 $startInfo=[Diagnostics.ProcessStartInfo]::new()
 $startInfo.FileName=$goPath;$startInfo.WorkingDirectory=$repoRoot;$startInfo.UseShellExecute=$false;$startInfo.CreateNoWindow=$true;$startInfo.RedirectStandardInput=$true;$startInfo.RedirectStandardOutput=$true;$startInfo.RedirectStandardError=$true
 foreach($argument in @('run','scripts/start-model-observation-acceptance.go','-binary-name',$BinaryName,'-model-delay',($ModelDelayMilliseconds.ToString()+'ms'),'-lifetime','30m')){$startInfo.ArgumentList.Add($argument)}
 $helperProcess=[Diagnostics.Process]::new();$helperProcess.StartInfo=$startInfo
 if(-not $helperProcess.Start()){throw 'Fixture could not be started.'}
 $helperErrors=$helperProcess.StandardError.ReadToEndAsync();$readLine=$helperProcess.StandardOutput.ReadLineAsync()
 $initial=Read-IsolatedStatus
 $stage='verify-fixture'
 $fixtureRoot=Resolve-LocalServicePath $initial.output
 $fixtureParent=(Join-Path $repoRoot '.local/model-observation-acceptance')+[IO.Path]::DirectorySeparatorChar
 if(-not $fixtureRoot.StartsWith($fixtureParent,[StringComparison]::OrdinalIgnoreCase) -or [IO.Path]::GetFileName($fixtureRoot) -notmatch '^run-[0-9]+$'){throw 'Fixture directory is outside the isolated workspace.'}
 $address=[uri]$initial.url
 if($address.Scheme -ne 'http' -or $address.Host -ne '127.0.0.1' -or $address.Port -in @(3010,6099,6700)){throw 'Fixture address is not isolated.'}
 $sampledBot=Get-Process -Id $initial.replyTasks.process.processId
 Test-IsolatedIdentity $initial
 if((Get-FileHash -LiteralPath (Join-Path $fixtureRoot 'zerobot-model-observation-acceptance.exe') -Algorithm SHA256).Hash -ne $programHash){throw 'Copied fixture program differs.'}
 foreach($stream in @('stdout','stderr')){$cursors+=New-LiveLogCursor (Join-Path $fixtureRoot ('start-1-'+$stream+'.log'))}
 $result.fixtureDirectory=$fixtureRoot
 $resourceResult=[ordered]@{startedAt=(Get-Date).ToString('o');processId=$sampledBot.Id;processStartedAt=$sampledBot.StartTime.ToString('o');programSHA256=$programHash;requestedDurationSeconds=$DurationSeconds;intervalSeconds=$IntervalSeconds;logicalCPUs=[Environment]::ProcessorCount;thresholds=@{cpuAverageCapacityPercent=20;privateGrowthMiB=64};completed=$false;passed=$false;identityStable=$true;stopReason='not-started';realQQMessagesSent=$false;modelCalls=0;businessLoadMeasured=$false;businessAcceptancePassed=$false;messageContentsSaved=$false;servicesStartedOrStopped=$false}
 $timer=[Diagnostics.Stopwatch]::StartNew();$nextBatch=0.0;$nextProgress=0.0
 while($true){
  if(Test-Path -LiteralPath (Join-Path $observationRoot 'stop.request')){$result.stopReason='stop-request';break}
  $stage='sample-resource';Test-IsolatedIdentity $initial
  $sample=[ordered]@{at=(Get-Date).ToString('o');atSeconds=$timer.Elapsed.TotalSeconds;cpuSeconds=$sampledBot.TotalProcessorTime.TotalSeconds;privateBytes=$sampledBot.PrivateMemorySize64;workingSetBytes=$sampledBot.WorkingSet64;threadCount=$sampledBot.Threads.Count;handleCount=$sampledBot.HandleCount}
  $samples.Add([pscustomobject]$sample);[IO.File]::AppendAllText((Join-Path $resourceRoot 'samples.jsonl'),(($sample|ConvertTo-Json -Compress)+[Environment]::NewLine),$utf8)
  foreach($cursor in $cursors){Read-LiveLogDelta $cursor}
  $elapsed=$sample.atSeconds-$samples[0].atSeconds
  if($null -eq $beforeStatus){$stage='start-snapshots';$beforeStatus=Send-IsolatedCommand 'status';Test-IsolatedIdentity $beforeStatus;Save-ReplySnapshot 'reply-before.json' $beforeStatus.replyTasks}
  if($null -eq $afterStatus -and $elapsed -ge $DurationSeconds-5){
   $stage='finish-snapshots';$afterStatus=Send-IsolatedCommand 'status';Test-IsolatedIdentity $afterStatus;Save-ReplySnapshot 'reply-after.json' $afterStatus.replyTasks
  }
  if($null -eq $afterStatus -and $elapsed -ge $nextBatch){
   $stage='synthetic-request-batch';$status=Send-IsolatedCommand 'requests';Test-IsolatedIdentity $status
   for($attempt=0;$attempt -lt 30 -and $status.replyTasks.active -ne 0;$attempt++){Start-Sleep -Milliseconds 100;$status=Send-IsolatedCommand 'status';Test-IsolatedIdentity $status}
   $result.syntheticRequestBatches++;$expected=$result.syntheticRequestBatches*3
   if($status.syntheticModelCalls -ne $expected -or $status.modelRequests.completed -ne $expected -or $status.modelRequests.active -ne 0 -or $status.modelRequests.failed -ne 0 -or $status.replyTasks.started -ne $expected -or $status.replyTasks.completed -ne $expected -or $status.replyTasks.active -ne 0 -or $status.replyTasks.replyAccepted -ne $expected -or $status.replyTasks.failed -ne 0 -or $status.replyTasks.missingIngress -ne 0){throw 'Synthetic model and reply counts differ.'}
   $result.syntheticModelCalls=$status.syntheticModelCalls;$nextBatch=$result.syntheticRequestBatches*$BatchIntervalSeconds
  }
  if($elapsed -ge $nextProgress){Save-IsolatedJSON 'progress.json' @{checkedAt=(Get-Date).ToString('o');state='observing';elapsedSeconds=$elapsed;sampleCount=$samples.Count;syntheticRequestBatches=$result.syntheticRequestBatches;syntheticModelCalls=$result.syntheticModelCalls};Write-Host ('OBSERVING seconds={0:N1} samples={1} syntheticCalls={2}' -f $elapsed,$samples.Count,$result.syntheticModelCalls);$nextProgress=$elapsed+60}
  if($elapsed -ge $DurationSeconds){
   $sampledBot.Refresh();$finalSample=[ordered]@{at=(Get-Date).ToString('o');atSeconds=$timer.Elapsed.TotalSeconds;cpuSeconds=$sampledBot.TotalProcessorTime.TotalSeconds;privateBytes=$sampledBot.PrivateMemorySize64;workingSetBytes=$sampledBot.WorkingSet64;threadCount=$sampledBot.Threads.Count;handleCount=$sampledBot.HandleCount}
   $samples.Add([pscustomobject]$finalSample);[IO.File]::AppendAllText((Join-Path $resourceRoot 'samples.jsonl'),(($finalSample|ConvertTo-Json -Compress)+[Environment]::NewLine),$utf8)
   $result.completed=$true;$result.stopReason='duration-complete';break
  }
  Start-Sleep -Milliseconds ([int][Math]::Max(1,[Math]::Min($IntervalSeconds*1000,($DurationSeconds-$elapsed)*1000)))
 }
 $timer.Stop()
}catch{$failed=$true;$result.stopReason='observation-error';$result.failureStage=$stage;$result.failureLine=$_.InvocationInfo.ScriptLineNumber;$result.failureType=$_.Exception.GetType().FullName}
finally{
 try{if($resourceResult){
  foreach($cursor in $cursors){for($attempt=0;$attempt -lt 4;$attempt++){Read-LiveLogDelta $cursor;if($cursor.UnreadBytes -eq 0){break}}}
  $logs=Get-LiveLogSummary $cursors;$resourceResult.logCounts=$logs;$resourceResult.noNewErrorLines=($logs.errorLines+$logs.fatalLines+$logs.panicLines) -eq 0
  $resourceResult.resources=Get-LiveResourceSummary $samples.ToArray() $DurationSeconds $IntervalSeconds $resourceResult.logicalCPUs
  $resourceResult.completed=$result.completed;$resourceResult.stopReason=$result.stopReason
  $resourceResult.programUnchanged=(Get-FileHash -LiteralPath (Join-Path $fixtureRoot 'zerobot-model-observation-acceptance.exe')).Hash -eq $programHash
  $resourceResult.identityStable=$identityStable -and $result.failureStage -ne 'verify-fixture'
  $resourceResult.passed=$resourceResult.completed -and -not $failed -and $resourceResult.identityStable -and $resourceResult.programUnchanged -and $resourceResult.resources.fullTiming -and $resourceResult.resources.completeWindow -and $resourceResult.resources.cadencePassed -and $resourceResult.resources.resourceThresholdsPassed -and $resourceResult.noNewErrorLines -and $logs.coverageComplete
  $resourceResult.completedAt=(Get-Date).ToString('o');Save-IsolatedJSON 'resource/result.json' $resourceResult
 }}catch{$failed=$true;$result.stopReason='observation-error';$result.failureStage='final-resource';$result.failureType=$_.Exception.GetType().FullName}
 if($helperProcess){
  try{if(-not $helperProcess.HasExited){$helperProcess.StandardInput.WriteLine('stop');$helperProcess.StandardInput.Close();if(-not $helperProcess.WaitForExit(15000)){$helperProcess.Kill($true);$helperProcess.WaitForExit(5000)|Out-Null}};$result.instanceStopped=$helperProcess.HasExited}catch{$result.cleanupFailed=$true;$failed=$true}
  $helperProcess.Dispose()
 }
 if($sampledBot){$sampledBot.Dispose()};foreach($cursor in $cursors){$cursor.Prefix=''}
}
if($result.completed -and -not $failed){
 try{
  $stage='offline-evidence'
  $beforeModel=$beforeStatus.modelRequests;$afterModel=$afterStatus.modelRequests
  $delta=@{};foreach($key in @('started','completed','failed','incomplete')){$delta[$key]=$afterModel.$key-$beforeModel.$key};$delta.activeAtStart=$beforeModel.active;$delta.activeAtEnd=$afterModel.active
  $rows=@();foreach($kind in @('gpt','chat','agent','vision')){$a=$beforeModel.categories|Where-Object kind -eq $kind;$b=$afterModel.categories|Where-Object kind -eq $kind;$row=@{kind=$kind;activeAtStart=$a.active;activeAtEnd=$b.active};foreach($key in @('started','completed','failed','incomplete')){$row[$key]=$b.$key-$a.$key};$rows+=$row}
  $delta.categories=$rows;$delta.durationSeconds=([datetimeoffset]$afterModel.checkedAt).ToUnixTimeMilliseconds()/1000-([datetimeoffset]$beforeModel.checkedAt).ToUnixTimeMilliseconds()/1000
  $delta.durationSeconds=[math]::Round($delta.durationSeconds,3)
  Save-IsolatedJSON 'model-report.json' @{format='zerobot-model-observation';version=2;exportedAt=[datetime]::UtcNow.ToString('o');scope=@{kind='model-http-exchanges';latencyScope='process-cumulative';businessAcceptancePassed=$false};before=$beforeModel;after=$afterModel;window=$delta}
  & $nodePath (Join-Path $PSScriptRoot 'check-observation-evidence.mjs') --resource $resourceRoot --model (Join-Path $observationRoot 'model-report.json') --output (Join-Path $observationRoot 'model-check.json') > (Join-Path $observationRoot 'model-check.txt')
  $modelExit=$LASTEXITCODE
  & $nodePath (Join-Path $PSScriptRoot 'check-reply-evidence.mjs') --before (Join-Path $observationRoot 'reply-before.json') --after (Join-Path $observationRoot 'reply-after.json') --resource $resourceRoot --output (Join-Path $observationRoot 'reply-check.json') > (Join-Path $observationRoot 'reply-check.txt')
  $replyExit=$LASTEXITCODE
  $modelCheck=Get-Content -LiteralPath (Join-Path $observationRoot 'model-check.json') -Raw|ConvertFrom-Json;$replyCheck=Get-Content -LiteralPath (Join-Path $observationRoot 'reply-check.json') -Raw|ConvertFrom-Json
  $result.modelEvidenceExitCode=$modelExit;$result.replyEvidenceExitCode=$replyExit
  $result.diagnosticChecksPassed=$modelCheck.evidenceValid -and $replyCheck.evidenceValid -and $modelCheck.sameProcessVerified -and $replyCheck.sameProcessVerified -and $modelCheck.temporalCoverage -eq 'inside' -and $replyCheck.temporalCoverage -eq 'inside' -and $modelCheck.model.window.completed -eq $result.syntheticModelCalls -and $replyCheck.window.completed -eq $result.syntheticModelCalls -and $result.instanceStopped
  $result.passed=$result.fullTiming -and $result.diagnosticChecksPassed -and $modelExit -eq 0 -and $replyExit -eq 0
 }catch{$failed=$true;$result.failureStage=$stage;$result.stopReason='evidence-error';$result.failureLine=$_.InvocationInfo.ScriptLineNumber;$result.failureType=$_.Exception.GetType().FullName}
}
$result.completedAt=(Get-Date).ToString('o');Save-IsolatedJSON 'result.json' $result;Save-IsolatedJSON 'progress.json' @{state='finished';completedAt=$result.completedAt;completed=$result.completed;passed=$result.passed;stopReason=$result.stopReason;syntheticModelCalls=$result.syntheticModelCalls}
Write-Host ('OBSERVATION_COMPLETE completed={0} fullTiming={1} passed={2}' -f $result.completed,$result.fullTiming,$result.passed)
if($failed){throw 'Isolated AI observation failed; inspect the recorded stage. Daily services were not changed.'}
