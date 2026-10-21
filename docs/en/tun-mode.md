# Linux TUN mode

Regular Linux builds use an optional root-owned helper service. Portable Linux
builds start a temporary elevated helper through polkit, without installing a
service or enabling autostart. Windows contains the shared
mode, protocol and configuration compiler plus an unavailable platform adapter;
TUN cannot be selected in the Windows menu. No Windows TUN tests are required.

## Setup

For a portable bundle, use `./scripts/build-linux-portable.sh`. No helper
installation is required. `bin/tun-helper`, pinned `bin/mihomo`, `pkexec` and a
graphical polkit authentication agent must be available. Selecting TUN requests
elevation for a temporary helper. Cancellation/denial aborts activation. An
existing VPN blocks activation before the elevation request. The helper exits
after TUN stops, its owner connection is lost, or the controller dies. Each new
helper launch can require authentication; no service or autostart is installed.

Only authorize a portable distribution you trust: the initial polkit request
approves that distribution's helper executable. It validates the caller UID and
controller PID, snapshots itself and mihomo into a root-only volatile directory
under `/run/wireglow-tun/`, verifies the pinned mihomo hash and re-executes the
protected copy. The temporary endpoint accepts only the initiating controller's
UID/PID. Its idle startup lifetime is bounded. If a helper crashes, a subsequent
elevated launch recovers the protected volatile journal before starting TUN;
a reboot clears this kernel/runtime state. Portable profiles and UI/backend
settings remain inside the bundle.

For a regular Linux bundle, build with `./scripts/build-linux.sh`. From that bundle,
install the helper explicitly for the intended non-root desktop user's numeric UID:

```sh
sudo ./install-tun-helper.sh "$(id -u)"
```

Only install a distribution you trust. `SHA256SUMS` detects corruption and does
not authenticate an untrusted distribution. Installation copies helper and mihomo
into `/usr/local/libexec/wireglow-tun/`, owned by root, and enables
`wireglow-tun.service`. It requires systemd, iproute2 (`/usr/bin/ip`) and
`/dev/net/tun`. The GUI never installs a service or elevates itself. This optional service is
only the regular Linux deployment; portable uses the temporary helper above.
Reinstall explicitly to update the privileged copy or change the permitted UID.

In the GUI choose TUN. Entering TUN reconnects the session to give FreeTurn plain
bootstrap DNS at `1.1.1.1`; this runtime override does not edit the stored profile.
The interface discloses this exception. Leaving TUN and switching between local
and system proxy modes preserves the current user engine processes.

An active TUN/WireGuard device blocks activation, including AmneziaVPN.
Disconnected devices (`DOWN` or `NO-CARRIER`) do not block activation; the error
includes the name of the detected active interface. Do not
turn off another VPN during an automated check. If the uplink changes or another
tunnel appears, the helper stops TUN on its next heartbeat; reconnect explicitly.

## Runtime and ownership

A second privileged mihomo captures IP packets and forwards them over SOCKS5 to
the existing user mihomo. Only the user engine holds WG/AWG profile secrets.
The helper compiles its entire configuration from a bounded typed request; IPC
cannot supply executables, YAML, providers, scripts or file paths. Root-owned
installation paths and the pinned decompressed mihomo SHA-256 are checked before serving. A mode-0600 Unix socket in
`/run/wireglow-tun/` checks SO_PEERCRED: the service permits one configured UID,
while a temporary portable helper additionally binds to its controller PID.
One connection owns a random session ID, request IDs and a 15-second lease;
controller heartbeats arrive every two seconds. No privileged mihomo API is exposed.

The helper owns table `20220` and policy priorities `10217`/`10218`/`10219`/`10220`. It refuses
occupied resources before starting. Mihomo `auto-route` and `auto-redirect` are
disabled: the helper journals and installs the precise IPv4/IPv6 policy itself.
Its marked outbound sockets bypass capture. FreeTurn has the first exact
PROCESS-PATH/DIRECT rule; loopback remains in the kernel local table. Only the
bootstrap resolver's IPv4 host address bypasses TUN via the original main table.
Captured DNS uses fake IPs and DoH through the existing proxy. AAAA answers are
disabled, while literal IPv6 traffic is captured and forwarded through SOCKS.
IPv6 connectivity still depends on the imported WG/AWG profile and remote peer.

