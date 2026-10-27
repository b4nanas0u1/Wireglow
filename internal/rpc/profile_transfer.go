package rpc

import (
	"context"
	"errors"

	"wireglow/internal/application/profiletransfer"
	"wireglow/internal/domain/profile"
	pb "wireglow/internal/rpc/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func transferError(err error) error {
	switch {
	case errors.Is(err, profiletransfer.ErrInvalid):
		return status.Error(codes.InvalidArgument, "invalid profile archive")
	case errors.Is(err, profiletransfer.ErrExpired):
		return status.Error(codes.FailedPrecondition, "profile preview expired")
	default:
		return rpcError(err)
	}
}
func (s *service) ExportProfiles(ctx context.Context, req *pb.ExportProfilesRequest) (*emptypb.Empty, error) {
	if err := s.transfer.Export(ctx, req.GetIds(), req.GetPath()); err != nil {
		return nil, transferError(err)
	}
	return &emptypb.Empty{}, nil
}
func (s *service) PreviewProfileFile(ctx context.Context, req *pb.ProfileFileRequest) (*pb.ProfileFilePreview, error) {
	preview, err := s.transfer.Preview(ctx, req.GetPath())
	if err != nil {
		return nil, transferError(err)
	}
	entries := make([]*pb.ProfileFileEntry, 0, len(preview.Entries))
	for _, entry := range preview.Entries {
		item := pb.ProfileFileEntry_builder{Profile: transferSummary(entry.Profile)}
		if entry.Existing != nil {
			item.Existing = transferSummary(*entry.Existing)
		}
		entries = append(entries, item.Build())
	}
	return pb.ProfileFilePreview_builder{Token: preview.Token, Entries: entries}.Build(), nil
}
func (s *service) ImportProfileFile(ctx context.Context, req *pb.ImportProfileFileRequest) (*pb.ImportProfileFileResponse, error) {
	// Serialize with connect/disconnect; saved profile changes never restart active engines.
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running != nil || s.stopping {
		return nil, status.Error(codes.FailedPrecondition, "session operation running")
	}
	actions := map[pb.ProfileImportAction]string{
		pb.ProfileImportAction_PROFILE_IMPORT_ACTION_ADD: "add", pb.ProfileImportAction_PROFILE_IMPORT_ACTION_SKIP: "skip", pb.ProfileImportAction_PROFILE_IMPORT_ACTION_COPY: "copy", pb.ProfileImportAction_PROFILE_IMPORT_ACTION_REPLACE: "replace",
	}
	choices := make([]profiletransfer.Choice, 0, len(req.GetChoices()))
	for _, c := range req.GetChoices() {
		choices = append(choices, profiletransfer.Choice{Index: int(c.GetIndex()), Action: actions[c.GetAction()]})
	}
	results, err := s.transfer.Import(ctx, req.GetToken(), choices)
	if err != nil {
		return nil, transferError(err)
	}
	out := make([]*pb.ProfileImportResult, 0, len(results))
	for _, r := range results {
		out = append(out, pb.ProfileImportResult_builder{Index: uint32(r.Index), Status: r.Status}.Build())
	}
	return pb.ImportProfileFileResponse_builder{Results: out}.Build(), nil
}

// File previews need only identity and provider; this also bounds 1000-entry replies.
func transferSummary(p profile.Summary) *pb.ProfileSummary {
	return pb.ProfileSummary_builder{Id: p.ID, Name: p.Name, Provider: p.Provider}.Build()
}
