package freeturn

const (
	// ProviderVK направляет соединение через TURN сервиса звонков VK.
	ProviderVK = "vk"
	// ProviderDirect направляет соединение прямо на сервер FreeTurn.
	ProviderDirect = "direct"
)

const (
	// TurnTransportTCP задаёт подключение к TURN по TCP/TLS.
	TurnTransportTCP = "tcp"
	// TurnTransportUDP задаёт подключение к TURN по UDP.
	TurnTransportUDP = "udp"
	// RelayModeTCP задаёт передачу потока байтов через KCP и smux.
	RelayModeTCP = "tcp"
	// RelayModeUDP задаёт передачу датаграмм.
	RelayModeUDP = "udp"
)

// Connection содержит настройки соединения FreeTurn и WG/AWG-конфиг из ссылки.
type Connection struct {
	// Name — название подключения для отображения пользователю.
	Name string
	// ClientID — идентификатор клиента, передаваемый серверу FreeTurn.
	// Сервер использует его для учёта подключений и проверки списка разрешённых клиентов.
	ClientID string
	// Provider выбирает способ выхода к серверу: vk — через TURN сервиса звонков, direct — напрямую.
	Provider string
	// TurnTransport выбирает транспорт между клиентом и TURN-сервером: tcp или udp.
	// Он не определяет протокол трафика внутри relay.
	TurnTransport string
	// RelayMode задаёт передачу датаграмм (udp) или потока байтов (tcp).
	// FreeTurn передаёт трафик выбранным способом, не разбирая внутренний протокол.
	RelayMode string
	// VK содержит ссылки звонков и настройки прохождения captcha для провайдера VK.
	// nil означает отсутствие настроек VK; для провайдера direct они не нужны.
	VK *VKSettings
	// DNS задаёт разрешение имён для соединений самого FreeTurn.
	DNS DNSSettings
	// Streams задаёт число параллельных relay-сессий на одну ссылку звонка; в direct — общее число сессий.
	// Каждая сессия создаёт отдельный DTLS-канал, а при работе через TURN — отдельную TURN allocation.
	Streams int
	// StreamsPerCredential задаёт число TURN-сессий на один набор учётных данных VK.
	// Определяет, сколько сессий используют одни и те же данные доступа к TURN.
	StreamsPerCredential int
	// Bond распределяет данные одного TCP-соединения между несколькими relay-сессиями.
	// На принимающей стороне части собираются обратно в упорядоченный поток.
	Bond bool
	// KCP задаёт подтверждения и повторные отправки для TCP relay.
	// nil означает, что отдельные настройки KCP не заданы.
	KCP *KCPSettings
	// WG содержит исходный конфиг WireGuard или AmneziaWG для настройки mihomo.
	// nil означает отсутствие WG/AWG-конфига.
	WG *WireGuardSettings
	// Peer — адрес сервера FreeTurn в формате host:port, включая доменное имя или IP.
	// Это адрес получателя relay-трафика, а не адрес бэкенда за сервером.
	Peer string
	// Obfuscation задаёт формат маскировки, общий ключ и интервалы отправки пакетов FreeTurn.
	Obfuscation Obfuscation
}
