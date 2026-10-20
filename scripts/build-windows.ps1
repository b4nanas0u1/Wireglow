#requires -Version 5.1
<#
.SYNOPSIS
Temporary Windows x64 equivalent of build-linux.sh and ui/build-linux.sh.
Replace these platform scripts with a shared cross-platform build system later.
#>
[CmdletBinding()]
param([switch]$Portable)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'This script requires Windows.' }
$architecture = $env:PROCESSOR_ARCHITECTURE
if ($env:PROCESSOR_ARCHITEW6432) { $architecture = $env:PROCESSOR_ARCHITEW6432 }
if ($architecture -ne 'AMD64') { throw 'Only Windows x64 is supported.' }

$scriptsRoot = $PSScriptRoot
$sourceRoot = Split-Path -Parent $scriptsRoot
$cacheRoot = Join-Path $sourceRoot '.cache'
$cacheBin = Join-Path $cacheRoot 'bin'
$distRoot = Join-Path $sourceRoot 'dist'
$variant = if ($Portable) { 'windows-x64-portable' } else { 'windows-x64' }
$destination = Join-Path $distRoot $variant
# Portable output may contain user data. Refuse before downloading or building.
if ($Portable -and (Test-Path -LiteralPath $destination)) {
    throw "Portable output already exists: $destination. Move it with its data before building."
}
$buildId = [guid]::NewGuid().ToString('N')
$staging = Join-Path $distRoot ".$variant.stage.$buildId"
$previous = Join-Path $distRoot ".$variant.previous.$buildId"
$manifestTool = Join-Path $cacheBin 'engine-manifest.exe'
$savedEnvironment = @{}
foreach ($name in @('GOCACHE', 'GOPATH', 'GOMODCACHE', 'ANALYZER_STATE_LOCATION_OVERRIDE', 'GOOS', 'GOARCH', 'PUB_CACHE', 'CI', 'FLUTTER_SUPPRESS_ANALYTICS')) {
    $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

function Invoke-Checked([string]$Program, [string[]]$Arguments) {
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed with exit code $LASTEXITCODE" }
}

function Get-ManifestField([string]$Field) {
    $value = & $manifestTool get $Field
    if ($LASTEXITCODE -ne 0) { throw "Cannot read manifest field: $Field" }
    return $value.Trim()
}

function Copy-VCRuntime([string]$CMakeCache, [string]$Output) {
    # Use the Visual Studio installation actually selected by Flutter/CMake.
    # Never copy runtime DLLs from System32 or an unrelated installed toolchain.
    $cache = Get-Content -LiteralPath $CMakeCache -Raw
    $match = [regex]::Match($cache, '(?m)^CMAKE_GENERATOR_INSTANCE:[^=]+=(.+)$')
    if (-not $match.Success) { throw 'Cannot identify the Visual Studio instance from CMakeCache.txt.' }
    $visualStudio = $match.Groups[1].Value.Trim() -replace ',version=[^,]+$', ''
    $versionFile = Join-Path $visualStudio 'VC/Auxiliary/Build/Microsoft.VCRedistVersion.default.txt'
    $version = (Get-Content -LiteralPath $versionFile -Raw).Trim()
    if ($version -notmatch '^\d+\.\d+\.\d+(\.\d+)?$') { throw 'Invalid Visual C++ redistributable version.' }
    $redist = Join-Path $visualStudio "VC/Redist/MSVC/$version/x64"
    $crt = @(Get-ChildItem -LiteralPath $redist -Directory -Filter 'Microsoft.VC*.CRT')
    if ($crt.Count -ne 1) { throw "Expected one x64 Visual C++ CRT directory in $redist" }
    $names = @('msvcp140.dll', 'vcruntime140.dll', 'vcruntime140_1.dll')
    # Fail before copying/publishing if the selected redistributable is incomplete.
    foreach ($name in $names) {
        if (-not (Test-Path -LiteralPath (Join-Path $crt[0].FullName $name) -PathType Leaf)) {
            throw "Missing Visual C++ runtime DLL: $name"
        }
    }
    foreach ($name in $names) {
        Copy-Item -LiteralPath (Join-Path $crt[0].FullName $name) -Destination $Output
    }
    Write-Host "App-local Visual C++ runtime: $version (x64)"
}

function Assert-DistPath([string]$Path) {
    $full = [IO.Path]::GetFullPath($Path)
    if ([IO.Path]::GetDirectoryName($full) -ine $distRoot -or
        [IO.Path]::GetFileName($full) -notmatch '^(windows-x64(-portable)?|\.windows-x64(-portable)?\.(stage|previous)\.[0-9a-f]{32})$') {
        throw "Unexpected output path: $Path"
    }
}

function Remove-TemporaryOutput([string]$Path) {
    Assert-DistPath $Path
    if ([IO.Path]::GetFileName($Path) -in @('windows-x64', 'windows-x64-portable')) { throw 'Refusing to delete published output.' }
    if (Test-Path -LiteralPath $Path) { Remove-Item -LiteralPath $Path -Recurse -Force }
}

try {
    $env:GOCACHE = Join-Path $cacheRoot 'go-build-windows'
    $env:GOPATH = Join-Path $cacheRoot 'go-windows'
    $env:GOMODCACHE = Join-Path $env:GOPATH 'pkg/mod'
    $env:ANALYZER_STATE_LOCATION_OVERRIDE = Join-Path $cacheRoot 'dart-analyzer-windows'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:PUB_CACHE = Join-Path $cacheRoot 'pub-windows'
    $env:CI = 'true'
    $env:FLUTTER_SUPPRESS_ANALYTICS = 'true'
    [void](New-Item -ItemType Directory -Path $cacheBin -Force)

    $flutter = Join-Path $cacheRoot 'flutter-windows/flutter/bin/flutter.bat'
    if (-not (Test-Path -LiteralPath $flutter -PathType Leaf)) {
        $flutter = (Get-Command flutter -ErrorAction Stop).Source
    }
    [void](Get-Command go -ErrorAction Stop)
    & (Join-Path $scriptsRoot 'download-windows-engines.ps1')

    Push-Location -LiteralPath $sourceRoot
    try {
        Invoke-Checked 'go' @('build', '-o', (Join-Path $cacheBin 'controller.exe'), './cmd/controller')
        Invoke-Checked 'go' @('build', '-o', $manifestTool, './cmd/engine-manifest')
        $commit = ''
        $tag = ''
        if (Get-Command git -ErrorAction SilentlyContinue) {
            $commit = & git rev-parse --short HEAD 2>$null
            if ($LASTEXITCODE -ne 0) { $commit = '' }
            if ($commit) { $tag = & git tag --points-at HEAD | Select-Object -First 1 }
        }
        $version = if ($tag) { $tag } else { 'dev' }
        $buildTime = (Get-Date).ToUniversalTime().ToString('yyyy-MM-dd HH:mm:ss UTC')
        $defines = @(
            "--dart-define=APP_VERSION=$version"
            "--dart-define=APP_COMMIT=$commit"
            "--dart-define=APP_TAG=$tag"
            "--dart-define=APP_BUILD_TIME=$buildTime"
            "--dart-define=MIHOMO_VERSION=$(Get-ManifestField 'mihomo.version')"
            "--dart-define=MIHOMO_REPOSITORY=$(Get-ManifestField 'mihomo.repository')"
            "--dart-define=FREE_TURN_VERSION=$(Get-ManifestField 'free-turn-proxy.version')"
            "--dart-define=FREE_TURN_REPOSITORY=$(Get-ManifestField 'free-turn-proxy.repository')"
        )
        $appName = [regex]::Match((Get-Content -LiteralPath 'ui/lib/app_info.dart' -Raw), "const appName = '([^'\r\n]+)';").Groups[1].Value
        if ($appName -notmatch '^[a-zA-Z0-9_-]+$') { throw 'Unsupported application filename.' }
    } finally { Pop-Location }

    Push-Location -LiteralPath (Join-Path $sourceRoot 'ui')
    try {
        Invoke-Checked $flutter @('clean')
        Invoke-Checked $flutter @('pub', 'get')
        Invoke-Checked $flutter (@('build', 'windows', '--release', '--no-pub') + $defines)
    } finally { Pop-Location }

    $bundle = Join-Path $sourceRoot 'ui/build/windows/x64/runner/Release'
    foreach ($relative in @("$appName.exe", 'flutter_windows.dll', 'data/app.so', 'data/icudtl.dat', 'data/flutter_assets/AssetManifest.bin')) {
        if (-not (Test-Path -LiteralPath (Join-Path $bundle $relative) -PathType Leaf)) {
            throw "Incomplete Flutter bundle: $relative"
        }
    }

    [void](New-Item -ItemType Directory -Path $distRoot -Force)
    [void](New-Item -ItemType Directory -Path $staging)
    $uiOutput = $staging
    $binOutput = Join-Path $staging 'bin'
    [void](New-Item -ItemType Directory -Path $binOutput)
    Get-ChildItem -LiteralPath $bundle -Force | Copy-Item -Destination $uiOutput -Recurse -Force
    Copy-VCRuntime (Join-Path $sourceRoot 'ui/build/windows/x64/CMakeCache.txt') $uiOutput
    Copy-Item -LiteralPath (Join-Path $cacheBin 'controller.exe') -Destination $binOutput
    foreach ($engine in @('free-turn-proxy', 'mihomo')) {
        Copy-Item -LiteralPath (Join-Path $cacheRoot "engines/windows-amd64/bin/$engine.exe") -Destination $binOutput
    }
    Copy-Item -LiteralPath (Join-Path $scriptsRoot 'windows/README.md') -Destination $staging
    if ($Portable) {
        [IO.File]::WriteAllText((Join-Path $staging 'portable.json'), "{}`n", [Text.Encoding]::ASCII)
        $readme = Join-Path $staging 'README.md'
        $portableReadme = Get-Content -LiteralPath (Join-Path $scriptsRoot 'windows/README-portable.md') -Raw -Encoding UTF8
        [IO.File]::AppendAllText($readme, "`n$portableReadme", [Text.UTF8Encoding]::new($false))
    }
    $sums = @(Get-ChildItem -LiteralPath $staging -File -Recurse | Sort-Object FullName | ForEach-Object {
        $relative = $_.FullName.Substring($staging.Length + 1).Replace('\', '/')
        "{0}  {1}" -f (Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant(), $relative
    })
    [IO.File]::WriteAllLines((Join-Path $staging 'SHA256SUMS'), [string[]]$sums, [Text.Encoding]::ASCII)

    # Retain the previous bundle until successful publication; roll back a failed move.
    Assert-DistPath $destination
    Assert-DistPath $staging
    Assert-DistPath $previous
    # Check again in case a portable folder appeared while the build was running.
    if ($Portable -and (Test-Path -LiteralPath $destination)) {
        throw "Portable output already exists: $destination. Move it with its data before building."
    }
    if (Test-Path -LiteralPath $destination) { Move-Item -LiteralPath $destination -Destination $previous }
    try { Move-Item -LiteralPath $staging -Destination $destination } catch {
        if (Test-Path -LiteralPath $previous) { Move-Item -LiteralPath $previous -Destination $destination }
        throw
    }
    Remove-TemporaryOutput $previous
    Write-Host "Bundle: $destination"
    Write-Host "Run: $(Join-Path $destination "$appName.exe")"
} finally {
    try { Remove-TemporaryOutput $staging } finally {
        foreach ($name in $savedEnvironment.Keys) {
            if ($null -eq $savedEnvironment[$name]) {
                Remove-Item -LiteralPath "Env:$name" -ErrorAction SilentlyContinue
            } else {
                [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
            }
        }
    }
}
