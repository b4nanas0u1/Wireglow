package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"wireglow/internal/domain/proxymode"
	"wireglow/internal/infrastructure/platform"
)

func TestSettingsPersistAcrossBackendRestart(t *testing.T) {
	root := t.TempDir()
	options := Options{Paths: platform.Paths{Data: filepath.Join(root, "data"), Runtime: filepath.Join(root, "runtime")}}
	b, err := Open(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	settings, err := b.GetSettings()
	if err != nil || settings.PreferredMode != proxymode.Local {
		t.Fatalf("default = %+v, %v", settings, err)
	}
	if err := b.SetSettings(Settings{PreferredMode: proxymode.System}); err != nil {
		t.Fatal(err)
	}
	if err := b.SetSettings(Settings{PreferredMode: "unknown-mode"}); err == nil {
		t.Fatal("unsupported mode accepted")
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	settings, err = restored.GetSettings()
	if err != nil || settings.PreferredMode != proxymode.System {
		t.Fatalf("restored = %+v, %v", settings, err)
	}
	if err := os.WriteFile(restored.settingsPath, []byte(`{"preferredMode":"unknown-mode"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.GetSettings(); err == nil {
		t.Fatal("invalid persisted mode accepted")
	}
}

func TestAddressesMigrationValidationAndReset(t *testing.T) {
	b := &Backend{settingsPath: filepath.Join(t.TempDir(), "client-settings.json")}
	if err := os.WriteFile(b.settingsPath, []byte(`{"preferredMode":"system_proxy"}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := b.GetSettings()
	if err != nil || got.ProxyAddress != DefaultSettings().ProxyAddress || got.RelayAddress != DefaultSettings().RelayAddress || got.PreferredMode != proxymode.System || got.APIPort != 17892 {
		t.Fatalf("legacy settings = %+v, %v", got, err)
	}
	custom := Settings{PreferredMode: proxymode.System, ProxyAddress: "127.0.0.1:28001", RelayAddress: "127.0.0.1:28002", APIPort: 17892}
	if err := b.SetSettings(custom); err != nil {
		t.Fatal(err)
	}
	for _, addresses := range [][2]string{
		{"0.0.0.0:28001", "127.0.0.1:28002"},
		{"127.0.0.1:0", "127.0.0.1:28002"},
		{"127.0.0.1:65536", "127.0.0.1:28002"},
		{"127.0.0.1:28001", "127.0.0.1:028001"},
		{"127.0.0.1:17892", "127.0.0.1:28002"},
		{"127.0.0.1:28001", "127.0.0.1:17892"},
	} {
		value := custom
		value.ProxyAddress, value.RelayAddress = addresses[0], addresses[1]
		if err := b.SetSettings(value); err == nil {
			t.Fatalf("accepted %+v", value)
		}
		got, err := b.GetSettings()
		if err != nil || got != custom {
			t.Fatalf("failed write changed settings: %+v, %v", got, err)
		}
	}
	for _, apiPort := range []int{-1, 65536, 28001, 28002} {
		value := custom
		value.APIPort = apiPort
		if err := b.SetSettings(value); err == nil {
			t.Fatalf("accepted API port %d", apiPort)
		}
		got, err := b.GetSettings()
		if err != nil || got != custom {
			t.Fatalf("invalid API write changed settings: %+v, %v", got, err)
		}
	}
	custom.APIPort = 28003
	custom.ProxyAddress = "127.0.0.1:17892"
	if err := b.SetSettings(custom); err != nil {
		t.Fatalf("17892 should be reusable: %v", err)
	}

	restored := &Backend{settingsPath: b.settingsPath}
	got, err = restored.GetSettings()
	if err != nil || got != custom {
		t.Fatalf("restored = %+v, %v", got, err)
	}
	if err := restored.SetSettings(DefaultSettings()); err != nil {
		t.Fatal(err)
	}
	got, err = b.GetSettings()
	if err != nil || got != DefaultSettings() {
		t.Fatalf("reset = %+v, %v", got, err)
	}
}
