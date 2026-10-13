# Pure statistics and incremental log counters. Never serialize the log cursor or its prefix.
function New-LiveLogCursor([string]$Path) {
    $stream=[IO.File]::Open($Path,'Open','Read',[IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete)
    try {
        return @{Path=$Path;Position=$stream.Length;Prefix='';LongLine=$false;Truncated=$false;BytesRead=0L;Lines=0L;Warnings=0L;Errors=0L;Fatals=0L;Panics=0L;Other=0L;UnreadBytes=0L}
    } finally {$stream.Dispose()}
}
function Read-LiveLogDelta([hashtable]$Cursor,[int]$MaxBytes=262144) {
    if($MaxBytes -lt 1 -or $MaxBytes -gt 1048576){throw 'Invalid log read limit.'}
    $stream=[IO.File]::Open($Cursor.Path,'Open','Read',[IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete)
    try {
        if($stream.Length -lt $Cursor.Position){$Cursor.Truncated=$true;$Cursor.Position=0L;$Cursor.Prefix='';$Cursor.LongLine=$false}
        $stream.Position=$Cursor.Position
        $size=[int][Math]::Min([long]$MaxBytes,[Math]::Max([long]0,$stream.Length-$Cursor.Position))
        if($size -gt 0){
            $buffer=[byte[]]::new($size)
            $read=$stream.Read($buffer,0,$size)
            $Cursor.Position+=$read;$Cursor.BytesRead+=$read
            $text=[Text.Encoding]::UTF8.GetString($buffer,0,$read)
            $pieces=$text.Split([char]10)
            for($index=0;$index -lt $pieces.Length;$index++){
                $piece=$pieces[$index];$room=4096-$Cursor.Prefix.Length
                if($room -gt 0){$Cursor.Prefix+=$piece.Substring(0,[Math]::Min($room,$piece.Length))}
                if($piece.Length -gt $room){$Cursor.LongLine=$true}
                if($index -lt $pieces.Length-1){
                    $prefix=$Cursor.Prefix -replace '\x1b\[[0-9;]*m',''
                    $Cursor.Lines++
                    if($prefix -match '^(?:time=(?:"[^"]*"|\S+)\s+)?level=(warning|warn|error|fatal|panic)\b'){
                        switch($Matches[1]){warning {$Cursor.Warnings++} warn {$Cursor.Warnings++} error {$Cursor.Errors++} fatal {$Cursor.Fatals++} panic {$Cursor.Panics++}}
                    }elseif($prefix -match '^panic:'){$Cursor.Panics++}else{$Cursor.Other++}
                    $Cursor.Prefix='';$Cursor.LongLine=$false
                }
            }
        }
        $Cursor.UnreadBytes=[Math]::Max(0L,$stream.Length-$Cursor.Position)
    }finally{$stream.Dispose()}
}
function Get-LiveLogSummary([hashtable[]]$Cursors) {
    $value=[ordered]@{appendedLines=0L;warningLines=0L;errorLines=0L;fatalLines=0L;panicLines=0L;otherLines=0L;bytesRead=0L;unreadBytes=0L;partialLines=0;logTruncated=$false;coverageComplete=$true}
    foreach($cursor in $Cursors){
        $value.appendedLines+=$cursor.Lines;$value.warningLines+=$cursor.Warnings;$value.errorLines+=$cursor.Errors;$value.fatalLines+=$cursor.Fatals;$value.panicLines+=$cursor.Panics;$value.otherLines+=$cursor.Other;$value.bytesRead+=$cursor.BytesRead;$value.unreadBytes+=$cursor.UnreadBytes
        if($cursor.Prefix.Length -or $cursor.LongLine){$value.partialLines++}
        if($cursor.Truncated){$value.logTruncated=$true}
    }
    $value.coverageComplete= -not $value.logTruncated -and $value.unreadBytes -eq 0 -and $value.partialLines -eq 0
    return $value
}
function Get-LiveResourceSummary([object[]]$Samples,[int]$DurationSeconds,[int]$IntervalSeconds,[int]$LogicalCPUs,[double]$MaxCpuPercent=20,[double]$MaxPrivateGrowthMiB=64) {
    if($DurationSeconds -lt 10 -or $IntervalSeconds -lt 1 -or $LogicalCPUs -lt 1){throw 'Invalid resource summary parameters.'}
    $value=[ordered]@{sampleCount=$Samples.Count;sampleSpanSeconds=0.0;maxSampleGapSeconds=0.0;cpuAveragePercent=$null;cpuPeakSamplePercent=$null;privateFirstBytes=$null;privateMaxBytes=$null;privateLastBytes=$null;privateGrowthMiB=$null;workingSetMaxBytes=$null;threadMax=$null;handleMax=$null;metricsValid=$false;completeWindow=$false;cadencePassed=$false;resourceThresholdsPassed=$false;fullTiming=$DurationSeconds -ge 900}
    if($Samples.Count -lt 2){return $value}
    $first=$Samples[0];$last=$Samples[-1];$span=$last.atSeconds-$first.atSeconds
    if($span -le 0){return $value}
    $gaps=@();$cpu=@();$valid=$true
    for($index=1;$index -lt $Samples.Count;$index++){
        $gap=$Samples[$index].atSeconds-$Samples[$index-1].atSeconds
        $delta=$Samples[$index].cpuSeconds-$Samples[$index-1].cpuSeconds
        if($gap -le 0 -or $delta -lt 0){$valid=$false;continue}
        $gaps+=$gap;$cpu+=($delta/$gap/$LogicalCPUs*100)
    }
    if(-not $valid -or $gaps.Count -ne $Samples.Count-1){return $value}
    $value.sampleSpanSeconds=$span;$value.maxSampleGapSeconds=($gaps|Measure-Object -Maximum).Maximum
    $value.cpuAveragePercent=($last.cpuSeconds-$first.cpuSeconds)/$span/$LogicalCPUs*100
    $value.cpuPeakSamplePercent=($cpu|Measure-Object -Maximum).Maximum
    $value.privateFirstBytes=$first.privateBytes;$value.privateLastBytes=$last.privateBytes
    $value.privateMaxBytes=($Samples.privateBytes|Measure-Object -Maximum).Maximum
    $value.privateGrowthMiB=[Math]::Max([double]0,[double](($value.privateMaxBytes-$first.privateBytes)/1MB))
    $value.workingSetMaxBytes=($Samples.workingSetBytes|Measure-Object -Maximum).Maximum
    $value.threadMax=($Samples.threadCount|Measure-Object -Maximum).Maximum;$value.handleMax=($Samples.handleCount|Measure-Object -Maximum).Maximum
    $value.metricsValid=$true;$value.completeWindow=$span -ge $DurationSeconds
    $value.cadencePassed=$value.maxSampleGapSeconds -le (2*$IntervalSeconds+0.5) -and $Samples.Count -ge [Math]::Ceiling($DurationSeconds/$IntervalSeconds*0.8)+1
    $value.resourceThresholdsPassed=$value.cpuAveragePercent -le $MaxCpuPercent -and $value.privateGrowthMiB -le $MaxPrivateGrowthMiB
    return $value
}
