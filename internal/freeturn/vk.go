package freeturn

const (
	// VKPlatformDesktop selects the desktop VK authorization persona.
	VKPlatformDesktop = "desktop"
	// VKPlatformMobile selects the mobile VK authorization persona.
	VKPlatformMobile = "mobile"
)

// VKSettings содержит настройки получения TURN-доступа через звонки VK.
type VKSettings struct {
	// CallLinks содержит ссылки звонков VK, через которые ядро получает доступ к TURN.
	// Несколько ссылок позволяют создавать relay-сессии через разные звонки.
	CallLinks []string
	// ManualCaptcha отключает автоматическое решение captcha и сразу предлагает ручное, если captcha возникла.
	// При false ядро сначала пытается решить её автоматически.
	ManualCaptcha bool
	// Platform selects the VK authorization persona: desktop or mobile.
	// Empty preserves the engine default (desktop).
	Platform string
	// TurnAddress optionally overrides the TURN host and port supplied by VK.
	// Credentials are still obtained through VK. Empty preserves provider selection.
	TurnAddress string
}
