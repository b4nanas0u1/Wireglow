package freeturn

// VKSettings содержит настройки получения TURN-доступа через звонки VK.
type VKSettings struct {
	// CallLinks содержит ссылки звонков VK, через которые ядро получает доступ к TURN.
	// Несколько ссылок позволяют создавать relay-сессии через разные звонки.
	CallLinks []string
	// ManualCaptcha отключает автоматическое решение captcha и сразу предлагает ручное, если captcha возникла.
	// При false ядро сначала пытается решить её автоматически.
	ManualCaptcha bool
}
