// Package proxymode задаёт способ использования подключения приложениями.
package proxymode

type Mode string

const (
	Local  Mode = "local_proxy"
	System Mode = "system_proxy"
	Tun    Mode = "tun"
)

func (m Mode) Valid() bool { return m == Local || m == System || m == Tun }
