#requires -Version 5.1
[CmdletBinding()]
param()

& (Join-Path $PSScriptRoot 'build-windows.ps1') -Portable
