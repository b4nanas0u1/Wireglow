package freeturn

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"
)

func TestParseSessionExample(t *testing.T) {
	// Обезличенная копия примера из сессии «Распарсить ссылку FreeTurn».
	const wg = `[Interface]
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Address = 10.13.13.3/32
DNS = 1.1.1.1
[Peer]
PublicKey = AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=
AllowedIPs = 0.0.0.0/0
Endpoint = 127.0.0.1:9000
PersistentKeepalive = 25`
	got, err := Parse(testLink(`{
		"v":1,"provider":"vk","peer":"192.0.2.1:56035","obf":"rtpopus",
		"key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"n":12,"spc":12,"cid":"0123456789abcdef0123456789abcdef","name":"sample",
		"vk":"https://vk.com/call/join/test-call",
		"wg":"[Interface]\nPrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\nAddress = 10.13.13.3/32\nDNS = 1.1.1.1\n[Peer]\nPublicKey = AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=\nAllowedIPs = 0.0.0.0/0\nEndpoint = 127.0.0.1:9000\nPersistentKeepalive = 25"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	want := &Connection{
		Name: "sample", ClientID: strings.Repeat("0123456789abcdef", 2),
		Provider: ProviderVK, Peer: "192.0.2.1:56035",
		Streams: 12, StreamsPerCredential: 12,
		VK:          &VKSettings{CallLinks: []string{"https://vk.com/call/join/test-call"}},
		WG:          &WireGuardSettings{Config: wg},
		Obfuscation: Obfuscation{Mode: ObfuscationModeRTPOpus, Key: strings.Repeat("0123456789abcdef", 4)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %+v, want %+v", got, want)
	}
}

func TestParseAllSettings(t *testing.T) {
	got, err := Parse(testLink(`{
		"v":1,"provider":"vk","peer":"relay.example:56000","name":"Сервер",
		"cid":"test-id","transport":"udp","mode":"tcp","n":3,"spc":2,"bond":true,
		"obf":"rtpopus3","key":"test-key","timing":20,
		"dns":"auto","dnss":" 1.1.1.1, ,8.8.8.8:53 ",
		"vk":" https://vk.com/call/join/one,https://vk.ru/call/join/two ","mcap":true,
		"wg":"[Interface]\nJc = 4\n","listen":"0.0.0.0:9000",
		"kcp":{"nodelay":1,"interval":40,"resend":2,"nc":1,
		"sndwnd":256,"rcvwnd":512,"mtu":1100,"acknodelay":true}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	want := &Connection{
		Name: "Сервер", ClientID: "test-id", Provider: ProviderVK, Peer: "relay.example:56000",
		TurnTransport: TurnTransportUDP, RelayMode: RelayModeTCP,
		Streams: 3, StreamsPerCredential: 2, Bond: true,
		Obfuscation: Obfuscation{Mode: ObfuscationModeRTPOpus3, Key: "test-key", TimingMs: 20},
		DNS:         DNSSettings{Mode: DNSModeAuto, Servers: []string{"1.1.1.1", "8.8.8.8:53"}},
		VK:          &VKSettings{CallLinks: []string{"https://vk.com/call/join/one", "https://vk.ru/call/join/two"}, ManualCaptcha: true},
		WG:          &WireGuardSettings{Config: "[Interface]\nJc = 4\n"},
		KCP:         &KCPSettings{NoDelay: 1, Interval: 40, Resend: 2, NC: 1, SndWnd: 256, RcvWnd: 512, MTU: 1100, ACKNoDelay: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %+v, want %+v", got, want)
	}
}

func TestParseOptionalSettings(t *testing.T) {
	got, err := Parse(testLink(`{"v":1,"provider":"direct","peer":"[2001:db8::1]:56000"}`))
	if err != nil {
		t.Fatal(err)
	}
	want := &Connection{Provider: ProviderDirect, Peer: "[2001:db8::1]:56000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %+v, want %+v", got, want)
	}
	got, err = Parse(testLink(`{"v":1,"provider":"direct","peer":"relay.example:56000","kcp":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.VK != nil || got.KCP == nil || *got.KCP != (KCPSettings{}) {
		t.Fatalf("explicit empty settings lost: %+v", got)
	}
}

func TestParseVKSettings(t *testing.T) {
	for name, fields := range map[string]string{
		"absent": "", "empty": `,"vk":""`, "blank": `,"vk":" , , "`,
		"captcha only": `,"mcap":true`, "blank with captcha": `,"vk":" , ","mcap":true`,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := Parse(testLink(`{"v":1,"provider":"vk","peer":"relay:56000"` + fields + `}`))
			if got != nil || err == nil || err.Error() != "freeturn: missing VK call links" {
				t.Fatalf("Parse() = %+v, %v; want missing VK call links", got, err)
			}
		})
	}
	for name, fields := range map[string]string{
		"links":   `,"vk":"https://vk.com/call/join/test-call"`,
		"captcha": `,"mcap":true`,
		"both":    `,"vk":"https://vk.com/call/join/test-call","mcap":true`,
	} {
		t.Run("direct "+name, func(t *testing.T) {
			got, err := Parse(testLink(`{"v":1,"provider":"direct","peer":"relay:56000"` + fields + `}`))
			if err != nil {
				t.Fatal(err)
			}
			if got.VK != nil {
				t.Fatalf("direct connection has VK settings: %+v", got.VK)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := map[string]string{
		"scheme": "https://example.com", "empty": "freeturn://", "base64": "freeturn://!",
		"json": testLink(`not json`), "null": testLink(`null`), "array": testLink(`[]`),
		"version absent":      testLink(`{"provider":"vk","peer":"relay:56000"}`),
		"version unsupported": testLink(`{"v":2,"provider":"vk","peer":"relay:56000"}`),
		"provider absent":     testLink(`{"v":1,"peer":"relay:56000"}`),
		"peer blank":          testLink(`{"v":1,"provider":"vk","peer":" "}`),
		"field type":          testLink(`{"v":1,"provider":"vk","peer":"relay:56000","n":"secret-marker"}`),
		"unknown field":       testLink(`{"v":1,"provider":"vk","peer":"relay:56000","secret-marker":true}`),
		"unknown KCP field":   testLink(`{"v":1,"provider":"vk","peer":"relay:56000","kcp":{"secret-marker":1}}`),
		"trailing object":     testLink(`{"v":1,"provider":"vk","peer":"relay:56000"}{}`),
		"trailing junk":       testLink(`{"v":1,"provider":"vk","peer":"relay:56000"}secret-marker`),
		"oversize":            testLink(strings.Repeat(" ", maxLinkPayloadSize+1)),
		"oversize encoded":    "freeturn://" + strings.Repeat("A", base64.RawURLEncoding.EncodedLen(maxLinkPayloadSize)+1),
	}
	for name, link := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Parse(link)
			if err == nil || got != nil {
				t.Fatalf("Parse() = %+v, %v; want nil and error", got, err)
			}
			if strings.Contains(err.Error(), "secret-marker") {
				t.Fatalf("error leaks input: %v", err)
			}
		})
	}
}

func TestParseSizeBoundary(t *testing.T) {
	data := `{"v":1,"provider":"direct","peer":"relay:56000"}`
	data += strings.Repeat(" ", maxLinkPayloadSize-len(data))
	if _, err := Parse(testLink(data)); err != nil {
		t.Fatalf("256 KiB payload rejected: %v", err)
	}
}

func testLink(data string) string {
	return "freeturn://" + base64.RawURLEncoding.EncodeToString([]byte(data))
}
