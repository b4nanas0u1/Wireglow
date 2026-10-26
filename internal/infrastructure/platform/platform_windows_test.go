package platform

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"wireglow/internal/application/ports"

	"golang.org/x/sys/windows"
)

func TestWinINetABI(t *testing.T) {
	// Это контракт C ABI, а не проверка структуры DTO: неверный layout повреждает память.
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if unsafe.Sizeof(inetOption{}) != 16 || unsafe.Offsetof(inetOption{}.Value) != 8 || unsafe.Sizeof(inetList{}) != 32 {
			t.Fatal("invalid WinINet x64 layout")
		}
	}
}

func TestIdentityOfExitedProcessWithRetainedHandle(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child, err := (Runner{}).Start(context.Background(), ports.Command{Path: executable, Args: []string{"-test.run=^TestWindowsJobHelperChild$", "--", "job-child"}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer child.Stop(ctx)
	pid, err := windows.GetProcessId(child.(*process).handle)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := ProcessIdentity(int(pid))
	if err != nil {
		t.Fatal(err)
	}
	retained, err := windows.OpenProcess(windows.SYNCHRONIZE, false, pid)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(retained)
	if err := child.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if alive, err := identity.Alive(); err != nil || alive {
		t.Fatalf("exited process should be dead even while its object is retained: %v, %v", alive, err)
	}
}

func TestWindowsJobHelperChild(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "job-child" {
		return
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestWindowsJobHelperOwner(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "job-owner" {
		return
	}
	executable, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}
	child, err := (Runner{}).Start(context.Background(), ports.Command{Path: executable, Args: []string{"-test.run=^TestWindowsJobHelperChild$", "--", "job-child"}})
	if err != nil {
		os.Exit(3)
	}
	p := child.(*process)
	id, err := windows.GetProcessId(p.handle)
	if err != nil || id == 0 {
		os.Exit(4)
	}
	identity, err := ProcessIdentity(int(id))
	if err != nil {
		os.Exit(5)
	}
	data, _ := json.Marshal(identity)
	if err := os.WriteFile(os.Args[len(os.Args)-1], data, 0600); err != nil {
		os.Exit(6)
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestWindowsLastJobHandleClosesOnOwnerCrash(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ready := filepath.Join(t.TempDir(), "child.json")
	// Owner вне создаваемого им Job; только его native handle удерживает ребёнка.
	owner, err := startWindows(context.Background(), ports.Command{Path: executable, Args: []string{"-test.run=^TestWindowsJobHelperOwner$", "--", "job-owner", ready}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = owner.Stop(ctx)
	}()
	deadline := time.Now().Add(5 * time.Second)
	var identity Identity
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(ready); err == nil {
			if err := json.Unmarshal(data, &identity); err != nil {
				t.Fatal(err)
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if identity.PID == 0 {
		t.Fatal("owner did not publish child")
	}
	// Retain the child's handle before killing its owner. Reopening its PID
	// during termination can return ACCESS_DENIED; the retained handle proves
	// that this exact child exits without races with PID reuse or image queries.
	child, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(identity.PID))
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(child)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := owner.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	result, err := windows.WaitForSingleObject(child, 3000)
	if err != nil || result != windows.WAIT_OBJECT_0 {
		t.Fatalf("Job child survived owner crash: wait result %v, error %v", result, err)
	}
}
