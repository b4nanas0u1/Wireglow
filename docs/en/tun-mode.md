# TUN mode

TUN is experimental. Regular Linux builds use an installed helper; portable
Linux and Windows builds start a temporary elevated helper. Run the GUI as an
ordinary user. Another active VPN or tunnel blocks activation; reconnect if the
outgoing interface changes.

## Windows portable

Build with `./scripts/build-windows-portable.ps1` and launch `Wireglow.exe`.
Selecting TUN requests UAC approval for the bundled `tun-helper.exe`.
No service or autostart is installed; each new helper launch requires elevation.
TUN is unavailable in regular Windows builds: the installed Windows service
has not been implemented.

TURN requires TCP, selected by default. TURN UDP is rejected before engines
start or UAC is requested. TCP/UDP application traffic through TUN and
direct-provider UDP forwarding are supported.

FreeTurn uses HTTPS bootstrap DNS without changing the stored profile.
Captured application DNS queries travel through the connection. AAAA replies
are disabled, but literal IPv6 traffic is captured; remote IPv6 support depends
on the profile.

Existing relay TCP destinations and bootstrap DNS addresses are exempt from
capture. Other applications reaching those addresses also bypass TUN.
Existing more-specific Windows routes, including LAN prefixes, can take precedence.
Configurable routing/LAN policy is not implemented.

Disable TUN and close the application before moving or updating the portable folder.
The recovery journal is at `data/runtime/tun/journal.json`; do not delete it to
retry after a recovery error. The next elevated launch checks remaining state
first. Temporary privileged files can be created outside the portable folder.
The Wintun driver can remain installed after disconnection and is not
uninstalled automatically.

This is not a kill switch: ordinary networking resumes when TUN is removed.
The mode is not a verified policy against every possible bypass.

## Linux

For portable builds, use `./scripts/build-linux-portable.sh`. The bundled
`bin/tun-helper` and `bin/mihomo`, `pkexec`, and a graphical polkit
authentication agent are required. Selecting TUN requests elevation for a temporary
helper; cancellation or denial aborts activation. No service or autostart is
installed. Each new helper launch can require authentication.
Only authorize a distribution you trust.

For regular builds, run `./scripts/build-linux.sh`, then install the helper
from the trusted distribution folder for the ordinary desktop user:

```sh
sudo ./install-tun-helper.sh "$(id -u)"
```

systemd, iproute2 (`/usr/bin/ip`), and `/dev/net/tun` are required.
Installation enables `wireglow-tun.service`; the GUI does not install the
service or elevate itself. Reinstall to update privileged components or change
the permitted UID. `SHA256SUMS` detects corruption but does not authenticate
an untrusted distribution.

TURN UDP requires Linux 5.6 or newer and security policy permitting
`pidfd_getfd` access. Access denial fails activation explicitly.
Update the controller and helper together.

If `/run` is mounted with `noexec`, the portable helper requires permission to
create a mount namespace; denial fails TUN activation.

Entering TUN reconnects the session and gives FreeTurn direct bootstrap DNS
at `1.1.1.1`; the stored profile is unchanged. Captured application DNS uses
DoH through the connection. AAAA replies are disabled; literal IPv6 traffic is
captured, with connectivity depending on the profile and remote peer.
Leaving TUN and switching between local/system proxy modes preserves the running
user engine processes.

Disconnected interfaces (`DOWN` or `NO-CARRIER`) do not block activation.
An error identifies the detected active interface.
In direct sessions, other applications with the same UID also bypass TUN when
sending UDP to the same remote address and port. Exemptions for existing UDP
sockets are fixed at activation and do not migrate automatically during a session.
Short-lived sockets can close before process lookup; attribution of already
closed sockets is not guaranteed.

Stopping or failure removes owned network resources. This is not a kill switch:
ordinary networking resumes when TUN is removed. The recovery journal is under
`/run/wireglow-tun/`; do not delete it to bypass a recovery error.
To remove an installed service, stop and disable `wireglow-tun.service` first,
resolve any remaining journal, then remove the unit and installation directory
`/usr/local/libexec/wireglow-tun/`.

## Isolated checks

Ordinary tests do not activate TUN. Do not activate TUN on the host or disable
its VPN during automated verification.

### Windows

Start Windows Sandbox first, then run from the repository root:

```powershell
./scripts/test-windows-tun-sandbox.ps1
```

The installed `wsb` CLI is required. Checks run inside Sandbox and refuse
the host OS. Logs are in `.cache/windows-tun-sandbox/<run ID>/output/`.
The suite checks IPv4/IPv6 TCP/UDP, DNS, disconnection, failures, and recovery
with real mihomo and a synthetic SOCKS server.

It does not prove real UAC approval/cancellation, authentication/CAPTCHA,
TURN renewal, remote WG/AWG connectivity, suspend/resume, or preservation of
established relay transport. The bypass check uses open sockets and does not
guarantee attribution after they close.

### Linux

Download engines with `./scripts/download-linux-engines.sh`. From the repository root:

```sh
source scripts/build-env.sh
mkdir -p "$cache_root/bin"
go test -c -o "$cache_root/bin/tun.test" ./internal/infrastructure/tun
unshare --user --map-root-user --net env WIREGLOW_TUN_TEST=1 WIREGLOW_HOST_NETNS="$(readlink /proc/self/ns/net)" \
  WIREGLOW_MIHOMO_TEST_BINARY="$cache_root/engines/linux-amd64/bin/mihomo" \
  "$cache_root/bin/tun.test" -test.v -test.run '^TestIsolated(Tun|EstablishedRelayTCP|EstablishedRelayUDP|EstablishedRelayUDPUnconnected|DirectUDP|TURNUDPIdle)$' -test.timeout 120s
```

Use a fresh network namespace with no physical interfaces or external connectivity.
Where user namespaces are disabled, use `sudo unshare --net`.
The test refuses the host network namespace. It checks IPv4/IPv6 TCP/UDP, DNS,
bypass for open relay sockets, cleanup, and recovery with real mihomo and a
synthetic SOCKS server.

This check does not validate graphical polkit approval, TURN credential/CAPTCHA
renewal, remote connectivity, installed-service operation across user IDs, or
suspend/resume. Synthetic checks on either platform do not replace an end-to-end
remote connection test.
