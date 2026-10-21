package tun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"wireglow/internal/infrastructure/platform"
)

const portableHelperRoot = "/run/wireglow-tun"

func (c *Client) ConfigurePortable(executable string) error {
	root := filepath.Dir(executable)
	if filepath.Base(root) == "bin" {
		root = filepath.Dir(root)
	}
	info, err := os.Stat(filepath.Join(root, "portable.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("tun: portable marker must be a regular file")
	}
	c.PortableHelper = filepath.Join(root, "bin", "tun-helper")
	c.Socket = filepath.Join(portableHelperRoot, "portable-"+strconv.Itoa(os.Getpid())+".sock")
	return nil
}
func portableUnavailable(ctx context.Context, helper string) string {
	for _, path := range []string{helper, "/usr/bin/pkexec"} {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			return "TUN requires bundled tun-helper and pkexec (polkit)"
		}
	}
	if err := checkEngineDigest(filepath.Join(filepath.Dir(helper), "mihomo")); err != nil {
		return err.Error()
	}
	if _, err := uplink(ctx, ""); err != nil {
		return err.Error()
	}
	return ""
}

// copyProtected snapshots a trusted bundle component before running it elevated.
// No caller-supplied command/config path crosses the IPC protocol.
func copyProtected(source, target string) error {
	var file *os.File
	if source == "/proc/self/exe" {
		opened, err := os.Open(source)
		if err != nil {
			return err
		}
		file = opened
	} else {
		fd, err := unix.Open(source, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if err != nil {
			return err
		}
		file = os.NewFile(uintptr(fd), source)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > 128<<20 {
		return errors.New("tun: invalid portable component")
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0500)
	if err != nil {
		return err
	}
	n, err := io.Copy(out, io.LimitReader(file, (128<<20)+1))
	if n > 128<<20 {
		err = errors.New("tun: portable component grew while being copied")
	}
	return errors.Join(err, out.Sync(), out.Close())
}

// RunPortableHelper is entered through pkexec, never installed or enabled as a
// service. The authorization approves the local distribution's helper code.
// mihomo and the running helper are snapshotted into a root-only volatile folder.
func RunPortableHelper(ctx context.Context, ownerPID int) error {
	if os.Geteuid() != 0 {
		return errors.New("tun: portable helper requires elevation")
	}
	uid64, err := strconv.ParseUint(os.Getenv("PKEXEC_UID"), 10, 32)
	if err != nil || ownerPID <= 0 {
		return errors.New("tun: pkexec caller identity required")
	}
	uid := uint32(uid64)
	info, err := os.Stat("/proc/" + strconv.Itoa(ownerPID))
	if err != nil {
		return errors.New("tun: portable owner not found")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uid {
		return errors.New("tun: portable owner UID mismatch")
	}
	owner, err := platform.ProcessIdentity(ownerPID)
	if err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if err = os.Mkdir(portableHelperRoot, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	if err = protected(portableHelperRoot); err != nil {
		return err
	}
	if err = os.Chmod(portableHelperRoot, 0755); err != nil {
		return err
	}
	dir := filepath.Dir(self)
	staged := filepath.Dir(dir) == portableHelperRoot && strings.HasPrefix(filepath.Base(dir), "bundle-")
	if !staged {
		dir, err = os.MkdirTemp(portableHelperRoot, "bundle-")
		if err != nil {
			return err
		}
		// /run is commonly noexec. Mount executable RAM only inside this
		// helper's private mount namespace; never remount the host's /run.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer func() { _ = unix.Unmount(dir, unix.MNT_DETACH); _ = os.RemoveAll(dir) }()
		if err = unix.Unshare(unix.CLONE_NEWNS); err != nil {
			return fmt.Errorf("tun: private mount namespace: %w", err)
		}
		if err = unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
			return fmt.Errorf("tun: private mount propagation: %w", err)
		}
		if err = unix.Mount("tmpfs", dir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=0700,size=256m"); err != nil {
			return fmt.Errorf("tun: temporary executable memory: %w", err)
		}
		if err = copyProtected("/proc/self/exe", filepath.Join(dir, "tun-helper")); err != nil {
			return err
		}
		if err = copyProtected(filepath.Join(filepath.Dir(self), "mihomo"), filepath.Join(dir, "mihomo")); err != nil {
			return err
		}
		if err = checkEngineDigest(filepath.Join(dir, "mihomo")); err != nil {
			return err
		}
		image := filepath.Join(dir, "tun-helper")
		return unix.Exec(image, []string{image, "--portable", "--owner", strconv.Itoa(ownerPID)}, []string{"PKEXEC_UID=" + strconv.FormatUint(uid64, 10), "PATH=/usr/sbin:/usr/bin", "HOME=" + dir})
	}
	if err = protected(self); err != nil {
		return err
	}
	defer func() { _ = unix.Unmount(dir, unix.MNT_DETACH); _ = os.RemoveAll(dir) }()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			alive, e := owner.Alive()
			if e != nil || !alive {
				cancel()
				return
			}
		}
	}()
	server := &Server{UID: uid, OwnerPID: ownerPID, OneShot: true, Binary: filepath.Join(dir, "mihomo"), Runtime: filepath.Join(portableHelperRoot, "private"), Socket: filepath.Join(portableHelperRoot, "portable-"+strconv.Itoa(ownerPID)+".sock")}
	if err = server.Serve(ctx); err != nil {
		return fmt.Errorf("tun: temporary helper: %w", err)
	}
	if server.cleanupFailed {
		return errors.New("tun: recovery incomplete; privileged journal retained")
	}
	return nil
}
