package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/MalyginaEkaterina/GophKeeper/internal/common"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

type GrpcClient struct {
	conn         *grpc.ClientConn
	userClient   pb.UserClient
	keeperClient pb.KeeperClient
	getToken     func() string
}

func newGrpcClient(address string, getToken func() string) *GrpcClient {
	grpcClient := &GrpcClient{getToken: getToken}
	tlsCredential := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(tlsCredential), grpc.WithUnaryInterceptor(grpcClient.clientInterceptor))
	if err != nil {
		log.Fatal(err)
	}
	grpcClient.conn = conn
	grpcClient.userClient = pb.NewUserClient(conn)
	grpcClient.keeperClient = pb.NewKeeperClient(conn)
	return grpcClient
}

func (g *GrpcClient) Close() {
	g.conn.Close()
}

func (g *GrpcClient) clientInterceptor(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	if method == "/common.User/Auth" || method == "/common.User/Register" {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	token := g.getToken()

	if token == "" {
		return status.Errorf(codes.Unauthenticated, "missing token")
	}
	md := metadata.New(map[string]string{common.AuthHeader: token})
	ctx = metadata.NewOutgoingContext(ctx, md)
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (g *GrpcClient) putByKey(putReq *pb.PutReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := g.keeperClient.Put(ctx, putReq)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.Unavailable || e.Code() == codes.DeadlineExceeded {
				return errServerUnavailable
			} else if e.Code() == codes.Unauthenticated {
				return errNeedLogin
			} else if e.Code() == codes.AlreadyExists {
				return errDataConflict
			} else {
				return fmt.Errorf("put data error: %s, %s", e.Code(), e.Message())
			}
		}
		return fmt.Errorf("put data error: %w", err)
	}
	return nil
}

func (g *GrpcClient) getByKey(key string) (*pb.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := g.keeperClient.Get(ctx, &pb.GetReq{Key: key})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.Unavailable || e.Code() == codes.DeadlineExceeded {
				return nil, errServerUnavailable
			} else if e.Code() == codes.Unauthenticated {
				return nil, errNeedLogin
			} else if e.Code() == codes.NotFound {
				return nil, errDataNotFound
			} else {
				return nil, fmt.Errorf("get data by key error: %s, %s", e.Code(), e.Message())
			}
		}
		return nil, fmt.Errorf("get data by key error: %w", err)
	}
	return resp.Value, nil
}

func (g *GrpcClient) getAll(version int32) (map[string]*pb.Value, int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := g.keeperClient.GetAll(ctx, &pb.GetAllReq{Version: version})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.Unavailable || e.Code() == codes.DeadlineExceeded {
				return nil, 0, errServerUnavailable
			} else if e.Code() == codes.NotFound {
				return nil, 0, errDataNotFound
			}
			return nil, 0, fmt.Errorf("getting all error: %s, %s", e.Code(), e.Message())
		}
		return nil, 0, fmt.Errorf("getting all error: %w", err)
	}
	return resp.Result, resp.Version, nil
}

func (g *GrpcClient) getKeyList() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := g.keeperClient.List(ctx, &pb.ListReq{})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.Unavailable || e.Code() == codes.DeadlineExceeded {
				return nil, errServerUnavailable
			} else if e.Code() == codes.Unauthenticated {
				return nil, errNeedLogin
			} else {
				return nil, fmt.Errorf("get keys list error: %s, %s", e.Code(), e.Message())
			}
		}
		return nil, fmt.Errorf("getting key list error: %w", err)
	}
	return resp.Keys, nil
}

func (g *GrpcClient) register(username, password string) error {
	_, err := g.userClient.Register(context.Background(), &pb.RegisterReq{Login: username, Password: password})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				return errAlreadyExists
			} else {
				return fmt.Errorf("registration error: %s, %s", e.Code(), e.Message())
			}
		}
		return fmt.Errorf("registration error: %w", err)
	}
	return nil
}

func (g *GrpcClient) auth(username, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	authResp, err := g.userClient.Auth(ctx, &pb.AuthReq{Login: username, Password: password})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.InvalidArgument || e.Code() == codes.Unauthenticated {
				return "", errIncorrectAuth
			} else if e.Code() == codes.Unavailable || e.Code() == codes.DeadlineExceeded {
				return "", errServerUnavailable
			}
			return "", fmt.Errorf("auth error: %s, %s", e.Code(), e.Message())
		}
		return "", fmt.Errorf("auth error: %w", err)
	}
	return authResp.Token, nil
}
