package freeturn

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const (
	maxLinkPayloadSize = 256 * 1024
	linkPrefix         = "freeturn://"
)

// Parse разбирает freeturn:// с JSON v1 в Base64URL без padding.
// Значения не дополняются defaults; WG/AWG остаётся исходным текстом.
// Проверка пригодности настроек для подключения выполняется отдельно.
func Parse(link string) (*Connection, error) {
	link = strings.TrimSpace(link)
	if !strings.HasPrefix(link, linkPrefix) {
		return nil, errors.New("freeturn: expected freeturn:// scheme")
	}
	payload := strings.TrimPrefix(link, linkPrefix)
	if payload == "" {
		return nil, errors.New("freeturn: empty payload")
	}
	if len(payload) > base64.RawURLEncoding.EncodedLen(maxLinkPayloadSize) {
		return nil, errors.New("freeturn: payload exceeds 256 KiB")
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, errors.New("freeturn: invalid base64url payload")
	}
	if len(raw) > maxLinkPayloadSize {
		return nil, errors.New("freeturn: payload exceeds 256 KiB")
	}

	var data linkData
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		// Ошибки JSON могут содержать значения секретных полей.
		return nil, errors.New("freeturn: invalid JSON or unknown field")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("freeturn: trailing JSON data")
	}
	if data.Version != 1 {
		return nil, errors.New("freeturn: unsupported link version")
	}
	if strings.TrimSpace(data.Provider) == "" {
		return nil, errors.New("freeturn: missing provider")
	}
	if strings.TrimSpace(data.Peer) == "" {
		return nil, errors.New("freeturn: missing peer")
	}

	connection := &Connection{
		Name: data.Name, ClientID: data.ClientID,
		Provider: data.Provider, Peer: data.Peer,
		TurnTransport: data.Transport, RelayMode: data.Mode,
		Streams: data.Streams, StreamsPerCredential: data.StreamsPerCredential,
		Bond:        data.Bond,
		DNS:         DNSSettings{Mode: data.DNSMode, Servers: splitLinkList(data.DNSServers)},
		Obfuscation: Obfuscation{Mode: data.Obfuscation, Key: data.Key, TimingMs: data.TimingMs},
	}
	if data.Provider == ProviderVK {
		callLinks := splitLinkList(data.VK)
		if len(callLinks) == 0 {
			return nil, errors.New("freeturn: missing VK call links")
		}
		connection.VK = &VKSettings{CallLinks: callLinks, ManualCaptcha: data.ManualCaptcha}
	}
	if data.WG != "" {
		connection.WG = &WireGuardSettings{Config: data.WG}
	}
	if kcp := data.KCP; kcp != nil {
		connection.KCP = &KCPSettings{
			NoDelay: kcp.NoDelay, Interval: kcp.Interval, Resend: kcp.Resend,
			NC: kcp.NC, SndWnd: kcp.SndWnd, RcvWnd: kcp.RcvWnd,
			MTU: kcp.MTU, ACKNoDelay: kcp.ACKNoDelay,
		}
	}
	return connection, nil
}

// linkData отделяет формат ссылки ядра от DTO подключения.
type linkData struct {
	Version              int          `json:"v"`
	Name                 string       `json:"name"`
	ClientID             string       `json:"cid"`
	Provider             string       `json:"provider"`
	Peer                 string       `json:"peer"`
	Transport            string       `json:"transport"`
	Mode                 string       `json:"mode"`
	Streams              int          `json:"n"`
	StreamsPerCredential int          `json:"spc"`
	Bond                 bool         `json:"bond"`
	DNSMode              string       `json:"dns"`
	DNSServers           string       `json:"dnss"`
	Obfuscation          string       `json:"obf"`
	Key                  string       `json:"key"`
	TimingMs             int          `json:"timing"`
	VK                   string       `json:"vk"`
	ManualCaptcha        bool         `json:"mcap"`
	WG                   string       `json:"wg"`
	KCP                  *linkKCPData `json:"kcp"`
	// Listen принимается для совместимости; локальный адрес выбирает supervisor.
	Listen string `json:"listen"`
}

type linkKCPData struct {
	NoDelay    int  `json:"nodelay"`
	Interval   int  `json:"interval"`
	Resend     int  `json:"resend"`
	NC         int  `json:"nc"`
	SndWnd     int  `json:"sndwnd"`
	RcvWnd     int  `json:"rcvwnd"`
	MTU        int  `json:"mtu"`
	ACKNoDelay bool `json:"acknodelay"`
}

func splitLinkList(value string) []string {
	var result []string
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
