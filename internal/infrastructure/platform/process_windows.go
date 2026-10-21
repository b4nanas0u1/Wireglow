package platform

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"sync"

	"unsafe"

	"wireglow/internal/application/ports"

	"golang.org/x/sys/windows"
)

type Runner struct{}
type process struct {
	pid    uint32
	handle windows.Handle
	job    windows.Handle
	done   chan struct{}
	err    error
	mu     sync.Mutex
}

// startWindows не позволяет коду процесса исполниться до присоединения к Job.
func startWindows(ctx context.Context, spec ports.Command, managed bool) (ports.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(spec.Path) {
		return nil, errors.New("process: absolute executable required")
	}
	var job windows.Handle
	var err error
	if managed {
		job, err = windows.CreateJobObject(nil, nil)
		if err != nil {
			return nil, err
		}
		info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
		info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
		if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
			windows.CloseHandle(job)
			return nil, err
		}
	}
	app, err := windows.UTF16PtrFromString(spec.Path)
	if err != nil {
		if job != 0 {
			windows.CloseHandle(job)
		}
		return nil, err
	}
	line, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(append([]string{spec.Path}, spec.Args...)))
	if err != nil {
		if job != 0 {
			windows.CloseHandle(job)
		}
		return nil, err
	}
	var dir *uint16
	if spec.Dir != "" {
		dir, err = windows.UTF16PtrFromString(spec.Dir)
		if err != nil {
			if job != 0 {
				windows.CloseHandle(job)
			}
			return nil, err
		}
	}
	si := windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfo{}))}
	var pi windows.ProcessInformation
	flags := uint32(windows.CREATE_SUSPENDED | windows.CREATE_NO_WINDOW)
	var environment *uint16
	if spec.ClearProxyEnv {
		var data []uint16
		for _, entry := range withoutProxyEnvironment() {
			encoded, e := windows.UTF16FromString(entry)
			if e != nil {
				if job != 0 {
					windows.CloseHandle(job)
				}
				return nil, e
			}
			data = append(data, encoded...)
		}
		data = append(data, 0)
		environment = &data[0]
		flags |= windows.CREATE_UNICODE_ENVIRONMENT
	}
	if err = windows.CreateProcess(app, line, nil, nil, false, flags, environment, dir, &si, &pi); err != nil {
		if job != 0 {
			windows.CloseHandle(job)
		}
		return nil, errors.New("process: CreateProcess failed")
	}
	defer windows.CloseHandle(pi.Thread)
	abort := func() {
		windows.TerminateProcess(pi.Process, 1)
		windows.WaitForSingleObject(pi.Process, 1000)
		windows.CloseHandle(pi.Process)
		if job != 0 {
			windows.CloseHandle(job)
		}
	}
	if managed {
		if err = windows.AssignProcessToJobObject(job, pi.Process); err != nil {
			abort()
			return nil, errors.New("process: Job assignment failed")
		}
	}
	if _, err = windows.ResumeThread(pi.Thread); err != nil {
		abort()
		return nil, err
	}
	p := &process{pid: pi.ProcessId, handle: pi.Process, job: job, done: make(chan struct{})}
	go func() {
		_, err := windows.WaitForSingleObject(pi.Process, windows.INFINITE)
		var code uint32
		if err == nil {
			err = windows.GetExitCodeProcess(pi.Process, &code)
		}
		if err == nil && code != 0 {
			err = errors.New("process: nonzero exit status")
		}
		p.mu.Lock()
		p.err = err
		windows.CloseHandle(p.handle)
		p.handle = 0
		if p.job != 0 {
			windows.CloseHandle(p.job)
			p.job = 0
		}
		p.mu.Unlock()
		close(p.done)
	}()
	runtime.KeepAlive(line)
	return p, nil
}

func (Runner) Start(ctx context.Context, spec ports.Command) (ports.Process, error) {
	return startWindows(ctx, spec, true)
}
func StartRecovery(ctx context.Context, spec ports.Command, _ string) (ports.Process, error) {
	return startWindows(ctx, spec, false)
}
func StartController(ctx context.Context, spec ports.Command, _ string) (ports.Process, error) {
	return startWindows(ctx, spec, true)
}
func (p *process) Done() <-chan struct{} { return p.done }
func (p *process) Err() error {
	select {
	case <-p.done:
		return p.err
	default:
		return nil
	}
}
func (p *process) Stop(ctx context.Context) error {
	p.mu.Lock()
	var err error
	if p.job != 0 {
		err = windows.TerminateJobObject(p.job, 1)
	} else if p.handle != 0 {
		err = windows.TerminateProcess(p.handle, 1)
	}
	p.mu.Unlock()
	if err != nil {
		return err
	}
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *process) PID() int { return int(p.pid) }
