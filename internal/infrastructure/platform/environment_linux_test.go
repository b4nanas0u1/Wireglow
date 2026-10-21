package platform

import (
	"strings"
	"testing"
)

func TestRelayEnvironmentStripsInheritedProxies(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:17891")
	t.Setenv("https_proxy", "http://127.0.0.1:17891")
	t.Setenv("ALL_PROXY", "socks5://127.0.0.1:17891")
	t.Setenv("WIREGLOW_ENV_TEST", "retained")
	found := false
	for _, entry := range withoutProxyEnvironment() {
		key, _, _ := strings.Cut(entry, "=")
		switch strings.ToLower(key) {
		case "http_proxy", "https_proxy", "all_proxy", "no_proxy":
			t.Fatal("inherited proxy retained")
		}
		if entry == "WIREGLOW_ENV_TEST=retained" {
			found = true
		}
	}
	if !found {
		t.Fatal("unrelated environment removed")
	}
}
