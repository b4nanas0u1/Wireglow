package profile

import "wireglow/internal/freeturn"

// Profile — сохранённое подключение клиента; ID не относится к серверу FreeTurn.
type Profile struct {
	ID         string              `json:"id"`
	Connection freeturn.Connection `json:"connection"`
}

// Summary не раскрывает ключи и исходные ссылки.
type Summary struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Provider             string `json:"provider"`
	Peer                 string `json:"peer"`
	TurnTransport        string `json:"turnTransport"`
	RelayMode            string `json:"relayMode"`
	Streams              int    `json:"streams"`
	StreamsPerCredential int    `json:"streamsPerCredential"`
	ObfuscationMode      string `json:"obfuscationMode"`
}

func (p Profile) Summary() Summary {
	c := p.Connection
	return Summary{ID: p.ID, Name: c.Name, Provider: c.Provider, Peer: c.Peer, TurnTransport: c.TurnTransport, RelayMode: c.RelayMode, Streams: c.Streams, StreamsPerCredential: c.StreamsPerCredential, ObfuscationMode: c.Obfuscation.Mode}
}
