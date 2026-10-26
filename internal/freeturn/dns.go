package freeturn

const (
	// DNSModePlain отправляет DNS-запросы по UDP.
	DNSModePlain = "plain"
	// DNSModeDoH отправляет DNS-запросы по HTTPS.
	DNSModeDoH = "doh"
	// DNSModeAuto использует UDP и переходит на DoH при отказе UDP-резолверов.
	DNSModeAuto = "auto"
)

// DNSSettings задаёт разрешение имён, нужных FreeTurn для установления соединения.
type DNSSettings struct {
	// Mode выбирает способ DNS-запросов: plain — UDP, doh — HTTPS, auto — UDP с переходом на DoH при отказе.
	Mode string
	// Servers содержит адреса UDP DNS-серверов в формате IP или IP:port.
	// Они нужны самому FreeTurn, а не для настройки DNS пользовательского трафика в mihomo.
	Servers []string
}
