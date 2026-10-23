package tun

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"wireglow/internal/infrastructure/engines/manifest"
	"wireglow/internal/infrastructure/platform"
)

type windowsOwner struct {
	Identity platform.Identity
	SID      string
}

func processOwner(pid int) (windowsOwner, error) {
	id, err := platform.ProcessIdentity(pid)
	if err != nil {
		return windowsOwner{}, err
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return windowsOwner{}, err
	}
	defer windows.CloseHandle(h)
	var token windows.Token
	if err = windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token); err != nil {
		return windowsOwner{}, err
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		return windowsOwner{}, err
	}
	return windowsOwner{Identity: id, SID: user.User.Sid.String()}, nil
}

// Files are read through non-reparse handles that disallow write/delete sharing.
// The verified handle stays open while copying; a writable bundle cannot swap it.
func openComponent(path string) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.READ_CONTROL, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(h, &info); err != nil || info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT) != 0 {
		windows.CloseHandle(h)
		return nil, errors.New("tun: unsafe component")
	}
	return os.NewFile(uintptr(h), path), nil
}

func checkWindowsEngine(path string) error {
	f, err := openComponent(path)
	if err != nil {
		return errors.New("tun: pinned mihomo.exe is unavailable")
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, (128<<20)+1))
	want, _ := manifest.Field("mihomo.platforms.windows-amd64.binary-sha256")
	if err != nil || n > 128<<20 || want == "" || hex.EncodeToString(h.Sum(nil)) != want {
		return errors.New("tun: mihomo.exe does not match the pinned digest")
	}
	return nil
}

const adminSDDL = "O:BAG:BAD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)"

func adminDirectory(path string) error {
	sd, err := windows.SecurityDescriptorFromString(adminSDDL)
	if err != nil {
		return err
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	err = windows.CreateDirectory(name, &sa)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return err
	}
	h, err := windows.CreateFile(name, windows.READ_CONTROL, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(h, &info); err != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return errors.New("tun: unsafe privileged directory")
	}
	return requireAdminACL(h)
}

func requireAdminACL(h windows.Handle) error {
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil || (!owner.IsWellKnown(windows.WinBuiltinAdministratorsSid) && !owner.IsWellKnown(windows.WinLocalSystemSid)) {
		return errors.New("tun: privileged state has an unexpected owner")
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil {
		return errors.New("tun: privileged state has no restrictive ACL")
	}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if windows.GetAce(acl, i, &ace) != nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return errors.New("tun: unsupported privileged ACL")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsWellKnown(windows.WinBuiltinAdministratorsSid) && !sid.IsWellKnown(windows.WinLocalSystemSid) {
			return errors.New("tun: privileged state is writable by a user")
		}
	}
	return nil
}

func copyWindowsEngine(source, target string) error {
	f, err := openComponent(source)
	if err != nil {
		return err
	}
	defer f.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	n, copyErr := io.Copy(out, io.LimitReader(f, (128<<20)+1))
	err = errors.Join(copyErr, out.Sync(), out.Close())
	if err != nil {
		return err
	}
	if n > 128<<20 {
		return errors.New("tun: oversized engine")
	}
	return checkWindowsEngine(target)
}

func stageRoot() (string, error) {
	root, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return "", err
	}
	root = filepath.Join(root, "Wireglow-Tun")
	return root, adminDirectory(root)
}

// Set the descriptor on the creation handle, never by a pathname that the
// portable directory's owner could replace between writing and signing it.
func writeAdminFile(path string, data []byte) error {
	sd, err := windows.SecurityDescriptorFromString(adminSDDL)
	if err != nil {
		return err
	}
	id, err := randomWindowsID()
	if err != nil {
		return err
	}
	temp := filepath.Join(filepath.Dir(path), ".journal-"+id)
	p, err := windows.UTF16PtrFromString(temp)
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	h, err := windows.CreateFile(p, windows.GENERIC_WRITE|windows.READ_CONTROL, 0, &sa, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(h), temp)
	_, err = f.Write(data)
	err = errors.Join(err, f.Sync(), f.Close())
	defer os.Remove(temp)
	if err != nil {
		return err
	}
	dst, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(p, dst, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

// Keep every ancestor open without delete sharing. A portable directory may be
// user-writable, but it must not be renamed or replaced by a junction while an
// elevated helper uses absolute journal paths below it.
func pinWindowsDirectories(path string) (func(), error) {
	path = filepath.Clean(path)
	volume := filepath.VolumeName(path)
	if len(volume) != 2 || volume[1] != ':' {
		return nil, errors.New("tun: portable TUN requires a local drive path")
	}
	var handles []windows.Handle
	closeAll := func() {
		for i := len(handles) - 1; i >= 0; i-- {
			windows.CloseHandle(handles[i])
		}
	}
	current := volume + string(filepath.Separator)
	parts := strings.Split(strings.TrimPrefix(path, current), string(filepath.Separator))
	for _, part := range append([]string{""}, parts...) {
		if part != "" {
			current = filepath.Join(current, part)
		}
		name, err := windows.UTF16PtrFromString(current)
		if err != nil {
			closeAll()
			return nil, err
		}
		h, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
		if err != nil {
			closeAll()
			return nil, err
		}
		handles = append(handles, h)
		var info windows.ByHandleFileInformation
		if err = windows.GetFileInformationByHandle(h, &info); err != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
			closeAll()
			return nil, errors.New("tun: reparse points are forbidden in privileged portable paths")
		}
	}
	return closeAll, nil
}
