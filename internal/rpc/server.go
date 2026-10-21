// Package rpc предоставляет локальный gRPC API поверх существующего backend.
package rpc

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"

	"wireglow/internal/application/controller"
	"wireglow/internal/backend"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
	"wireglow/internal/freeturn"
	"wireglow/internal/infrastructure/persistence"
	pb "wireglow/internal/rpc/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const TokenMetadataKey = "x-wireglow-session-token"
const MaxMessageBytes = 1024 * 1024

// Options задаёт идентичность процесса и его lifetime, независимый от RPC клиента.
type Options struct {
	Context context.Context
	Token   []byte // 32 случайных байта; в metadata передаются lowercase hex.
	BuildID string
	Epoch   string
}

type service struct {
	pb.UnimplementedControllerServiceServer
	backend  *backend.Backend
	options  Options
	mu       sync.Mutex
	running  chan struct{}
	cancel   context.CancelFunc
	stopping bool
}

// NewServer не открывает endpoint. Владелец вызывает Serve с закрытым UDS listener,
// сохраняет launcher lease и при выходе останавливает сервер и закрывает backend.
// Управление Controller должно идти через этот адаптер, кроме Shutdown владельца.
func NewServer(b *backend.Backend, o Options) (*grpc.Server, error) {
	if b == nil || b.Controller == nil || b.Profiles == nil || o.Context == nil || len(o.Token) != 32 || o.BuildID == "" || o.Epoch == "" {
		return nil, errors.New("rpc: incomplete server options")
	}
	o.Token = append([]byte(nil), o.Token...)
	s := &service{backend: b, options: o}
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(MaxMessageBytes), grpc.MaxSendMsgSize(MaxMessageBytes),
		grpc.UnaryInterceptor(s.unaryAuth), grpc.StreamInterceptor(s.streamAuth),
	)
	pb.RegisterControllerServiceServer(server, s)
	return server, nil
}

func (s *service) authenticate(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get(TokenMetadataKey)
	if len(values) != 1 || len(values[0]) != 64 {
		return status.Error(codes.Unauthenticated, "session token required")
	}
	token, err := hex.DecodeString(values[0])
	if err != nil || subtle.ConstantTimeCompare(token, s.options.Token) != 1 {
		return status.Error(codes.Unauthenticated, "invalid session token")
	}
	if s.options.Context.Err() != nil {
		return status.Error(codes.Unavailable, "controller stopping")
	}
	return nil
}
func (s *service) unaryAuth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := s.authenticate(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}
func (s *service) streamAuth(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := s.authenticate(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}
func (s *service) Hello(context.Context, *emptypb.Empty) (*pb.HelloResponse, error) {
	return pb.HelloResponse_builder{BuildId: s.options.BuildID, Epoch: s.options.Epoch}.Build(), nil
}

func parseMode(mode pb.ProxyMode) (proxymode.Mode, error) {
	switch mode {
	case pb.ProxyMode_PROXY_MODE_LOCAL:
		return proxymode.Local, nil
	case pb.ProxyMode_PROXY_MODE_SYSTEM:
		return proxymode.System, nil
	case pb.ProxyMode_PROXY_MODE_TUN:
		return proxymode.Tun, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid proxy mode")
	}
}
func modeMessage(mode proxymode.Mode) pb.ProxyMode {
	switch mode {
	case proxymode.Local:
		return pb.ProxyMode_PROXY_MODE_LOCAL
	case proxymode.System:
		return pb.ProxyMode_PROXY_MODE_SYSTEM
	case proxymode.Tun:
		return pb.ProxyMode_PROXY_MODE_TUN
	default:
		return pb.ProxyMode_PROXY_MODE_UNSPECIFIED
	}
}
func settingsMessage(s backend.Settings) *pb.ClientSettings {
	return pb.ClientSettings_builder{PreferredMode: modeMessage(s.PreferredMode), ProxyAddress: s.ProxyAddress, RelayAddress: s.RelayAddress, ApiPort: uint32(s.APIPort)}.Build()
}
func (s *service) GetSettings(context.Context, *emptypb.Empty) (*pb.ClientSettings, error) {
	settings, err := s.backend.GetSettings()
	if err != nil {
		return nil, rpcError(err)
	}
	return settingsMessage(settings), nil
}
func (s *service) SetSettings(ctx context.Context, req *pb.ClientSettings) (*pb.ClientSettings, error) {
	mode, err := parseMode(req.GetPreferredMode())
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, rpcError(err)
	}
	if err := s.backend.SetSettings(backend.Settings{PreferredMode: mode, ProxyAddress: req.GetProxyAddress(), RelayAddress: req.GetRelayAddress(), APIPort: int(req.GetApiPort())}); err != nil {
		return nil, rpcError(err)
	}
	return s.GetSettings(ctx, &emptypb.Empty{})
}
func (s *service) GetCapabilities(ctx context.Context, _ *emptypb.Empty) (*pb.Capabilities, error) {
	return pb.Capabilities_builder{SystemProxyUnavailableReason: s.backend.SystemProxyUnavailable(ctx), TunUnavailableReason: s.backend.TunUnavailable(ctx)}.Build(), nil
}

