[CmdletBinding()]
param([string]$QQAppPath)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$version = 'v4.18.28'
$assetName = 'NapCat.Shell.Windows.Node.zip'
$expectedHash = 'fb64fa3b036ad2df1a5d7c204c482694c20e4b763978c8a4968fd3474c05b4a8'
$downloadRoot = Join-Path $repoRoot '.local/downloads'
$archivePath = Join-Path $downloadRoot $assetName
$bundleRoot = Join-Path $repoRoot '.local/onebot/napcat-v4.18.28'
New-Item -ItemType Directory -Path $downloadRoot -Force | Out-Null
if (-not (Test-Path -LiteralPath $archivePath)) {
    & gh release download $version --repo NapNeko/NapCatQQ --pattern $assetName --dir $downloadRoot
    if ($LASTEXITCODE -ne 0) { throw 'Official NapCat download failed.' }
}
$actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $archivePath).Hash.ToLowerInvariant()
if ($actualHash -ne $expectedHash) { throw 'NapCat archive checksum mismatch; refusing to extract.' }
if (-not (Test-Path -LiteralPath (Join-Path $bundleRoot 'node.exe')) -or -not (Test-Path -LiteralPath (Join-Path $bundleRoot 'wrapper.node'))) {
    Expand-Archive -LiteralPath $archivePath -DestinationPath $bundleRoot -Force
}
if (-not $QQAppPath) {
    $qqRoot = Join-Path $env:ProgramFiles 'Tencent/QQNT'
    $qqVersions = Get-Content -LiteralPath (Join-Path $qqRoot 'versions/config.json') -Raw | ConvertFrom-Json
    $QQAppPath = Join-Path $qqRoot ('versions/' + $qqVersions.curVersion + '/resources/app')
}
$nativeFiles = @()
foreach ($dllName in @('crypto.dll','ssl.dll')) {
    $sourcePath = Join-Path $QQAppPath $dllName
    $destinationPath = Join-Path $bundleRoot $dllName
    if (-not (Test-Path -LiteralPath $destinationPath)) {
        if (-not (Test-Path -LiteralPath $sourcePath)) { throw "Required QQ runtime library is missing: $dllName" }
        Copy-Item -LiteralPath $sourcePath -Destination $destinationPath
    }
    $nativeFiles += [ordered]@{name=$dllName;source=$sourcePath;sha256=(Get-FileHash -Algorithm SHA256 -LiteralPath $destinationPath).Hash.ToLowerInvariant()}
}
$manifest = [ordered]@{
    napcatVersion = $version; bundledQQVersion = '9.9.31-49738'; bundledNodeVersion = '22.11.0'
    releaseURL = 'https://github.com/NapNeko/NapCatQQ/releases/tag/v4.18.28'
    archive = $assetName; archiveSHA256 = $actualHash; addedNativeLibraries = $nativeFiles
}
[IO.File]::WriteAllText((Join-Path $repoRoot '.local/onebot/install-manifest.json'), ($manifest | ConvertTo-Json -Depth 8), [Text.UTF8Encoding]::new($false))
Write-Host "NapCat $version installed locally; official archive SHA256 verified."
& (Join-Path $PSScriptRoot 'setup-onebot.ps1')
