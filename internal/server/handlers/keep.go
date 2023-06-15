package handlers

import (
	"context"
	"errors"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type KeeperServer struct {
	pb.UnimplementedKeeperServer
	dataStore storage.DataStorage
}

func NewKeeperServer(dataStore storage.DataStorage) *KeeperServer {
	return &KeeperServer{dataStore: dataStore}
}

// Get calls GetByKey of DataStorage.
// Returns NotFound code, if data was not found.
func (s *KeeperServer) Get(ctx context.Context, in *pb.GetReq) (*pb.GetResp, error) {
	userID := GetUserIDFromContext(ctx)
	result, err := s.dataStore.GetByKey(ctx, userID, in.Key)
	if errors.Is(err, storage.ErrNotFound) {
		LogWithUserID(ctx, "Data with key = %s not found", in.Key)
		return nil, status.Errorf(codes.NotFound, "Data with key = %s not found", in.Key)
	} else if err != nil {
		LogErrorWithUserID(ctx, "Get data by key error", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.GetResp{Value: result}, nil
}

// Put calls Put of DataStorage.
// Returns AlreadyExists if there is the conflict in versions.
func (s *KeeperServer) Put(ctx context.Context, in *pb.PutReq) (*pb.PutResp, error) {
	userID := GetUserIDFromContext(ctx)
	err := s.dataStore.Put(ctx, userID, in.Key, in.Value)
	if errors.Is(err, storage.ErrConflict) {
		LogWithUserID(ctx, "Data conflict with key = %s", in.Key)
		return nil, status.Errorf(codes.AlreadyExists, "Data conflict with key = %s", in.Key)
	} else if err != nil {
		LogErrorWithUserID(ctx, "Put data error", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.PutResp{}, nil
}

// List calls GetAllKeysByUser of DataStorage.
func (s *KeeperServer) List(ctx context.Context, _ *pb.ListReq) (*pb.ListResp, error) {
	userID := GetUserIDFromContext(ctx)
	keys, err := s.dataStore.GetAllKeysByUser(ctx, userID)
	if err != nil {
		LogErrorWithUserID(ctx, "Get keys list error", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.ListResp{Keys: keys}, nil
}

// GetAll calls GetAllByUser of DataStorage.
// Returns NotFound code, if there is no data with new version.
func (s *KeeperServer) GetAll(ctx context.Context, in *pb.GetAllReq) (*pb.GetAllResp, error) {
	userID := GetUserIDFromContext(ctx)
	result, currVersion, err := s.dataStore.GetAllByUser(ctx, userID, in.Version)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "New data not found")
	} else if err != nil {
		LogErrorWithUserID(ctx, "Get all data error", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.GetAllResp{Result: result, Version: currVersion}, nil
}
