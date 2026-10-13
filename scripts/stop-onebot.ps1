[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$statePath = Join-Path $repoRoot '.local/onebot/runtime/process.json'
if (-not (Test-Path -LiteralPath $statePath)) { Write-Host 'No local NapCat process recorded.'; return }
$state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
$process = Get-Process -Id $state.processId -ErrorAction SilentlyContinue
if ($process) {
    $expectedPath = Join-Path $repoRoot '.local/onebot/napcat-v4.18.28/node.exe'
    if ($process.Path -ne $expectedPath -or $state.executable -ne $expectedPath) {
        throw 'Recorded PID does not match the local NapCat executable; refusing to stop it.'
    }
    # NapCat forks a Node worker. Stop only the descendants of our recorded launcher.
    $snapshot = @(Get-CimInstance Win32_Process)
    $ownedIds = [Collections.Generic.List[int]]::new()
    $ownedIds.Add([int]$process.Id)
    for ($index = 0; $index -lt $ownedIds.Count; $index++) {
        foreach ($child in $snapshot | Where-Object { $_.ParentProcessId -eq $ownedIds[$index] }) {
            if ($child.ExecutablePath -ne $expectedPath) {
                throw 'Unexpected child executable; refusing to stop the process tree.'
            }
            $ownedIds.Add([int]$child.ProcessId)
        }
    }
    # Stop the launcher first so it cannot restart a stopped worker.
    foreach ($ownedId in $ownedIds) { Stop-Process -Id $ownedId -Force -ErrorAction SilentlyContinue }
    Write-Host "Stopped local NapCat launcher and workers ($($ownedIds.Count) processes)."
} else { Write-Host 'Recorded NapCat process has already exited.' }
Remove-Item -LiteralPath $statePath
