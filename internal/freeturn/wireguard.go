package freeturn

// WireGuardSettings содержит данные для WG/AWG-подключения через FreeTurn.
type WireGuardSettings struct {
	// Config — исходный текст конфигурации WireGuard или AmneziaWG с ключами, адресами и параметрами соединения.
	// Интеграция с mihomo разбирает его при подготовке конфига.
	Config string
}
