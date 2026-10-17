package rpc_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"wireglow/internal/application/controller"
	"wireglow/internal/application/ports"
	"wireglow/internal/backend"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
	"wireglow/internal/rpc"
	pb "wireglow/internal/rpc/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

type process struct {
	done chan struct{}
	once sync.Once
}

func (p *process) Done() <-chan struct{}      { return p.done }
func (*process) Err() error                   { return nil }
func (p *process) Stop(context.Context) error { p.once.Do(func() { close(p.done) }); return nil }

type runner struct{ starts atomic.Int32 }

func (r *runner) Start(ctx context.Context, _ ports.Command) (ports.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.starts.Add(1)
	return &process{done: make(chan struct{})}, nil
}

type preparer struct {
	prepare func(context.Context, profile.Profile) (*ports.Plan, error)
}

func (p preparer) Prepare(ctx context.Context, v profile.Profile) (*ports.Plan, error) {
	return p.prepare(ctx, v)
}

type network struct {
	recoveries atomic.Int32
	applies    atomic.Int32
}

func (n *network) Recover(context.Context) error       { n.recoveries.Add(1); return nil }
func (n *network) Apply(context.Context, string) error { n.applies.Add(1); return nil }

type fixture struct {
	client     pb.ControllerServiceClient
	ctx        context.Context
	cancel     context.CancelFunc
	store      *persistence.Profiles
	controller *controller.Controller
	runner     *runner
	network    *network
}

