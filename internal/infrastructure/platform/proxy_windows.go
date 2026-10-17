package platform

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"unsafe"
	"wireglow/internal/infrastructure/network"

	"golang.org/x/sys/windows"
)

type ProxySettings struct{}
type proxyValue struct {
	Flags   uint32 `json:"flags"`
	Server  string `json:"server"`
	Bypass  string `json:"bypass"`
	AutoURL string `json:"autoUrl"`
}
type inetOption struct {
	Option uint32
	Value  uint64
}
type inetList struct {
	Size       uint32
	Connection *uint16
	Count      uint32
	Error      uint32
	Options    *inetOption
}

var wininet = windows.NewLazySystemDLL("wininet.dll")
var inetQuery = wininet.NewProc("InternetQueryOptionW")
var inetSet = wininet.NewProc("InternetSetOptionW")
var globalFree = windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalFree")

func (ProxySettings) Desired(address string) ([]network.Change, error) {
	if _, err := ProxyPort(address); err != nil {
		return nil, err
	}
	data, _ := json.Marshal(proxyValue{Flags: 3, Server: "http=" + address + ";https=" + address, Bypass: "localhost;127.*;[::1]"})
	return []network.Change{{Key: "wininet", After: string(data)}}, nil
}

func (ProxySettings) Read(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if key != "wininet" {
		return "", errors.New("proxy: unsupported setting")
	}
	// FLAGS_UI отражает настройки пользователя, включая autodetect.
	options := []inetOption{{Option: 10}, {Option: 2}, {Option: 3}, {Option: 4}}
	list := inetList{Size: uint32(unsafe.Sizeof(inetList{})), Count: 4, Options: &options[0]}
	size := list.Size
	ok, _, _ := inetQuery.Call(0, 75, uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Pointer(&size)))
	if ok == 0 {
		return "", errors.New("proxy: WinINet query failed")
	}
	texts := make([]string, 3)
	for i := 1; i < 4; i++ {
		ptr := *(*unsafe.Pointer)(unsafe.Pointer(&options[i].Value))
		if ptr != nil {
			texts[i-1] = windows.UTF16PtrToString((*uint16)(ptr))
			globalFree.Call(uintptr(ptr))
		}
	}
	data, _ := json.Marshal(proxyValue{uint32(options[0].Value), texts[0], texts[1], texts[2]})
	return string(data), nil
}

func (ProxySettings) Write(ctx context.Context, key, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key != "wininet" {
		return errors.New("proxy: unsupported setting")
	}
	var v proxyValue
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return errors.New("proxy: invalid snapshot")
	}
	texts := []string{v.Server, v.Bypass, v.AutoURL}
	ptrs := make([]*uint16, 3)
	options := []inetOption{{Option: 1, Value: uint64(v.Flags)}, {Option: 2}, {Option: 3}, {Option: 4}}
	for i, text := range texts {
		p, err := windows.UTF16PtrFromString(text)
		if err != nil {
			return err
		}
		ptrs[i] = p
		options[i+1].Value = uint64(uintptr(unsafe.Pointer(p)))
	}
	list := inetList{Size: uint32(unsafe.Sizeof(inetList{})), Count: 4, Options: &options[0]}
	ok, _, _ := inetSet.Call(0, 75, uintptr(unsafe.Pointer(&list)), uintptr(list.Size))
	runtime.KeepAlive(ptrs)
	if ok == 0 {
		return errors.New("proxy: WinINet update denied")
	}
	for _, option := range []uintptr{39, 37} {
		ok, _, _ := inetSet.Call(0, option, 0, 0)
		if ok == 0 {
			return errors.New("proxy: WinINet notification failed")
		}
	}
	return nil
}

func SystemProxyUnavailable(ctx context.Context) string {
	if _, err := (ProxySettings{}).Read(ctx, "wininet"); err != nil {
		return "Windows proxy settings are unavailable"
	}
	return ""
}
