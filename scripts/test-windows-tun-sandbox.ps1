<# Runs real packet tests only in an already running Windows Sandbox.
   Host adapter, routing, DNS, VPN, and firewall settings are never changed. #>
[CmdletBinding()]
param([string]$SandboxId)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Windows is required.' }
$sourceRoot = Split-Path -Parent $PSScriptRoot
$runId = [guid]::NewGuid().ToString('N')
$runRoot = Join-Path $sourceRoot ".cache/windows-tun-sandbox/$runId"
$inputRoot = Join-Path $runRoot 'input'
$outputRoot = Join-Path $runRoot 'output'
$guestInput = "C:\WireglowTunInput-$runId"
$guestOutput = "C:\WireglowTunOutput-$runId"
function Invoke-Checked([string]$Program, [string[]]$Arguments) {
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed: $LASTEXITCODE" }
}
if (-not $SandboxId) {
    $list = & wsb list --raw
    if ($LASTEXITCODE -ne 0) { throw 'Windows Sandbox CLI is required.' }
    $sessions = @((($list -join "`n") | ConvertFrom-Json).WindowsSandboxEnvironments)
    if ($sessions.Count -ne 1) { throw 'Start one Windows Sandbox first, or specify -SandboxId.' }
    $SandboxId = $sessions[0].Id
}
if ($SandboxId -notmatch '^[a-fA-F0-9-]{36}$') { throw 'Invalid Sandbox ID.' }
[void](New-Item -ItemType Directory -Path "$inputRoot/bin", $outputRoot -Force)
$saved = @{}
foreach ($key in @('GOPATH','GOCACHE','GOMODCACHE','GOOS','GOARCH')) { $saved[$key] = [Environment]::GetEnvironmentVariable($key,'Process') }
try {
    $env:GOPATH = Join-Path $sourceRoot '.cache/go-windows'
    $env:GOCACHE = Join-Path $sourceRoot '.cache/go-build-windows'
    $env:GOMODCACHE = Join-Path $env:GOPATH 'pkg/mod'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    & (Join-Path $PSScriptRoot 'download-windows-engines.ps1')
    Push-Location -LiteralPath $sourceRoot
    try {
        Invoke-Checked 'go' @('build','-o',"$inputRoot/bin/tun-helper.exe",'./cmd/tun-helper')
        Invoke-Checked 'go' @('test','-c','-o',"$inputRoot/bin/controller.exe",'./internal/infrastructure/tun')
    } finally { Pop-Location }
    Copy-Item -LiteralPath (Join-Path $sourceRoot '.cache/engines/windows-amd64/bin/mihomo.exe') -Destination "$inputRoot/bin/mihomo.exe"
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'windows/test-tun-guest.ps1') -Destination "$inputRoot/test-tun-guest.ps1"
    Invoke-Checked 'wsb' @('share','--id',$SandboxId,'--host-path',$inputRoot,'--sandbox-path',$guestInput)
    Invoke-Checked 'wsb' @('share','--id',$SandboxId,'--host-path',$outputRoot,'--sandbox-path',$guestOutput,'--allow-write')
    $command = "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -ExecutionPolicy Bypass -File $guestInput\test-tun-guest.ps1 -InputRoot $guestInput -OutputRoot $guestOutput"
    Invoke-Checked 'wsb' @('exec','--id',$SandboxId,'--run-as','System','--command',$command)
    $exitPath = Join-Path $outputRoot 'exit-code.txt'
    if (-not (Test-Path -LiteralPath $exitPath)) { throw "Guest did not report completion. Inspect $outputRoot" }
    Get-Content -LiteralPath (Join-Path $outputRoot 'tests.log')
    if ((Get-Content -LiteralPath $exitPath -Raw).Trim() -ne '0') { throw "Guest tests failed. Inspect $outputRoot" }
    Write-Host "Windows Sandbox checks passed. Logs: $outputRoot"
} finally {
    foreach ($key in $saved.Keys) {
        if ($null -eq $saved[$key]) { Remove-Item -LiteralPath "Env:$key" -ErrorAction SilentlyContinue }
        else { [Environment]::SetEnvironmentVariable($key,$saved[$key],'Process') }
    }
}
