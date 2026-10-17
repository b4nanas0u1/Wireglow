package backend

import (
	"context"
	"errors"
	"os"

	"wireglow/internal/domain/proxymode"
	"wireglow/internal/infrastructure/engines"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
)

type Settings struct {
	PreferredMode proxymode.Mode `json:"preferredMode"`
	ProxyAddress  string         `json:"proxyAddress"`
	RelayAddress  string         `json:"relayAddress"`
	APIPort       int            `json:"apiPort"`
}

func (b *Backend) GetSettings() (Settings, error) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	result := DefaultSettings()
	err := persistence.ReadJSON(b.settingsPath, &result)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return Settings{}, err
	}
	if err := validateSettings(result); err != nil {
		return Settings{}, err
	}
	return result, nil
}

func (b *Backend) SetSettings(value Settings) error {
	value = settingsDefaults(value)
	if err := validateSettings(value); err != nil {
		return err
	}
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	return persistence.WriteJSON(b.settingsPath, value)
}

// Проверка доступности без изменений ОС; активация отдельно проверяет recovery.
func (b *Backend) SystemProxyUnavailable(ctx context.Context) string {
	return platform.SystemProxyUnavailable(ctx)
}

func DefaultSettings() Settings {
	return Settings{PreferredMode: proxymode.Local, ProxyAddress: engines.DefaultProxyAddress, RelayAddress: engines.DefaultRelayAddress, APIPort: engines.DefaultAPIPort}
}

// Empty endpoints preserve compatibility with clients that only set the mode.
func settingsDefaults(s Settings) Settings {
	if s.APIPort == 0 {
		s.APIPort = engines.DefaultAPIPort
	}
	if s.ProxyAddress == "" {
		s.ProxyAddress = engines.DefaultProxyAddress
	}
	if s.RelayAddress == "" {
		s.RelayAddress = engines.DefaultRelayAddress
	}
	return s
}
func validateSettings(s Settings) error {
	if !s.PreferredMode.Valid() {
		return errors.New("settings: invalid proxy mode")
	}
	return engines.ValidateAddresses(s.RelayAddress, s.ProxyAddress, s.APIPort)
}
