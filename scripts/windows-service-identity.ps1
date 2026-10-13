# Resolve parent directory junctions as well as the final file's link.
function Resolve-LocalServicePath {
    param([Parameter(Mandatory=$true)][string]$Path,[int]$Depth=0)
    if($Depth -gt 16){throw 'Too many service path links.'}
    $full=[IO.Path]::GetFullPath($Path)
    $root=[IO.Path]::GetPathRoot($full)
    $current=$root
    foreach($part in ($full.Substring($root.Length) -split '[\\/]' | Where-Object {$_ -ne ''})){
        $current=Join-Path $current $part
        if(Test-Path -LiteralPath $current){
            $item=Get-Item -LiteralPath $current -Force -ErrorAction Stop
            if($item.Attributes -band [IO.FileAttributes]::ReparsePoint){
                $targets=@($item.Target)
                if($targets.Count -ne 1 -or -not $targets[0]){throw 'Service path link target cannot be verified.'}
                $target=[string]$targets[0]
                if(-not [IO.Path]::IsPathRooted($target)){$target=Join-Path (Split-Path -Parent $current) $target}
                $current=Resolve-LocalServicePath -Path $target -Depth ($Depth+1)
            }
        }
    }
    return $current
}

function Test-LocalServiceTask {
    param($Task,[string]$RepoRoot,[string]$Service)
    $actions=@($Task.Actions)
    if($actions.Count -ne 1){return $false}
    $action=$actions[0]
    if($action.Arguments -notmatch '^\-NoProfile \-WindowStyle Hidden \-File "([^"]+)" \-Service (OneBot|ZeroBot)$'){return $false}
    $script=$Matches[1];$kind=$Matches[2]
    if($kind -ne $Service){return $false}
    $pwsh=(Get-Command pwsh.exe -ErrorAction SilentlyContinue).Source
    $shells=@((Join-Path $env:WINDIR 'System32/WindowsPowerShell/v1.0/powershell.exe'))
    if($pwsh){$shells+=$pwsh}
    $shells=@($shells | ForEach-Object {Resolve-LocalServicePath $_})
    if((Resolve-LocalServicePath $action.Execute) -notin $shells){return $false}
    return ((Resolve-LocalServicePath $script) -eq (Join-Path $RepoRoot 'scripts/service-host.ps1') -and
        $action.WorkingDirectory -and (Resolve-LocalServicePath $action.WorkingDirectory) -eq $RepoRoot)
}

function Test-LocalServiceProcess {
    param($State,$Process,[string]$ExpectedExecutable)
    if($State.managedBy -ne 'Windows Task Scheduler' -or $State.processId -ne $Process.Id){return $false}
    if(-not $Process.Path -or (Resolve-LocalServicePath $State.executable) -ne $ExpectedExecutable -or
        (Resolve-LocalServicePath $Process.Path) -ne $ExpectedExecutable){return $false}
    if($State.processStartedAt){return $Process.StartTime.ToUniversalTime() -eq ([datetime]$State.processStartedAt).ToUniversalTime()}
    return [math]::Abs(($Process.StartTime.ToUniversalTime()-([datetime]$State.startedAt).ToUniversalTime()).TotalSeconds) -le 5
}
