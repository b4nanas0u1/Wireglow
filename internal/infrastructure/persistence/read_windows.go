package persistence

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// Allow replacement of a journal/lease while the reader keeps its snapshot.
// Use Win32 opening semantics: Root.Open's NT reparse handling rejects some
// otherwise readable AppData files. Do not fall back to os.Open, which would
// reintroduce the sharing violation during publication.
func openReadFile(path string) (*os.File, error) {
	name := path
	if len(name) >= 248 && !strings.HasPrefix(name, `\\?\`) && !strings.HasPrefix(name, `\\.\`) {
		absolute, err := filepath.Abs(name)
		if err != nil {
			return nil, &os.PathError{Op: "open", Path: path, Err: err}
		}
		if strings.HasPrefix(absolute, `\\`) {
			name = `\\?\UNC\` + absolute[2:]
		} else {
			name = `\\?\` + absolute
		}
	}
	encoded, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	handle, err := windows.CreateFile(encoded, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}
