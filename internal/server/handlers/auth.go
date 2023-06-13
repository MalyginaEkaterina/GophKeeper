package handlers

import (
	"context"
	"errors"
	"github.com/MalyginaEkaterina/GophKeeper/internal/common"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/service"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

const (
	userIDKey = authContextKey("userID")
)

type authContextKey string

type UserServer struct {
	pb.UnimplementedUserServer
	authService service.AuthService
}

func NewUserServer(authService service.AuthService) *UserServer {
	return &UserServer{authService: authService}
}

// Register calls RegisterUser of AuthService.
// Returns InvalidArgument code, if login or password is empty.
// Returns AlreadyExists code, if login has been registered already.
func (s *UserServer) Register(ctx context.Context, in *pb.RegisterReq) (*pb.RegisterResp, error) {
	if err := validateAuthData(in.Login, in.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err := s.authService.RegisterUser(ctx, in.Login, in.Password)
	if errors.Is(err, storage.ErrAlreadyExists) {
		return nil, status.Errorf(codes.AlreadyExists, "User with login %s already exists", in.Login)
	} else if err != nil {
		log.Println("Register user error: ", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	return &pb.RegisterResp{}, nil
}

// Auth calls AuthUser of AuthService.
// Returns InvalidArgument code, if login or password is empty.
// Returns Unauthenticated code, if password is incorrect.
func (s *UserServer) Auth(ctx context.Context, in *pb.AuthReq) (*pb.AuthResp, error) {
	if err := validateAuthData(in.Login, in.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	token, err := s.authService.AuthUser(ctx, in.Login, in.Password)
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, service.ErrIncorrectPassword) {
		return nil, status.Error(codes.Unauthenticated, "Incorrect login/password")
	} else if err != nil {
		log.Println("Authentication user error: ", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	response := pb.AuthResp{Token: string(token)}
	return &response, nil
}

func GetUserIDFromContext(ctx context.Context) server.UserID {
	return ctx.Value(userIDKey).(server.UserID)
}

// AuthInterceptor is interceptor that gets token from token required requests, checks it and gets userId from it.
// Then puts userId into context. If checks failed returns Unauthenticated code in response.
func (s *UserServer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	switch info.Server.(type) {
	case *UserServer:
		return handler(ctx, req)
	case *KeeperServer:
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Internal error")
		}
		v := md.Get(common.AuthHeader)
		if len(v) > 0 {
			token := v[0]
			userID, err := s.authService.CheckToken(token)
			if errors.Is(err, service.ErrUnauthorized) {
				return nil, status.Errorf(codes.Unauthenticated, err.Error())
			} else if err != nil {
				log.Println("Check token error: ", err)
				return nil, status.Errorf(codes.Internal, "Internal error")
			}
			ctx = context.WithValue(ctx, userIDKey, userID)
			return handler(ctx, req)
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "missing token")
		}
	default:
		return nil, status.Error(codes.Internal, "Internal server error")
	}
}

func validateAuthData(login, password string) error {
	if login == "" {
		return errors.New("login is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	return nil
}
