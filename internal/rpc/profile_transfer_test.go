package rpc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	pb "wireglow/internal/rpc/gen"
)

func TestProfileArchiveRPC(t *testing.T) {
	f := newFixture(t, nil)
	id := importProfile(t, f)
	path := filepath.Join(t.TempDir(), "profiles.json")
	if _, err := f.client.ExportProfiles(f.ctx, pb.ExportProfilesRequest_builder{Ids: []string{id}, Path: path}.Build()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "private-config") {
		t.Fatal("export omitted secrets", err)
	}
	preview, err := f.client.PreviewProfileFile(f.ctx, pb.ProfileFileRequest_builder{Path: path}.Build())
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.GetEntries()) != 1 || !preview.GetEntries()[0].HasExisting() {
		t.Fatal("missing conflict")
	}
	json, err := protojson.Marshal(preview)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"private-config", "secret-key"} {
		if strings.Contains(string(json), secret) {
			t.Fatal("preview leaked a secret")
		}
	}
	response, err := f.client.ImportProfileFile(f.ctx, pb.ImportProfileFileRequest_builder{Token: preview.GetToken(), Choices: []*pb.ProfileImportChoice{pb.ProfileImportChoice_builder{Index: 0, Action: pb.ProfileImportAction_PROFILE_IMPORT_ACTION_COPY}.Build()}}.Build())
	if err != nil || response.GetResults()[0].GetStatus() != "imported" {
		t.Fatal(response, err)
	}
	list, _ := f.store.List(f.ctx)
	if len(list) != 2 {
		t.Fatal("copy did not persist")
	}
	if f.runner.starts.Load() != 0 || f.network.applies.Load() != 0 {
		t.Fatal("transfer started a connection")
	}
	_, err = f.client.ImportProfileFile(f.ctx, pb.ImportProfileFileRequest_builder{Token: preview.GetToken()}.Build())
	requireCode(t, err, codes.FailedPrecondition)
	if err = os.WriteFile(path, []byte(`{"secret":"should never appear in RPC errors"}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = f.client.PreviewProfileFile(f.ctx, pb.ProfileFileRequest_builder{Path: path}.Build())
	requireCode(t, err, codes.InvalidArgument)
	if strings.Contains(err.Error(), "should never appear") {
		t.Fatal("raw JSON error leaked")
	}
}
