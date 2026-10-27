package engines

import (
	"slices"
	"testing"
	"wireglow/internal/freeturn"
)

func TestVKEditorRelayArguments(t *testing.T) {
	c := freeturn.Connection{Provider: "vk", Peer: "peer.example:56000", VK: &freeturn.VKSettings{CallLinks: []string{"https://vk.ru/call/join/synthetic"}, Platform: freeturn.VKPlatformMobile, ManualCaptcha: true, TurnAddress: "[2001:db8::1]:443"}}
	args, err := RelayArgs(c, "127.0.0.1:17890")
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"-turn", "2001:db8::1"}, {"-port", "443"}, {"-platform", "mobile"}, {"-transport", "tcp"}} {
		i := slices.Index(args, pair[0])
		if i < 0 || i+1 == len(args) || args[i+1] != pair[1] {
			t.Fatalf("missing %s", pair[0])
		}
	}
	if !slices.Contains(args, "-manual-captcha=true") {
		t.Fatal("manual CAPTCHA lost")
	}
	c.VK.TurnAddress = "turn.example.org"
	c.VK.Platform = ""
	args, err = RelayArgs(c, "127.0.0.1:17890")
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(args, "-port") || slices.Contains(args, "-platform") {
		t.Fatal("defaults should remain with engine")
	}
	c.VK.TurnAddress = ""
	args, err = RelayArgs(c, "127.0.0.1:17890")
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(args, "-turn") {
		t.Fatal("provider selection overridden")
	}
	c.VK.Platform = "invalid"
	if _, err := RelayArgs(c, "127.0.0.1:17890"); err == nil {
		t.Fatal("invalid platform accepted")
	}
}
