$ErrorActionPreference='Stop'
$repoRoot=Split-Path -Parent $PSScriptRoot
$entry=Join-Path $PSScriptRoot 'observe-isolated-ai.ps1'
$parent=Join-Path $repoRoot '.local/ai-load-observation'
$before=@(Get-ChildItem -LiteralPath $parent -Directory -ErrorAction SilentlyContinue).Count
$cases=@(
 @{name='duration below diagnostic limit';parameters=@{DurationSeconds=9}},
 @{name='duration beyond workload bound';parameters=@{DurationSeconds=1201}},
 @{name='zero resource interval';parameters=@{IntervalSeconds=0}},
 @{name='resource interval beyond bound';parameters=@{IntervalSeconds=31}},
 @{name='batch interval below safe limit';parameters=@{BatchIntervalSeconds=59}},
 @{name='batch interval beyond bound';parameters=@{BatchIntervalSeconds=121}},
 @{name='negative synthetic model delay';parameters=@{ModelDelayMilliseconds=-1}},
 @{name='synthetic model delay beyond bound';parameters=@{ModelDelayMilliseconds=2001}},
 @{name='parent directory binary traversal';parameters=@{BinaryName='../zerobot.exe'}},
 @{name='absolute binary path';parameters=@{BinaryName='D:\zerobot.exe'}},
 @{name='non-executable binary name';parameters=@{BinaryName='zerobot.json'}},
 @{name='command-like binary name';parameters=@{BinaryName='zerobot.exe;test'}},
 @{name='missing binary fails before output or process startup';parameters=@{BinaryName='missing-isolated-ai-test-9a3871.exe'}}
)
$passed=@()
foreach($case in $cases){
 $rejected=$false;$parameters=$case.parameters
 try{& $entry @parameters *> $null}catch{$rejected=$true}
 if(-not $rejected){throw ('Unsafe test parameter was accepted: '+$case.name)}
 $after=@(Get-ChildItem -LiteralPath $parent -Directory -ErrorAction SilentlyContinue).Count
 if($after -ne $before){throw 'Rejected parameters created an observation directory.'}
 $passed+=$case.name
}
$tokens=$null;$errors=$null
$tree=[Management.Automation.Language.Parser]::ParseFile($entry,[ref]$tokens,[ref]$errors)
if($errors.Count){throw 'Observer syntax is invalid.'}
$forbidden=$tree.FindAll({param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -in @('Start-ScheduledTask','Stop-ScheduledTask','Register-ScheduledTask','Stop-Process','Invoke-WebRequest','Invoke-RestMethod')},$true)
if($forbidden.Count){throw 'Observer unexpectedly contains daily service or direct network commands.'}
$passed+='script parses and has no daily service-control or direct HTTP commands'
$result=@{completedAt=(Get-Date).ToString('o');passed=$true;tests=$passed.Count;cases=$passed;rejectedArgumentsCreatedDirectories=$false;servicesChanged=$false;realQQMessagesSent=$false;realModelCalls=0}
$result|ConvertTo-Json -Depth 4|Set-Content (Join-Path $repoRoot '.local/isolated-ai-parameter-tests.json') -Encoding utf8
$result|ConvertTo-Json -Depth 4
