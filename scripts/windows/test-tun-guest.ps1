# Called only by test-windows-tun-sandbox.ps1 inside the guest. Shares contain
# synthetic test binaries and logs, never a user's profiles or credentials.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$InputRoot, [Parameter(Mandatory)][string]$OutputRoot)
$ErrorActionPreference = 'Stop'
$testAddress = $null
$testRoute = $null
$restoreDhcp = $false
$code = 99
try {
    $containerService = Get-CimInstance Win32_Service -Filter "Name='cexecsvc'"
    if (-not $containerService -or $containerService.PathName -notmatch '(?i)cexecsvc\.exe') { throw 'Refusing to run outside Windows Sandbox.' }
    if ($InputRoot -notmatch '^C:\\WireglowTunInput-[a-f0-9]{32}$' -or $OutputRoot -notmatch '^C:\\WireglowTunOutput-[a-f0-9]{32}$') { throw 'Unexpected guest share paths.' }
    $root = Join-Path 'C:\' ('WireglowTunTest-' + [guid]::NewGuid().ToString('N'))
    [void](New-Item -ItemType Directory -Path "$root/bin" -Force)
    foreach ($name in @('controller.exe','tun-helper.exe','mihomo.exe')) { Copy-Item -LiteralPath (Join-Path $InputRoot "bin/$name") -Destination "$root/bin/$name" }
    # No real connectivity is needed. If DHCP did not provide a gateway, use
    # documentation-only addresses on the guest Ethernet for the physical path.
    $defaults = @(Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction SilentlyContinue)
    if ($defaults.Count -eq 0) {
        $adapter = @(Get-NetAdapter | Where-Object Status -eq 'Up')
        if ($adapter.Count -ne 1) { throw 'A single guest Ethernet adapter is required.' }
        $adapter = $adapter[0]
        $ipInterface = Get-NetIPInterface -InterfaceIndex $adapter.ifIndex -AddressFamily IPv4
        $restoreDhcp = $ipInterface.Dhcp -eq 'Enabled'
        $testAddress = New-NetIPAddress -InterfaceIndex $adapter.ifIndex -IPAddress '192.0.2.2' -PrefixLength 24 -PolicyStore ActiveStore
        $testRoute = New-NetRoute -InterfaceIndex $adapter.ifIndex -DestinationPrefix '0.0.0.0/0' -NextHop '192.0.2.1' -RouteMetric 5000 -PolicyStore ActiveStore
        # Let DAD complete and the old APIPA route disappear before recording baseline.
        Start-Sleep -Seconds 5
    }
    $env:WIREGLOW_WINDOWS_TUN_TEST = '1'
    $env:WIREGLOW_MIHOMO_TEST_BINARY = "$root/bin/mihomo.exe"
    $testProcess = Start-Process -FilePath "$root/bin/controller.exe" -ArgumentList @('-test.v','-test.run=^TestWindows','-test.timeout=240s') -WindowStyle Hidden -RedirectStandardOutput "$OutputRoot/tests.log" -RedirectStandardError "$OutputRoot/tests.stderr" -Wait -PassThru
    $code = $testProcess.ExitCode
} catch {
    $_ | Out-String | Set-Content "$OutputRoot/runner-error.txt"
} finally {
    try { if ($testRoute) { $testRoute | Remove-NetRoute -Confirm:$false -ErrorAction Stop } }
    catch { $code = 98; $_ | Out-String | Add-Content "$OutputRoot/runner-error.txt" }
    try { if ($testAddress) { $testAddress | Remove-NetIPAddress -Confirm:$false -ErrorAction Stop } }
    catch { $code = 98; $_ | Out-String | Add-Content "$OutputRoot/runner-error.txt" }
    try { if ($restoreDhcp) { Set-NetIPInterface -InterfaceIndex $adapter.ifIndex -AddressFamily IPv4 -Dhcp Enabled -ErrorAction Stop } }
    catch { $code = 98; $_ | Out-String | Add-Content "$OutputRoot/runner-error.txt" }
    [IO.File]::WriteAllText("$OutputRoot/exit-code.txt", [string]$code)
}
