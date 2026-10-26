package tun

import (
	"errors"
	"golang.org/x/sys/windows"
	"path/filepath"
	"runtime"
	"unsafe"
)

// ShellExecuteEx supplies an actual process handle; ShellExecute's pseudo-handle
// cannot authenticate a pipe server or report UAC cancellation reliably.
type shellExecuteInfo struct {
	Size                              uint32
	Mask                              uint32
	Window                            windows.Handle
	Verb, File, Parameters, Directory *uint16
	Show                              int32
	Instance                          windows.Handle
	IDList                            uintptr
	Class                             *uint16
	ClassKey                          windows.Handle
	HotKey                            uint32
	Icon                              windows.Handle
	Process                           windows.Handle
}

func elevateHelper(path string, args []string) (*helperProcess, error) {
	verb, _ := windows.UTF16PtrFromString("runas")
	file, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	params, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(args))
	if err != nil {
		return nil, err
	}
	dir, _ := windows.UTF16PtrFromString(filepath.Dir(path))
	info := shellExecuteInfo{Mask: 0x40 | 0x100 | 0x400, Verb: verb, File: file, Parameters: params, Directory: dir, Show: windows.SW_HIDE}
	info.Size = uint32(unsafe.Sizeof(info))
	// Shell extensions require a COM apartment. Keep it on this OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	err = windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED)
	if err == nil {
		defer windows.CoUninitialize()
	}
	proc := windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteExW")
	ok, _, callErr := proc.Call(uintptr(unsafe.Pointer(&info)))
	runtime.KeepAlive(file)
	runtime.KeepAlive(params)
	runtime.KeepAlive(dir)
	runtime.KeepAlive(verb)
	if ok == 0 {
		if errors.Is(callErr, windows.ERROR_CANCELLED) {
			return nil, errors.New("tun: UAC elevation cancelled")
		}
		return nil, errors.New("tun: could not elevate the bundled helper")
	}
	if info.Process == 0 {
		return nil, errors.New("tun: elevation returned no process handle")
	}
	pid, err := windows.GetProcessId(info.Process)
	if err != nil {
		windows.CloseHandle(info.Process)
		return nil, err
	}
	p := &helperProcess{pid: pid, handle: info.Process, done: make(chan struct{})}
	go func() {
		_, err := windows.WaitForSingleObject(p.handle, windows.INFINITE)
		var code uint32
		if err == nil {
			err = windows.GetExitCodeProcess(p.handle, &code)
		}
		if err == nil && code != 0 {
			err = errors.New("tun: helper failed; recovery journal retained if cleanup is incomplete")
		}
		p.err = err
		close(p.done)
	}()
	return p, nil
}
