"""Black-box portable helper test. Run only in fresh mount/PID/network namespaces.

Authorization is already granted by the isolated user namespace; this does not
exercise pkexec or the desktop authentication agent. No host installation occurs.
"""
import json
import os
import pathlib
import shutil
import socket
import subprocess
import sys
import tempfile
import time

project = pathlib.Path(sys.argv[1]).resolve()
if os.geteuid() != 0:
    raise RuntimeError("an isolated root identity is required")
if len(sys.argv)!=3 or os.readlink("/proc/self/ns/net") == sys.argv[2]:
    raise RuntimeError("refusing helper test in host network namespace")


def run(*args):
    subprocess.run(args, check=True, stdout=subprocess.DEVNULL)


def rules(family):
    return subprocess.check_output(["/usr/bin/ip", family, "-j", "rule", "show"])


def copy_program(source, destination, root):
    destination.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source, destination)
    destination.chmod(0o755)
    libraries = subprocess.run(
        ["ldd", str(source)], capture_output=True, text=True
    ).stdout
    for line in libraries.splitlines():
        for part in line.split():
            if part.startswith("/"):
                target = root / part.lstrip("/")
                target.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(part, target)
                target.chmod(0o755)


def exchange(connection, request, allow_error=False):
    connection.sendall(json.dumps(request).encode() + b"\n")
    data = b""
    while not data.endswith(b"\n"):
        chunk = connection.recv(4096)
        if not chunk:
            raise AssertionError("helper connection closed")
        data += chunk
    reply = json.loads(data)
    if not allow_error:
        assert not reply.get("error"), reply.get("error")
    return reply


for args in [
    ("link", "set", "lo", "up"),
    ("link", "add", "uplink0", "type", "dummy"),
    ("link", "set", "uplink0", "up"),
    ("addr", "add", "192.0.2.2/24", "dev", "uplink0"),
    ("route", "add", "default", "dev", "uplink0"),
]:
    run("/usr/bin/ip", *args)