The root-only journal and config live in `/run/wireglow-tun/private/`.
Stop, owner socket loss, lease expiration and engine failure stop the privileged
engine and delete only matching owned rules and its random TUN device. This
includes mihomo's IPv6 output-interface rule, which can exist with auto-route off.
Unexpected rule ownership keeps the journal and reports an error. A restarted
helper recovers the journal before accepting connections. For the installed service, RuntimeDirectoryPreserve
retains the journal across service restarts; reboot removes volatile state and
kernel routes together. This is not a kill switch: ordinary networking resumes
when TUN is removed. The helper does not edit host default routes, resolver files,
desktop proxy settings or firewall rules.

For removal of an installed service, stop and disable `wireglow-tun.service` first. Resolve any retained
recovery journal before removing the unit and privileged installation directory.
Never discard a journal merely to bypass a recovery error.

## Verification

Ordinary tests never activate TUN. Run the real packet test in a **fresh network
namespace**, with no physical interfaces or external connectivity:

```sh
source scripts/build-env.sh
go test -c -o "$cache_root/bin/tun.test" ./internal/infrastructure/tun
unshare --user --map-root-user --net env WIREGLOW_TUN_TEST=1 WIREGLOW_HOST_NETNS="$(readlink /proc/self/ns/net)" \
  WIREGLOW_MIHOMO_TEST_BINARY="$cache_root/engines/linux-amd64/bin/mihomo" \
  "$cache_root/bin/tun.test" -test.v -test.run '^TestIsolated(Tun|EstablishedRelayTCP|EstablishedRelayUDP|EstablishedRelayUDPUnconnected)$' -test.timeout 60s
```

Download pinned engines first with `./scripts/download-linux-engines.sh` and
create `.cache/bin/` if needed. Where unprivileged user namespaces are disabled,
use `sudo unshare --net` instead. The test refuses the host network namespace.
CI runs this only in the Linux integration job; Windows test jobs are unchanged.

The test uses real pinned mihomo and a synthetic SOCKS peer. It checks IPv4/IPv6
TCP, UDP, DNS interception, live relay socket process matching, normal cleanup,
engine crash, owner socket loss and refusal of another tunnel/occupied policy
resources. It compares routing rules before and after recovery. Controller tests
also check helper loss and refusal to switch an unprepared relay into TUN.

Limits: closed, very short-lived sockets can disappear before mihomo attributes
them to a process. The bypass test holds relay transport sockets open; it does
not prove attribution of closed sockets. The real graphical polkit elevation flow, live FreeTurn/TURN credential renewal,
CAPTCHA renewal, remote connectivity, installed-service operation across user
IDs, suspend/resume and Windows TUN have not been validated by this isolated
suite. Do not treat the synthetic SOCKS test as an end-to-end remote VPN test.

## Verification record — 2026-10-21

Linux x64, Go 1.27.1, Flutter 3.47.7 / Dart 3.13.5, mihomo v1.19.32:
Go vet for Linux and Windows, Linux Go tests, lifecycle/recovery integrations,
controller/TUN race checks, Flutter analysis/tests, Dart↔Go IPC smoke and a full
Linux bundle build passed. The packet suite ran only under
`unshare --user --map-root-user --net`, using synthetic endpoints. The helper
was not installed on the laptop and host traffic was never captured.

Diagnostics also found and fixed an atomicity issue in the existing owner-clock
test and an overlong socket path in the smoke harness. Runtime socket state for
that harness now uses a short system temporary directory. No desktop proxy,
VPN or firewall settings were changed by these diagnostics.

The portable bootstrap test additionally runs the real helper in a synthetic
chroot inside mount/PID/network namespaces. It validates protected re-exec,
normal stop, owner socket loss, process exit, binary/journal cleanup and unchanged
routing rules, with no service files created. Local command after building
`.cache/bin/tun-helper` and downloading engines:

