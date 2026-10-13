[CmdletBinding()]
param()
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'windows-service-identity.ps1')
$repoRoot=Resolve-LocalServicePath (Split-Path -Parent $PSScriptRoot)
$fixtureRoot=Join-Path $repoRoot ('.local/service-path-test-'+[guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $fixtureRoot|Out-Null
$aliasRoot=Join-Path $fixtureRoot 'alias'
try {
New-Item -ItemType Junction -Path $aliasRoot -Target $repoRoot|Out-Null
if((Resolve-LocalServicePath $aliasRoot) -ne $repoRoot){throw 'Fixture junction did not resolve to this repository.'}
$shell=(Get-Command pwsh.exe).Source
function New-FixtureTask([string]$root,[string]$service){
    return [pscustomobject]@{Actions=@([pscustomobject]@{Execute=$shell;Arguments='-NoProfile -WindowStyle Hidden -File "'+(Join-Path $root 'scripts/service-host.ps1')+'" -Service '+$service;WorkingDirectory=$root})}
}
foreach($service in @('OneBot','ZeroBot')){
    foreach($root in @($repoRoot,$aliasRoot)){
        if(-not (Test-LocalServiceTask (New-FixtureTask $root $service) $repoRoot $service)){throw 'Same-project task alias rejected.'}
    }
}
$foreign=New-FixtureTask (Join-Path $repoRoot 'different-project') 'ZeroBot'
if(Test-LocalServiceTask $foreign $repoRoot 'ZeroBot'){throw 'Foreign project task accepted.'}
$bad=New-FixtureTask $repoRoot 'OneBot'
if(Test-LocalServiceTask $bad $repoRoot 'ZeroBot'){throw 'Wrong service task accepted.'}
$bad=New-FixtureTask $repoRoot 'ZeroBot';$bad.Actions[0].Execute='D:\unrelated\pwsh.exe'
if(Test-LocalServiceTask $bad $repoRoot 'ZeroBot'){throw 'Unverified executable accepted.'}
$bad=New-FixtureTask $repoRoot 'ZeroBot';$bad.Actions[0].Arguments+=' -Command injected'
if(Test-LocalServiceTask $bad $repoRoot 'ZeroBot'){throw 'Extra task arguments accepted.'}
$expected=Join-Path $repoRoot '.local/bin/zerobot.exe'
$timestamp=[datetime]'2026-10-08T00:00:00Z'
$state=[pscustomobject]@{managedBy='Windows Task Scheduler';processId=123;executable=(Join-Path $aliasRoot '.local/bin/zerobot.exe');processStartedAt=$timestamp.ToString('o')}
$process=[pscustomobject]@{Id=123;Path=$expected;StartTime=$timestamp}
if(-not (Test-LocalServiceProcess $state $process $expected)){throw 'Same-process alias rejected.'}
$process.StartTime=$timestamp.AddSeconds(1)
if(Test-LocalServiceProcess $state $process $expected){throw 'PID reuse timestamp accepted.'}
$process.StartTime=$timestamp;$process.Path=Join-Path $repoRoot 'different-project/zerobot.exe'
if(Test-LocalServiceProcess $state $process $expected){throw 'Foreign process accepted.'}
$process.Path=$expected;$process.Id=456
if(Test-LocalServiceProcess $state $process $expected){throw 'Different PID accepted.'}
$files=@('windows-service-identity.ps1','start-local-services.ps1','stop-local-services.ps1','service-host.ps1','test-service-paths.ps1')
foreach($file in $files){
    $tokens=$null;$parseErrors=$null
    [Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot $file),[ref]$tokens,[ref]$parseErrors) | Out-Null
    if($parseErrors){throw ('Script syntax invalid: '+$file)}
}
[pscustomobject]@{completedAt=(Get-Date).ToString('o');passed=$true;canonicalRoot=$repoRoot;aliasRoot=$aliasRoot;taskFixturesPassed=8;processFixturesPassed=4;scriptsParsed=$files.Count;liveProcessesStopped=0} | ConvertTo-Json
} finally {
    if(Test-Path -LiteralPath $aliasRoot){
        $link=Get-Item -LiteralPath $aliasRoot -Force
        if($link.LinkType -ne 'Junction' -or (Resolve-LocalServicePath $aliasRoot) -ne $repoRoot){throw 'Refusing to remove an unexpected fixture path.'}
        # Remove only this verified junction, never recursively remove its target.
        Remove-Item -LiteralPath $aliasRoot -Force
    }
    if(Test-Path -LiteralPath $fixtureRoot){Remove-Item -LiteralPath $fixtureRoot -Force}
}
