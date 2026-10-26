#requires -Version 5.1
<#
.SYNOPSIS
Downloads pinned Windows x64 engines into the local build cache.
Engine assets and Go caches stay in the repository's .cache directory.
#>
[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'This script requires Windows.' }
$architecture = $env:PROCESSOR_ARCHITECTURE
if ($env:PROCESSOR_ARCHITEW6432) { $architecture = $env:PROCESSOR_ARCHITEW6432 }
if ($architecture -ne 'AMD64') { throw 'Only Windows x64 is supported.' }

$sourceRoot = Split-Path -Parent $PSScriptRoot
$cacheRoot = Join-Path $sourceRoot '.cache'
$CacheDir = Join-Path $cacheRoot 'engines/windows-amd64'
$binDir = Join-Path $CacheDir 'bin'
$staging = Join-Path $CacheDir ('.download-' + [guid]::NewGuid().ToString('N'))
[void](New-Item -ItemType Directory -Path $binDir -Force)
[void](New-Item -ItemType Directory -Path $staging)
$manifestTool = Join-Path $staging 'engine-manifest.exe'

function Get-ManifestField([string]$Field) {
    $value = & $manifestTool get $Field
    if ($LASTEXITCODE -ne 0) { throw "Cannot read manifest field: $Field" }
    return $value.Trim()
}

$savedEnvironment = @{}
foreach ($name in @('GOCACHE', 'GOPATH', 'GOMODCACHE')) {
    $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
try {
    $env:GOCACHE = Join-Path $cacheRoot 'go-build-windows'
    $env:GOPATH = Join-Path $cacheRoot 'go-windows'
    $env:GOMODCACHE = Join-Path $env:GOPATH 'pkg/mod'
    # Use the same embedded YAML as the controller and Linux downloader.
    Push-Location -LiteralPath $sourceRoot
    try {
        & go build -o $manifestTool ./cmd/engine-manifest
        if ($LASTEXITCODE -ne 0) { throw 'Cannot build engine-manifest (Go is required).' }
    } finally { Pop-Location }

    foreach ($engine in @('free-turn-proxy', 'mihomo')) {
        $repo = Get-ManifestField "$engine.repository"
        $version = Get-ManifestField "$engine.version"
        $asset = Get-ManifestField "$engine.platforms.windows-amd64.asset"
        $checksum = Get-ManifestField "$engine.platforms.windows-amd64.sha256"
        $compression = Get-ManifestField "$engine.platforms.windows-amd64.compression"
        $assetDir = Join-Path $CacheDir "$repo/$version"
        [void](New-Item -ItemType Directory -Path $assetDir -Force)
        $cached = Join-Path $assetDir $asset
        $valid = (Test-Path -LiteralPath $cached -PathType Leaf) -and
            ((Get-FileHash -LiteralPath $cached -Algorithm SHA256).Hash -ieq $checksum)
        if ($valid) {
            Write-Host "Cache: $repo $version / $asset"
        } else {
            Write-Host "Downloading: $repo $version / $asset"
            $download = Join-Path $staging $asset
            & curl.exe --fail --location --show-error --silent --retry 3 --connect-timeout 30 --max-time 300 `
                --output $download "https://github.com/$repo/releases/download/$version/$asset"
            if ($LASTEXITCODE -ne 0) { throw "Download failed: $asset" }
            if ((Get-FileHash -LiteralPath $download -Algorithm SHA256).Hash -ine $checksum) {
                throw "SHA-256 mismatch: $asset"
            }
            Move-Item -LiteralPath $download -Destination $cached -Force
        }

        $prepared = Join-Path $staging "$engine.exe"
        switch ($compression) {
            'none' { Copy-Item -LiteralPath $cached -Destination $prepared }
            'zip' {
                $unpacked = Join-Path $staging $engine
                Expand-Archive -LiteralPath $cached -DestinationPath $unpacked
                $executables = @(Get-ChildItem -LiteralPath $unpacked -Recurse -File -Filter '*.exe')
                if ($executables.Count -ne 1) { throw "Expected one executable in $asset" }
                Copy-Item -LiteralPath $executables[0].FullName -Destination $prepared
            }
            'gzip' {
                $inputFile = [IO.File]::OpenRead($cached)
                try {
                    $gzip = [IO.Compression.GZipStream]::new($inputFile, [IO.Compression.CompressionMode]::Decompress)
                    try {
                        $outputFile = [IO.File]::Create($prepared)
                        try { $gzip.CopyTo($outputFile) } finally { $outputFile.Dispose() }
                    } finally { $gzip.Dispose() }
                } finally { $inputFile.Dispose() }
            }
            default { throw "Unsupported compression: $compression" }
        }
    }

    # Publish only after both downloads and decompression have succeeded.
    foreach ($engine in @('free-turn-proxy', 'mihomo')) {
        Copy-Item -LiteralPath (Join-Path $staging "$engine.exe") -Destination (Join-Path $binDir "$engine.exe") -Force
    }
    Write-Host "Ready: $binDir"
} finally {
    foreach ($name in $savedEnvironment.Keys) {
        if ($null -eq $savedEnvironment[$name]) {
            Remove-Item -LiteralPath "Env:$name" -ErrorAction SilentlyContinue
        } else {
            [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
        }
    }
    # Never recursively remove anything outside this invocation's staging directory.
    $stagingFull = [IO.Path]::GetFullPath($staging)
    $cachePrefix = $CacheDir.TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
    if (-not $stagingFull.StartsWith($cachePrefix, [StringComparison]::OrdinalIgnoreCase) -or
        [IO.Path]::GetFileName($stagingFull) -notmatch '^\.download-[0-9a-f]{32}$') {
        throw 'Unexpected staging directory; refusing cleanup.'
    }
    if (Test-Path -LiteralPath $stagingFull) { Remove-Item -LiteralPath $stagingFull -Recurse -Force }
}
