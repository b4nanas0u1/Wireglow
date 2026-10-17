// Package proxymode задаёт способ использования подключения приложениями.
package proxymode

type Mode string

const (
	Local  Mode = "local_proxy"
	System Mode = "system_proxy"
)

func (m Mode) Valid() bool { return m == Local || m == System }