func (s *service) ListProfiles(ctx context.Context, _ *emptypb.Empty) (*pb.ListProfilesResponse, error) {
	profiles, err := s.backend.Profiles.List(ctx)
	if err != nil {
		return nil, rpcError(err)
	}
	result := make([]*pb.ProfileSummary, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, summary(p))
	}
	return pb.ListProfilesResponse_builder{Profiles: result}.Build(), nil
}
func validateImport(req *pb.ImportProfileRequest) (*freeturn.Connection, error) {
	if len(req.GetFreeturnLink()) > 512*1024 {
		return nil, status.Error(codes.InvalidArgument, "import too large")
	}
	c, err := freeturn.Parse(req.GetFreeturnLink())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid FreeTurn link")
	}
	return c, nil
}
func (s *service) PreviewImport(ctx context.Context, req *pb.ImportProfileRequest) (*pb.ProfileSummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, rpcError(err)
	}
	c, err := validateImport(req)
	if err != nil {
		return nil, err
	}
	return summary((profile.Profile{Connection: *c}).Summary()), nil
}
func (s *service) ImportProfile(ctx context.Context, req *pb.ImportProfileRequest) (*pb.ProfileSummary, error) {
	if _, err := validateImport(req); err != nil {
		return nil, err
	}
	p, err := s.backend.Profiles.Import(ctx, req.GetFreeturnLink())
	if err != nil {
		return nil, rpcError(err)
	}
	return summary(p), nil
}
func validProfileRequest(req interface{ GetId() string }) error {
	id := req.GetId()
	decoded, err := hex.DecodeString(id)
	if err != nil || len(decoded) != 16 || hex.EncodeToString(decoded) != id {
		return status.Error(codes.InvalidArgument, "invalid profile ID")
	}
	return nil
}
func (s *service) GetProfile(ctx context.Context, req *pb.ProfileRequest) (*pb.Profile, error) {
	if err := validProfileRequest(req); err != nil {
		return nil, err
	}
	p, err := s.backend.Profiles.Get(ctx, req.GetId())
	if err != nil {
		return nil, rpcError(err)
	}
	return profileMessage(p), nil
}
func (s *service) UpdateProfile(ctx context.Context, req *pb.ProfileEdits) (*pb.ProfileSummary, error) {
	if err := validProfileRequest(req); err != nil {
		return nil, err
	}
	p, err := s.backend.Profiles.Get(ctx, req.GetId())
	if err != nil {
		return nil, rpcError(err)
	}
	name := strings.TrimSpace(req.GetName())
	if name == "" || len([]rune(name)) > 80 || req.GetStreams() < 1 || req.GetStreamsPerCredential() < 1 {
		return nil, status.Error(codes.InvalidArgument, "name and stream counts must be valid")
	}
	p.Connection.Name = name
	p.Connection.Streams = int(req.GetStreams())
	p.Connection.StreamsPerCredential = int(req.GetStreamsPerCredential())
	if p.Connection.Provider == freeturn.ProviderVK {
		links := make([]string, 0, len(req.GetCallLinks()))
		for _, link := range req.GetCallLinks() {
			if link = strings.TrimSpace(link); link != "" {
				links = append(links, link)
			}
		}
		if len(links) == 0 {
			return nil, status.Error(codes.InvalidArgument, "at least one call link is required")
		}
		if p.Connection.VK == nil {
			p.Connection.VK = &freeturn.VKSettings{}
		}
		p.Connection.VK.CallLinks = links
	}
	if err := s.backend.Profiles.Update(ctx, p); err != nil {
		return nil, rpcError(err)
	}
	return summary(p.Summary()), nil
}
func (s *service) DeleteProfile(ctx context.Context, req *pb.ProfileRequest) (*emptypb.Empty, error) {
	if err := validProfileRequest(req); err != nil {
		return nil, err
	}
	if err := s.backend.Profiles.Delete(ctx, req.GetId()); err != nil {
		return nil, rpcError(err)
	}
	return &emptypb.Empty{}, nil
}
func (s *service) Connect(ctx context.Context, req *pb.ConnectRequest) (*emptypb.Empty, error) {
	if err := validProfileRequest(req); err != nil {
		return nil, err
	}
	mode, err := parseMode(req.GetMode())
	if err != nil {
		return nil, err
	}
	if mode == proxymode.System {
		if reason := s.backend.SystemProxyUnavailable(ctx); reason != "" {
			return nil, status.Error(codes.FailedPrecondition, reason)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.backend.Controller.Snapshot().State
	if s.running != nil || s.stopping || (state != controller.Idle && state != controller.Failed) {
		return nil, status.Error(codes.FailedPrecondition, "session already active")
	}
	p, err := s.backend.Profiles.Get(ctx, req.GetId())
	if err != nil {
		return nil, rpcError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, rpcError(err)
	}
	if s.options.Context.Err() != nil {
		return nil, status.Error(codes.Unavailable, "controller stopping")
	}
	connectCtx, cancel := context.WithCancel(s.options.Context)
	done := make(chan struct{})
	s.running, s.cancel = done, cancel
	go func() {
		// Controller публикует очищенную ошибку и выполняет cleanup при отказе.
		_ = s.backend.Controller.ConnectWithMode(connectCtx, p, mode)
		cancel()
		s.mu.Lock()
		s.running, s.cancel = nil, nil
		close(done)
		s.mu.Unlock()
	}()
	return &emptypb.Empty{}, nil
}
func (s *service) Disconnect(ctx context.Context, _ *emptypb.Empty) (*pb.Snapshot, error) {
	s.mu.Lock()
	if s.stopping {
		s.mu.Unlock()
		return nil, status.Error(codes.Aborted, "disconnect already running")
	}
	s.stopping = true
	done := s.running
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.stopping = false; s.mu.Unlock() }()
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return nil, rpcError(ctx.Err())
		}
	}
	if err := s.backend.Controller.Disconnect(); err != nil {
		return nil, rpcError(err)
	}
	return snapshot(s.backend.Controller.Snapshot()), nil
}
func (s *service) SwitchProxyMode(ctx context.Context, req *pb.ProxyModeRequest) (*pb.Snapshot, error) {
	mode, err := parseMode(req.GetMode())
	if err != nil {
		return nil, err
	}
	if mode == proxymode.System {
		if reason := s.backend.SystemProxyUnavailable(ctx); reason != "" {
			return nil, status.Error(codes.FailedPrecondition, reason)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running != nil || s.stopping {
		return nil, status.Error(codes.FailedPrecondition, "session operation running")
	}
	if err := s.backend.Controller.SwitchMode(ctx, mode); err != nil {
		return nil, rpcError(err)
	}
	return snapshot(s.backend.Controller.Snapshot()), nil
}

func (s *service) CheckLatency(ctx context.Context, _ *emptypb.Empty) (*pb.Snapshot, error) {
	if err := s.backend.Controller.CheckLatency(ctx); err != nil {
		if errors.Is(err, controller.ErrBusy) {
			return nil, status.Error(codes.FailedPrecondition, "latency check unavailable or running")
		}
		return nil, rpcError(err)
	}
	return snapshot(s.backend.Controller.Snapshot()), nil
}

func (s *service) GetSnapshot(context.Context, *emptypb.Empty) (*pb.Snapshot, error) {
	return snapshot(s.backend.Controller.Snapshot()), nil
}
func (s *service) WatchSnapshots(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.Snapshot]) error {
	events, unsubscribe := s.backend.Controller.Subscribe()
	defer unsubscribe()
	for {
		select {
		case <-stream.Context().Done():
			return rpcError(stream.Context().Err())
		case <-s.options.Context.Done():
			return status.Error(codes.Unavailable, "controller stopping")
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if err := stream.Send(snapshot(event)); err != nil {
				return err
			}
		}
	}
}

// Не пересылаем ошибки с путями, URI, ключами и выводом дочерних процессов.
func rpcError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	case errors.Is(err, persistence.ErrNotFound):
		return status.Error(codes.NotFound, "profile not found")
	case errors.Is(err, controller.ErrModeSwitch):
		return status.Error(codes.FailedPrecondition, "proxy mode switch failed; see connection status")
	case errors.Is(err, controller.ErrBusy):
		return status.Error(codes.FailedPrecondition, "session already active")
	case errors.Is(err, controller.ErrClosed):
		return status.Error(codes.Unavailable, "controller closed")
	default:
		return status.Error(codes.Internal, "backend operation failed")
	}
}
