package rpc

import (
	"wireglow/internal/application/controller"
	"wireglow/internal/domain/profile"
	pb "wireglow/internal/rpc/gen"
)

func summary(p profile.Summary) *pb.ProfileSummary {
	return pb.ProfileSummary_builder{Id: p.ID, Name: p.Name, Provider: p.Provider, Peer: p.Peer, TurnTransport: p.TurnTransport, RelayMode: p.RelayMode, Streams: int64(p.Streams), StreamsPerCredential: int64(p.StreamsPerCredential), ObfuscationMode: p.ObfuscationMode}.Build()
}
func profileMessage(p profile.Profile) *pb.Profile {
	c := p.Connection
	result := pb.Connection_builder{
		Name: c.Name, ClientId: c.ClientID, Provider: c.Provider, TurnTransport: c.TurnTransport, RelayMode: c.RelayMode,
		Dns:     pb.DNSSettings_builder{Mode: c.DNS.Mode, Servers: append([]string(nil), c.DNS.Servers...)}.Build(),
		Streams: int64(c.Streams), StreamsPerCredential: int64(c.StreamsPerCredential), Bond: c.Bond, Peer: c.Peer,
		Obfuscation: pb.Obfuscation_builder{Mode: c.Obfuscation.Mode, Key: c.Obfuscation.Key, TimingMs: int64(c.Obfuscation.TimingMs)}.Build(),
	}
	if c.VK != nil {
		result.Vk = pb.VKSettings_builder{CallLinks: append([]string(nil), c.VK.CallLinks...), ManualCaptcha: c.VK.ManualCaptcha}.Build()
	}
	if c.KCP != nil {
		k := c.KCP
		result.Kcp = pb.KCPSettings_builder{NoDelay: int64(k.NoDelay), Interval: int64(k.Interval), Resend: int64(k.Resend), Nc: int64(k.NC), SndWnd: int64(k.SndWnd), RcvWnd: int64(k.RcvWnd), Mtu: int64(k.MTU), AckNoDelay: k.ACKNoDelay}.Build()
	}
	if c.WG != nil {
		result.Wg = pb.WireGuardSettings_builder{Config: c.WG.Config}.Build()
	}
	return pb.Profile_builder{Id: p.ID, Connection: result.Build()}.Build()
}
func snapshot(s controller.Snapshot) *pb.Snapshot {
	state := map[controller.State]pb.State{
		controller.Idle: pb.State_STATE_IDLE, controller.Validating: pb.State_STATE_VALIDATING,
		controller.StartingRelay: pb.State_STATE_STARTING_RELAY, controller.StartingMihomo: pb.State_STATE_STARTING_MIHOMO,
		controller.Probing: pb.State_STATE_PROBING, controller.ActivatingNetwork: pb.State_STATE_ACTIVATING_NETWORK,
		controller.Connected: pb.State_STATE_CONNECTED, controller.Stopping: pb.State_STATE_STOPPING, controller.Failed: pb.State_STATE_FAILED,
	}[s.State]
	return pb.Snapshot_builder{ActiveMode: modeMessage(s.ActiveMode), State: state, Generation: s.Generation, ProfileId: s.ProfileID, MixedAddress: s.MixedAddress, Error: s.Error, Latency: pb.Latency_builder{Checking: s.Latency.Checking, DelayMs: s.Latency.DelayMs, CheckedAt: s.Latency.CheckedAt, SuccessAt: s.Latency.SuccessAt, Error: s.Latency.Error}.Build()}.Build()
}
