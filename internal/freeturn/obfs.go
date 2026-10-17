// Package freeturn содержит настройки подключения к ядру FreeTurn.
package freeturn

const (
	// ObfuscationModeNone отключает обфускацию пакетов.
	ObfuscationModeNone = "none"
	// ObfuscationModeRTPOpus маскирует пакеты под RTP/Opus и защищает их общим ключом.
	ObfuscationModeRTPOpus = "rtpopus"
	// ObfuscationModeRTPOpus2 добавляет к RTP/Opus расширения заголовка WebRTC.
	ObfuscationModeRTPOpus2 = "rtpopus2"
	// ObfuscationModeRTPOpus3 использует расширенный формат маскировки под аудиопоток WebRTC.
	ObfuscationModeRTPOpus3 = "rtpopus3"
)

// Obfuscation задаёт маскировку пакетов FreeTurn под медиатрафик.
type Obfuscation struct {
	// Mode выбирает формат маскировки пакетов; none отключает обфускацию.
	Mode string
	// Key — общий ключ клиента и сервера в виде 64 шестнадцатеричных символов.
	// Он используется для защиты пакетов обфускации и должен совпадать на обеих сторонах.
	Key string
	// TimingMs задаёт базовый интервал отправки пакетов обфускации в миллисекундах.
	// Ядро добавляет вариации задержки; ноль отключает искусственное замедление отправки.
	TimingMs int
}