baseline = (rules("-4"), rules("-6"))
with tempfile.TemporaryDirectory(prefix="wg-portable-root-", dir="/tmp") as temporary:
    root = pathlib.Path(temporary)
    root.chmod(0o755)
    mounts = []
    relay = None
    proxy = None
    resident = None
    try:
        # A real root mountpoint is needed for recursive private propagation
        # after chroot. All mounts belong to the outer isolated test namespace.
        run("mount", "--bind", str(root), str(root))
        mounts.append(root)
        for name in ["run", "proc", "dev/net"]:
            (root / name).mkdir(parents=True, exist_ok=True)
        # Match the laptop: /run is writable but disallows executable images.
        run("mount", "-t", "tmpfs", "-o", "noexec,nosuid,nodev,mode=0755", "tmpfs", str(root / "run"))
        mounts.append(root / "run")
        copy_program(project / ".cache/bin/tun-helper", root / "app/bin/tun-helper", root)
        copy_program(pathlib.Path("/usr/bin/ip"), root / "usr/bin/ip", root)
        copy_program(pathlib.Path("/usr/bin/sleep"), root / "app/bin/free-turn-proxy", root)
        copy_program(
            project / ".cache/engines/linux-amd64/bin/mihomo",
            root / "app/bin/mihomo", root,
        )
        for directory in root.rglob("*"):
            if directory.is_dir():
                directory.chmod(0o755)
        for device in ["/dev/null", "/dev/net/tun"]:
            target = root / device.lstrip("/")
            target.touch()
            run("mount", "--bind", device, str(target))
            mounts.append(target)
        run("mount", "-t", "proc", "proc", str(root / "proc"))
        mounts.append(root / "proc")
        relay = subprocess.Popen([
            "chroot", str(root), "/app/bin/free-turn-proxy", "60",
        ])
        proxy = socket.socket()
        proxy.bind(("127.0.0.1", 0))
        proxy.listen()
        port = proxy.getsockname()[1]
        for mode in ["stop", "owner-socket-loss", "cancel-before-start", "with-resident-helper"]:
            if mode == "with-resident-helper":
                installed = root / "usr/local/libexec/wireglow-tun/mihomo"
                copy_program(project / ".cache/engines/linux-amd64/bin/mihomo", installed, root)
                for directory in [root / "usr/local", root / "usr/local/libexec", installed.parent]:
                    directory.chmod(0o755)
                resident = subprocess.Popen(
                    ["chroot", str(root), "/app/bin/tun-helper", "--uid", "0"],
                    stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True,
                )
                deadline = time.monotonic() + 10
                while True:
                    if resident.poll() is not None:
                        raise AssertionError("resident helper exited: " + resident.stderr.read())
                    candidate = socket.socket(socket.AF_UNIX)
                    try:
                        candidate.connect(str(root / "run/wireglow-tun/control.sock"))
                        candidate.close()
                        break
                    except OSError:
                        candidate.close()
                        if time.monotonic() > deadline:
                            raise AssertionError("resident helper readiness timeout")
                        time.sleep(0.05)
            process = subprocess.Popen(
                ["chroot", str(root), "/app/bin/tun-helper", "--portable",
                 "--owner", str(os.getpid())],
                env=dict(os.environ, PKEXEC_UID="0"),
                stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True,
            )
            connection = None
            try:
                deadline = time.monotonic() + 10
                while True:
                    if process.poll() is not None:
                        raise AssertionError("bootstrap exited: " + process.stderr.read())
                    candidate = socket.socket(socket.AF_UNIX)
                    try:
                        candidate.connect(str(root / ("run/wireglow-tun/portable-" + str(os.getpid()) + ".sock")))
                        connection = candidate
                        break
                    except OSError:
                        candidate.close()
                        if time.monotonic() > deadline:
                            raise AssertionError("helper readiness timeout")
                        time.sleep(0.05)
                connection.settimeout(15)
                if mode == "cancel-before-start":
                    exchange(connection, dict(version=1, id=1, operation="cancel"))
                else:
                    session = exchange(connection, dict(
                        version=1, id=1, operation="start", proxyPort=port, relayPID=relay.pid,
                    ))["session"]
                    exchange(connection, dict(version=1, id=2, operation="heartbeat", session=session))
                    assert list((root / "run/wireglow-tun").glob("bundle-*"))
                    assert not list((root / "run/wireglow-tun").glob("bundle-*/tun-helper")), "executable mount leaked to parent"
                    if mode == "with-resident-helper":
                        other = socket.socket(socket.AF_UNIX)
                        other.settimeout(5)
                        other.connect(str(root / "run/wireglow-tun/control.sock"))
                        rejected = exchange(other, dict(version=1, id=1, operation="start", proxyPort=port, relayPID=relay.pid), allow_error=True)
                        other.close()
                        assert "another helper owns" in rejected.get("error", "")
                    if mode in ["stop", "with-resident-helper"]:
                        exchange(connection, dict(version=1, id=3, operation="stop", session=session))
                connection.close()
                process.wait(timeout=15)
                assert process.returncode == 0, process.stderr.read()
                assert not list((root / "run/wireglow-tun").glob("bundle-*"))
                assert not (root / "run/wireglow-tun/private/journal.json").exists()
                assert not (root / "etc/systemd").exists()
                if mode != "with-resident-helper":
                    assert not (root / "usr/local").exists()
                else:
                    assert resident.poll() is None, "portable stopped the existing helper"
                assert (rules("-4"), rules("-6")) == baseline
                print("PASS: protected portable bootstrap/re-exec and " + mode, flush=True)
            finally:
                if connection is not None:
                    connection.close()
                if process.poll() is None:
                    process.terminate()
                    process.wait(timeout=20)
    finally:
        if resident is not None and resident.poll() is None:
            resident.terminate()
            resident.wait(timeout=20)
        if relay is not None:
            relay.terminate()
            relay.wait()
        if proxy is not None:
            proxy.close()
        for mount in reversed(mounts):
            run("umount", str(mount))