func newFixture(t *testing.T, prepare func(context.Context, profile.Profile) (*ports.Plan, error)) *fixture {
	t.Helper()
	if prepare == nil {
		prepare = func(context.Context, profile.Profile) (*ports.Plan, error) {
			return &ports.Plan{
				Relay: ports.Command{Path: "relay"}, Engine: ports.Command{Path: "engine"}, MixedAddress: "127.0.0.1:19876",
				Ready: func(context.Context) error { return nil }, Probe: func(context.Context) error { return nil }, Close: func() error { return nil },
			}, nil
		}
	}
	r, n := &runner{}, &network{}
	c := controller.New(r, preparer{prepare}, n, controller.Options{StartupTimeout: 2 * time.Second, CleanupTimeout: time.Second, NetworkForMode: func(proxymode.Mode) (ports.Network, error) { return n, nil }})
	store := &persistence.Profiles{Dir: filepath.Join(t.TempDir(), "profiles")}
	lifetime, stop := context.WithCancel(context.Background())
	token := []byte(strings.Repeat("s", 32))
	root := t.TempDir()
	b, err := backend.Open(lifetime, backend.Options{Paths: platform.Paths{Data: filepath.Join(root, "data"), Runtime: filepath.Join(root, "runtime")}})
	if err != nil {
		t.Fatal(err)
	}
	b.Controller = c
	b.Profiles = store
	t.Cleanup(func() { _ = b.Close() })
	server, err := rpc.NewServer(b, rpc.Options{Context: lifetime, Token: token, BuildID: "test-build", Epoch: "test-epoch"})
	if err != nil {
		t.Fatal(err)
	}
	// Проверяем, что изменение исходного буфера не меняет credential сервера.
	for i := range token {
		token[i] = 0
	}
	listener := bufconn.Listen(2 * rpc.MaxMessageBytes)
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	conn, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	requestCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx := metadata.AppendToOutgoingContext(requestCtx, rpc.TokenMetadataKey, hex.EncodeToString([]byte(strings.Repeat("s", 32))))
	f := &fixture{pb.NewControllerServiceClient(conn), ctx, stop, store, c, r, n}
	t.Cleanup(func() {
		cancel()
		stop()
		server.Stop()
		conn.Close()
		listener.Close()
		<-serveDone
		if err := c.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	return f
}
func link(json string) string {
	return "freeturn://" + base64.RawURLEncoding.EncodeToString([]byte(json))
}
func request() *pb.ImportProfileRequest {
	return pb.ImportProfileRequest_builder{
		FreeturnLink: link(`{"v":1,"provider":"direct","peer":"server.example:443","name":"Личный","cid":"client","wg":"private-config","key":"secret-key","dns":"doh","dnss":"1.1.1.1","kcp":{},"n":3,"spc":2,"bond":true}`),
	}.Build()
}
func importProfile(t *testing.T, f *fixture) string {
	t.Helper()
	p, err := f.client.ImportProfile(f.ctx, request())
	if err != nil {
		t.Fatal(err)
	}
	return p.GetId()
}
func requireCode(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if status.Code(err) != want {
		t.Fatalf("code=%v want=%v err=%v", status.Code(err), want, err)
	}
}
func waitState(t *testing.T, f *fixture, want pb.State) *pb.Snapshot {
	t.Helper()
	stream, err := f.client.WatchSnapshots(f.ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	for {
		s, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if s.GetState() == want {
			return s
		}
	}
}

func TestSessionTokenHeader(t *testing.T) {
	f := newFixture(t, nil)
	token := hex.EncodeToString([]byte(strings.Repeat("s", 32)))
	for _, tc := range []struct {
		header string
		code   codes.Code
	}{
		{"x-wireglow-session-token", codes.OK},
		{"x-unrelated-session-token", codes.Unauthenticated},
	} {
		t.Run(tc.header, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(tc.header, token))
			_, err := f.client.Hello(ctx, &emptypb.Empty{})
			requireCode(t, err, tc.code)
		})
	}
}

func TestHelloAuthentication(t *testing.T) {
	f := newFixture(t, nil)
	for _, values := range [][]string{nil, {"bad"}, {strings.Repeat("0", 64)}, {strings.Repeat("z", 64)}, {hex.EncodeToString([]byte(strings.Repeat("s", 32))), "duplicate"}} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if values != nil {
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD{rpc.TokenMetadataKey: values})
		}
		_, err := f.client.Hello(ctx, &emptypb.Empty{})
		requireCode(t, err, codes.Unauthenticated)
		_, err = f.client.ImportProfile(ctx, request())
		requireCode(t, err, codes.Unauthenticated)
		stream, err := f.client.WatchSnapshots(ctx, &emptypb.Empty{})
		if err == nil {
			_, err = stream.Recv()
		}
		requireCode(t, err, codes.Unauthenticated)
		cancel()
	}
	list, err := f.store.List(f.ctx)
	if err != nil || len(list) != 0 {
		t.Fatalf("unauthorized write: %v %v", list, err)
	}
	hello, err := f.client.Hello(f.ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if hello.GetBuildId() != "test-build" || hello.GetEpoch() != "test-epoch" {
		t.Fatal(hello)
	}
}
func TestProfilesPreviewImportReadDelete(t *testing.T) {
	f := newFixture(t, nil)
	preview, err := f.client.PreviewImport(f.ctx, request())
	if err != nil {
		t.Fatal(err)
	}
	if preview.GetId() != "" || preview.GetName() != "Личный" {
		t.Fatal(preview)
	}
	list, err := f.client.ListProfiles(f.ctx, &emptypb.Empty{})
	if err != nil || len(list.GetProfiles()) != 0 {
		t.Fatalf("preview saved data: %v %v", list, err)
	}
	id := importProfile(t, f)
	list, err = f.client.ListProfiles(f.ctx, &emptypb.Empty{})
	if err != nil || len(list.GetProfiles()) != 1 {
		t.Fatalf("list: %v %v", list, err)
	}
	public := list.GetProfiles()[0]
	if public.GetPeer() != "server.example:443" || public.GetStreams() != 3 || public.GetStreamsPerCredential() != 2 {
		t.Fatal("public properties missing", public)
	}
	encoded, _ := protojson.Marshal(list)
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "private-config") {
		t.Fatal("summary leaked secrets")
	}
	p, err := f.client.GetProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
	if err != nil {
		t.Fatal(err)
	}
	c := p.GetConnection()
	if c.GetName() != "Личный" || c.GetClientId() != "client" || c.GetPeer() != "server.example:443" || c.GetStreams() != 3 || c.GetStreamsPerCredential() != 2 || !c.GetBond() || c.GetWg().GetConfig() != "private-config" || c.GetObfuscation().GetKey() != "secret-key" {
		t.Fatal(p)
	}
	if c.HasVk() || !c.HasKcp() || c.GetKcp().GetMtu() != 0 || c.GetDns().GetMode() != "doh" || len(c.GetDns().GetServers()) != 1 {
		t.Fatal("optional sections/defaults lost", p)
	}
	_, err = f.client.DeleteProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.client.GetProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
	requireCode(t, err, codes.NotFound)
	_, err = f.client.DeleteProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
	requireCode(t, err, codes.NotFound)
}
func TestValidationAndSanitizedStorageErrors(t *testing.T) {
	f := newFixture(t, nil)
	_, err := f.client.ImportProfile(f.ctx, pb.ImportProfileRequest_builder{FreeturnLink: "secret-invalid-link"}.Build())
	requireCode(t, err, codes.InvalidArgument)
	if strings.Contains(err.Error(), "secret-invalid-link") {
		t.Fatal(err)
	}
	for _, id := range []string{"", "../secret", strings.Repeat("A", 32)} {
		_, err := f.client.GetProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
		requireCode(t, err, codes.InvalidArgument)
		_, err = f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
		requireCode(t, err, codes.InvalidArgument)
		_, err = f.client.DeleteProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
		requireCode(t, err, codes.InvalidArgument)
	}
	_, err = f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: strings.Repeat("0", 32), Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	requireCode(t, err, codes.NotFound)
	id := importProfile(t, f)
	if err := os.WriteFile(filepath.Join(f.store.Dir, id+".json"), []byte("broken-secret-json"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = f.client.GetProfile(f.ctx, pb.ProfileRequest_builder{Id: id}.Build())
	requireCode(t, err, codes.Internal)
	if strings.Contains(err.Error(), f.store.Dir) || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
	_, err = f.client.ImportProfile(f.ctx, pb.ImportProfileRequest_builder{FreeturnLink: strings.Repeat("a", rpc.MaxMessageBytes+1)}.Build())
	requireCode(t, err, codes.ResourceExhausted)
}
func TestConnectStreamsAndDisconnect(t *testing.T) {
	f := newFixture(t, nil)
	initial, err := f.client.GetSnapshot(f.ctx, &emptypb.Empty{})
	if err != nil || initial.GetState() != pb.State_STATE_IDLE {
		t.Fatalf("initial=%v %v", initial, err)
	}
	id := importProfile(t, f)
	req := pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build()
	_, err = f.client.Connect(f.ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	connected := waitState(t, f, pb.State_STATE_CONNECTED)
	if connected.GetProfileId() != id || connected.GetGeneration() != 1 || connected.GetMixedAddress() != "127.0.0.1:19876" {
		t.Fatal(connected)
	}
	_, err = f.client.Connect(f.ctx, req)
	requireCode(t, err, codes.FailedPrecondition)
	idle, err := f.client.Disconnect(f.ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if idle.GetState() != pb.State_STATE_IDLE || idle.GetMixedAddress() != "" || idle.GetProfileId() != "" {
		t.Fatal(idle)
	}
	if f.runner.starts.Load() != 2 || f.network.applies.Load() != 1 || f.network.recoveries.Load() < 2 {
		t.Fatal("lifecycle not executed")
	}
	_, err = f.client.Disconnect(f.ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
}
func TestConnectACKSurvivesRequestCancellationAndDisconnectCancelsStartup(t *testing.T) {
	entered := make(chan struct{})
	f := newFixture(t, func(ctx context.Context, _ profile.Profile) (*ports.Plan, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	id := importProfile(t, f)
	ctx, cancel := context.WithCancel(f.ctx)
	_, err := f.client.Connect(ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case <-entered:
	case <-f.ctx.Done():
		t.Fatal("startup did not begin")
	}
	state, err := f.client.GetSnapshot(f.ctx, &emptypb.Empty{})
	if err != nil || state.GetState() != pb.State_STATE_VALIDATING {
		t.Fatalf("RPC cancellation stopped startup: %v %v", state, err)
	}
	_, err = f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	requireCode(t, err, codes.FailedPrecondition)
	idle, err := f.client.Disconnect(f.ctx, &emptypb.Empty{})
	if err != nil || idle.GetState() != pb.State_STATE_IDLE {
		t.Fatalf("cancel startup: %v %v", idle, err)
	}
	if f.runner.starts.Load() != 0 {
		t.Fatal("process started after cancellation")
	}
}
func TestStartupFailureDoesNotLeakEngineError(t *testing.T) {
	f := newFixture(t, func(context.Context, profile.Profile) (*ports.Plan, error) {
		return nil, errors.New("private-key=very-secret")
	})
	id := importProfile(t, f)
	_, err := f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	if err != nil {
		t.Fatal(err)
	}
	failed := waitState(t, f, pb.State_STATE_FAILED)
	if failed.GetError() != "connection preparation failed" || strings.Contains(failed.GetError(), "secret") {
		t.Fatal(failed)
	}
}
func TestWatchCancellationDoesNotDisconnectAndLifetimeClosesStream(t *testing.T) {
	f := newFixture(t, nil)
	id := importProfile(t, f)
	_, err := f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	if err != nil {
		t.Fatal(err)
	}
	waitState(t, f, pb.State_STATE_CONNECTED)
	ctx, cancel := context.WithCancel(f.ctx)
	stream, err := f.client.WatchSnapshots(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	first, err := stream.Recv()
	if err != nil || first.GetState() != pb.State_STATE_CONNECTED {
		t.Fatalf("initial snapshot: %v %v", first, err)
	}
	cancel()
	_, err = stream.Recv()
	requireCode(t, err, codes.Canceled)
	current, err := f.client.GetSnapshot(f.ctx, &emptypb.Empty{})
	if err != nil || current.GetState() != pb.State_STATE_CONNECTED {
		t.Fatal("subscription became owner")
	}
	stream, err = f.client.WatchSnapshots(f.ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = stream.Recv(); err != nil {
		t.Fatal(err)
	}
	f.cancel()
	_, err = stream.Recv()
	requireCode(t, err, codes.Unavailable)
	_, err = f.client.Hello(f.ctx, &emptypb.Empty{})
	requireCode(t, err, codes.Unavailable)
}
func TestImmediateDisconnectCannotStartLater(t *testing.T) {
	f := newFixture(t, nil)
	id := importProfile(t, f)
	for range 20 {
		_, err := f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
		if err != nil {
			t.Fatal(err)
		}
		idle, err := f.client.Disconnect(f.ctx, &emptypb.Empty{})
		if err != nil || idle.GetState() != pb.State_STATE_IDLE {
			t.Fatalf("disconnect: %v %v", idle, err)
		}
		state, err := f.client.GetSnapshot(f.ctx, &emptypb.Empty{})
		if err != nil || state.GetState() != pb.State_STATE_IDLE {
			t.Fatal("startup outlived disconnect", state, err)
		}
	}
}

func TestVKAndKCPSettingsOverWire(t *testing.T) {
	f := newFixture(t, nil)
	input := pb.ImportProfileRequest_builder{FreeturnLink: link(`{"v":1,"provider":"vk","peer":"host:42","transport":"tcp","mode":"udp","obf":"rtpopus3","timing":17,"vk":"https://vk.example/one,https://vk.example/two","mcap":true,"kcp":{"nodelay":1,"interval":20,"resend":2,"nc":1,"sndwnd":128,"rcvwnd":256,"mtu":1200,"acknodelay":true}}`)}.Build()
	imported, err := f.client.ImportProfile(f.ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.client.GetProfile(f.ctx, pb.ProfileRequest_builder{Id: imported.GetId()}.Build())
	if err != nil {
		t.Fatal(err)
	}
	c := p.GetConnection()
	k := c.GetKcp()
	if !c.HasVk() || len(c.GetVk().GetCallLinks()) != 2 || !c.GetVk().GetManualCaptcha() || c.HasWg() || c.GetTurnTransport() != "tcp" || c.GetRelayMode() != "udp" || c.GetObfuscation().GetMode() != "rtpopus3" || c.GetObfuscation().GetTimingMs() != 17 {
		t.Fatal(p)
	}
	if k.GetNoDelay() != 1 || k.GetInterval() != 20 || k.GetResend() != 2 || k.GetNc() != 1 || k.GetSndWnd() != 128 || k.GetRcvWnd() != 256 || k.GetMtu() != 1200 || !k.GetAckNoDelay() {
		t.Fatal(k)
	}
}
func TestLifetimeCancellationStopsStartup(t *testing.T) {
	entered := make(chan struct{})
	f := newFixture(t, func(ctx context.Context, _ profile.Profile) (*ports.Plan, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	id := importProfile(t, f)
	_, err := f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-f.ctx.Done():
		t.Fatal("startup did not begin")
	}
	events, unsubscribe := f.controller.Subscribe()
	defer unsubscribe()
	f.cancel()
	for {
		select {
		case s := <-events:
			if s.State == controller.Failed {
				if f.network.applies.Load() != 0 || f.runner.starts.Load() != 0 || f.network.recoveries.Load() < 2 {
					t.Fatal("cancelled startup did not clean up")
				}
				return
			}
		case <-f.ctx.Done():
			t.Fatal("startup ignored lifetime")
		}
	}
}

func TestConcurrentConnectAdmitsOneStartup(t *testing.T) {
	f := newFixture(t, func(ctx context.Context, _ profile.Profile) (*ports.Plan, error) { <-ctx.Done(); return nil, ctx.Err() })
	id := importProfile(t, f)
	req := pb.ConnectRequest_builder{Id: id, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build()
	results := make(chan error, 16)
	for range 16 {
		go func() { _, err := f.client.Connect(f.ctx, req); results <- err }()
	}
	accepted := 0
	for range 16 {
		err := <-results
		if err == nil {
			accepted++
		} else {
			requireCode(t, err, codes.FailedPrecondition)
		}
	}
	if accepted != 1 {
		t.Fatalf("accepted %d concurrent startups", accepted)
	}
	idle, err := f.client.Disconnect(f.ctx, &emptypb.Empty{})
	if err != nil || idle.GetState() != pb.State_STATE_IDLE {
		t.Fatalf("disconnect: %v %v", idle, err)
	}
}

func TestLatencyRPCAndSnapshots(t *testing.T) {
	f := newFixture(t, func(context.Context, profile.Profile) (*ports.Plan, error) {
		return &ports.Plan{
			Relay: ports.Command{Path: "relay"}, Engine: ports.Command{Path: "engine"},
			Ready: func(context.Context) error { return nil }, Probe: func(context.Context) error { return nil }, Close: func() error { return nil },
			MeasureLatency: func(context.Context) (time.Duration, error) { return 87 * time.Millisecond, nil },
		}, nil
	})
	if _, err := f.client.CheckLatency(f.ctx, &emptypb.Empty{}); status.Code(err) != codes.FailedPrecondition {
		t.Fatal(err)
	}
	if _, err := f.client.CheckLatency(context.Background(), &emptypb.Empty{}); status.Code(err) != codes.Unauthenticated {
		t.Fatal(err)
	}
	if err := f.controller.Connect(f.ctx, profile.Profile{ID: "test"}); err != nil {
		t.Fatal(err)
	}
	updates, unsubscribe := f.controller.Subscribe()
	defer unsubscribe()
	for s := range updates {
		if s.Latency.SuccessAt > 0 && !s.Latency.Checking {
			break
		}
	}
	stream, err := f.client.WatchSnapshots(f.ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	initial, err := stream.Recv()
	if err != nil || initial.GetLatency().GetDelayMs() != 87 {
		t.Fatal(initial, err)
	}
	result, err := f.client.CheckLatency(f.ctx, &emptypb.Empty{})
	if err != nil || result.GetLatency().GetDelayMs() != 87 || result.GetLatency().GetSuccessAt() == 0 {
		t.Fatal(result, err)
	}
}

func TestProxySettingsAndModeValidation(t *testing.T) {
	f := newFixture(t, nil)
	settings, err := f.client.GetSettings(f.ctx, &emptypb.Empty{})
	if err != nil || settings.GetPreferredMode() != pb.ProxyMode_PROXY_MODE_LOCAL {
		t.Fatalf("default = %v, %v", settings, err)
	}
	_, err = f.client.SetSettings(f.ctx, pb.ClientSettings_builder{PreferredMode: pb.ProxyMode_PROXY_MODE_SYSTEM}.Build())
	if err != nil {
		t.Fatal(err)
	}
	settings, err = f.client.GetSettings(f.ctx, &emptypb.Empty{})
	if err != nil || settings.GetPreferredMode() != pb.ProxyMode_PROXY_MODE_SYSTEM {
		t.Fatalf("settings = %v, %v", settings, err)
	}
	for _, mode := range []pb.ProxyMode{pb.ProxyMode_PROXY_MODE_UNSPECIFIED, pb.ProxyMode(99)} {
		_, err = f.client.SetSettings(f.ctx, pb.ClientSettings_builder{PreferredMode: mode}.Build())
		requireCode(t, err, codes.InvalidArgument)
		_, err = f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: importProfile(t, f), Mode: mode}.Build())
		requireCode(t, err, codes.InvalidArgument)
	}
	_, err = f.client.Connect(f.ctx, pb.ConnectRequest_builder{Id: importProfile(t, f), Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	if err != nil {
		t.Fatal(err)
	}
	connected := waitState(t, f, pb.State_STATE_CONNECTED)
	if connected.GetActiveMode() != pb.ProxyMode_PROXY_MODE_LOCAL {
		t.Fatal(connected)
	}
	// Saving the next preference does not mutate the current session.
	_, err = f.client.SetSettings(f.ctx, pb.ClientSettings_builder{PreferredMode: pb.ProxyMode_PROXY_MODE_SYSTEM}.Build())
	if err != nil {
		t.Fatal(err)
	}
	current, err := f.client.GetSnapshot(f.ctx, &emptypb.Empty{})
	if err != nil || current.GetActiveMode() != pb.ProxyMode_PROXY_MODE_LOCAL {
		t.Fatalf("active = %v, %v", current, err)
	}
}

func TestSwitchProxyModePreservesSession(t *testing.T) {
	f := newFixture(t, nil)
	_, err := f.client.SwitchProxyMode(f.ctx, pb.ProxyModeRequest_builder{Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	requireCode(t, err, codes.FailedPrecondition)
	for _, mode := range []pb.ProxyMode{pb.ProxyMode_PROXY_MODE_UNSPECIFIED, pb.ProxyMode(99)} {
		_, err := f.client.SwitchProxyMode(f.ctx, pb.ProxyModeRequest_builder{Mode: mode}.Build())
		requireCode(t, err, codes.InvalidArgument)
	}
	if err := f.controller.ConnectWithMode(f.ctx, profile.Profile{ID: "one"}, proxymode.System); err != nil {
		t.Fatal(err)
	}
	before := f.controller.Snapshot()
	actual, err := f.client.SwitchProxyMode(f.ctx, pb.ProxyModeRequest_builder{Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build())
	if err != nil {
		t.Fatal(err)
	}
	if actual.GetActiveMode() != pb.ProxyMode_PROXY_MODE_LOCAL || actual.GetState() != pb.State_STATE_CONNECTED || actual.GetGeneration() != before.Generation || actual.GetMixedAddress() != before.MixedAddress || f.runner.starts.Load() != 2 {
		t.Fatal(actual)
	}
}

func TestClientAddressesOverWire(t *testing.T) {
	f := newFixture(t, nil)
	request := pb.ClientSettings_builder{PreferredMode: pb.ProxyMode_PROXY_MODE_LOCAL, ProxyAddress: "127.0.0.1:28001", RelayAddress: "127.0.0.1:28002", ApiPort: 28003}.Build()
	result, err := f.client.SetSettings(f.ctx, request)
	if err != nil || result.GetProxyAddress() != request.GetProxyAddress() || result.GetRelayAddress() != request.GetRelayAddress() || result.GetApiPort() != 28003 {
		t.Fatalf("set = %v, %v", result, err)
	}
	result, err = f.client.GetSettings(f.ctx, &emptypb.Empty{})
	if err != nil || result.GetProxyAddress() != request.GetProxyAddress() || result.GetRelayAddress() != request.GetRelayAddress() || result.GetApiPort() != 28003 {
		t.Fatalf("get = %v, %v", result, err)
	}
	for _, port := range []uint32{28001, 28002, 65536} {
		request.SetApiPort(port)
		if _, err := f.client.SetSettings(f.ctx, request); err == nil {
			t.Fatalf("accepted API port %d", port)
		}
	}
	request.SetApiPort(28003)

	request.SetRelayAddress("127.0.0.1:28001")
	if _, err := f.client.SetSettings(f.ctx, request); err == nil {
		t.Fatal("colliding endpoints accepted")
	}
	result, err = f.client.GetSettings(f.ctx, &emptypb.Empty{})
	if err != nil || result.GetRelayAddress() != "127.0.0.1:28002" {
		t.Fatalf("invalid write changed settings = %v, %v", result, err)
	}
}