```sh
unshare --user --map-root-user --mount --net --pid --fork python3 \
  internal/infrastructure/tun/testdata/portable_helper.py "$PWD" "$(readlink /proc/self/ns/net)"
```

This simulates an already authorized caller and does not test the graphical
polkit prompt. CI runs it only on Linux with `sudo unshare`.

Portable uses its own `portable-<controller PID>.sock` endpoint and never reuses the installed service. Helpers of this version share one network-session lock/journal, so an idle resident helper does not block portable and two helpers cannot activate TUN simultaneously. The bootstrap test also checks coexistence with an idle resident helper and cancellation before Start.

The mode menus on the main screen and in settings refresh capabilities through
the controller before each opening. A VPN started after the client disables TUN
when the menu is opened again. If the check fails, the menu stays closed; TUN
activation still performs its own safety check.

When Linux mounts `/run` with `noexec`, the portable helper creates a private
mount namespace with an executable tmpfs visible only inside that namespace.
Host `/run` flags remain unchanged. Temporary images are removed on exit; no
service installation is needed. The elevated helper must be allowed to create
a mount namespace. Activation failures expose the TUN adapter diagnostic in UI.

Probe failures include a sanitized cause: HTTPS/TLS, UDP SOCKS, a DNS reply
or a timeout. Raw URLs, profile data and engine output remain private.
Improved diagnostics alone do not establish that a connectivity failure is fixed.

During a TUN session, the inner WG engine uses its own DNS in `redir-host`
mode: `https://1.0.0.1/dns-query#PROXY`. It avoids the system resolver, which
the outer TUN can feed fake IPs belonging to a different mapping cache. Local
and system proxy modes retain their previous DNS behavior. A synthetic test
with real pinned mihomo verifies DNS over SOCKS and a real IP answer. See the
[mihomo DNS configuration](https://wiki.metacubex.one/en/config/dns/).

HTTPS probe diagnostics distinguish the local proxy CONNECT response, TLS
and reading the HTTPS response. EOF, reset/refused/unreachable, certificate
errors and timeouts use fixed labels. Unknown errors expose only their Go type,
without raw error text containing URLs.

Before capture, the helper installs exact priority `10218` rules for existing
FreeTurn TCP sockets: address family, local/remote addresses and both ports,
using lookup main. An established TCP stream cannot first enter a userspace TCP
stack midway through its connection: PROCESS-PATH is too late for that stream.
Exemptions are journaled and cleanup deletes only matching rules. New relay
connections retain PROCESS-PATH/DIRECT routing. An isolated regression opens
TCP over a veth to a peer namespace before TUN, then checks stream survival
and exact rule cleanup after an engine crash.

Linux mihomo diagnostics retain only fixed warning categories (DNS, refused
connections, timeouts, etc.). Lines are bounded to 4096 bytes and discarded
after classification; raw logs are never written to disk. Categories reset
before TUN activation. Failure snapshots include mihomo categories and, for
portable sessions, exact TCP exemption counts and remaining socket counts.
No warnings does not establish a healthy connection: mihomo does not log
some failures occurring after a lazy handshake.

For existing relay UDP sockets, priority `10217` rules select
`uidrange <UID>-<UID> ipproto udp sport <port> lookup main`. They preserve
the original source address and port; userspace DIRECT forwarding created
a different UDP endpoint. Ports come only from FreeTurn process fd/inode
ownership, and socket UID must match the caller. A port shared with another
process prevents activation. Dual-stack wildcard UDP receives IPv4 and IPv6
rules. UDP rules are journaled and deleted only on an exact match. Regression
tests compare the peer-visible endpoint before/after TUN for connected and
unconnected UDP and verify cleanup after an engine crash.

Known limitation: UDP exemptions are a snapshot taken at TUN activation.
New or recreated FreeTurn sockets do not receive kernel exemptions during the
session and depend on mihomo's process rule. Old port exemptions remain until
TUN cleanup. FreeTurn can recreate transports during stream recovery; dynamic
socket handling remains unfinished.
