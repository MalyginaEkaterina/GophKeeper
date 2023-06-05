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

func (s *KeeperServer) Get(ctx context.Context, in *pb.GetReq) (*pb.GetResp, error) {
	userID := GetUserIDFromContext(ctx)
	result, err := s.dataStore.GetByKey(ctx, userID, in.Key)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "Data with key = %s not found", in.Key)
	} else if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.GetResp{Value: result}, nil
}

func (s *KeeperServer) Put(ctx context.Context, in *pb.PutReq) (*pb.PutResp, error) {
	userID := GetUserIDFromContext(ctx)
	err := s.dataStore.Put(ctx, userID, in.Key, in.Value)
	if errors.Is(err, storage.ErrConflict) {
		return nil, status.Errorf(codes.AlreadyExists, "Data conflict with key = %s", in.Key)
	} else if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.PutResp{}, nil
}

func (s *KeeperServer) List(ctx context.Context, _ *pb.ListReq) (*pb.ListResp, error) {
	userID := GetUserIDFromContext(ctx)
	keys, err := s.dataStore.GetAllKeysByUser(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.ListResp{Keys: keys}, nil
}

func (s *KeeperServer) GetAll(ctx context.Context, in *pb.GetAllReq) (*pb.GetAllResp, error) {
	userID := GetUserIDFromContext(ctx)
	result, currVersion, err := s.dataStore.GetAllByUser(ctx, userID, in.Version)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "New data not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.GetAllResp{Result: result, Version: currVersion}, nil
}
