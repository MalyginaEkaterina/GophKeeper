package handlers

import (
	"context"
	"github.com/MalyginaEkaterina/GophKeeper/internal/common"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/service"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

func newUserServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	store := storage.NewCachedStorage()
	authService := &service.AuthServiceImpl{Store: store, SecretKey: []byte("my key")}
	userServer := NewUserServer(authService)

	s := grpc.NewServer(grpc.UnaryInterceptor(userServer.AuthInterceptor))
	pb.RegisterUserServer(s, userServer)
	l := bufconn.Listen(128 * 1024)
	go func() {
		err := s.Serve(l)
		require.NoError(t, err)
	}()
	return s, l
}

func newKeeperServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	store := storage.NewCachedStorage()
	authService := &service.AuthServiceImpl{Store: store, SecretKey: []byte("my key")}
	userServer := NewUserServer(authService)
	keeperServer := NewKeeperServer(store)

	s := grpc.NewServer(grpc.UnaryInterceptor(userServer.AuthInterceptor))
	pb.RegisterUserServer(s, userServer)
	pb.RegisterKeeperServer(s, keeperServer)
	l := bufconn.Listen(128 * 1024)
	go func() {
		err := s.Serve(l)
		require.NoError(t, err)
	}()
	return s, l
}

func newClientConn(t *testing.T, l *bufconn.Listener) *grpc.ClientConn {
	dialer := func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	}
	conn, err := grpc.Dial("", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}

func testUserServer(t *testing.T, name string, f func(t *testing.T, client pb.UserClient)) {
	t.Run(name, func(t *testing.T) {
		s, l := newUserServer(t)
		defer s.Stop()
		conn := newClientConn(t, l)
		defer conn.Close()
		client := pb.NewUserClient(conn)
		f(t, client)
	})
}

func testKeeperServer(t *testing.T, name string, f func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient)) {
	t.Run(name, func(t *testing.T) {
		s, l := newKeeperServer(t)
		defer s.Stop()
		conn := newClientConn(t, l)
		defer conn.Close()
		userClient := pb.NewUserClient(conn)
		keeperClient := pb.NewKeeperClient(conn)
		f(t, userClient, keeperClient)
	})
}

func TestRegister(t *testing.T) {
	testUserServer(t, "positive register", func(t *testing.T, client pb.UserClient) {
		_, err := client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
	})
	testUserServer(t, "negative register test with empty password", func(t *testing.T, client pb.UserClient) {
		_, err := client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: ""})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.InvalidArgument, e.Code())
	})
	testUserServer(t, "negative register test with existed login", func(t *testing.T, client pb.UserClient) {
		_, err := client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		_, err = client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "321"})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.AlreadyExists, e.Code())
	})
}

func TestAuth(t *testing.T) {
	testUserServer(t, "positive auth", func(t *testing.T, client pb.UserClient) {
		_, err := client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := client.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		assert.Greater(t, len(respAuth.Token), 0)
	})
	testUserServer(t, "negative auth with wrong password", func(t *testing.T, client pb.UserClient) {
		_, err := client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		_, err = client.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "12"})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.Unauthenticated, e.Code())
	})
	testUserServer(t, "negative auth with empty password", func(t *testing.T, client pb.UserClient) {
		_, err := client.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		_, err = client.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: ""})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.InvalidArgument, e.Code())
	})
}

func TestPut(t *testing.T) {
	testKeeperServer(t, "positive put", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		v := &pb.Value{Data: []byte("key1 test"), Metadata: "metadata for key1", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key1", Value: v})
		require.NoError(t, err)
	})
	testKeeperServer(t, "negative put without token", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		v := &pb.Value{Data: []byte("key1 test"), Metadata: "metadata for key1", Version: 1}
		_, err := keeperClient.Put(context.Background(), &pb.PutReq{Key: "key1", Value: v})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.Unauthenticated, e.Code())
	})
	testKeeperServer(t, "negative put with conflict", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		v1 := &pb.Value{Data: []byte("key1 test"), Metadata: "metadata for key1", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key1", Value: v1})
		require.NoError(t, err)
		v2 := &pb.Value{Data: []byte("key1 test new"), Metadata: "metadata for key1", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key1", Value: v2})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.AlreadyExists, e.Code())
	})
}

func TestGet(t *testing.T) {
	testKeeperServer(t, "positive get", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		v := &pb.Value{Data: []byte("key1 test"), Metadata: "metadata for key1", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key1", Value: v})
		require.NoError(t, err)
		getResp, err := keeperClient.Get(ctx, &pb.GetReq{Key: "key1"})
		require.NoError(t, err)
		assert.Equal(t, v.Data, getResp.Value.Data)
		assert.Equal(t, v.Metadata, getResp.Value.Metadata)
		assert.Equal(t, v.Version, getResp.Value.Version)
	})
	testKeeperServer(t, "negative get with not existing key", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		_, err = keeperClient.Get(ctx, &pb.GetReq{Key: "key1"})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.NotFound, e.Code())
	})
	testKeeperServer(t, "negative get without token", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := keeperClient.Get(context.Background(), &pb.GetReq{Key: "key1"})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.Unauthenticated, e.Code())
	})
}

func TestList(t *testing.T) {
	testKeeperServer(t, "positive list", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		v1 := &pb.Value{Data: []byte("key1 test"), Metadata: "metadata for key1", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key1", Value: v1})
		require.NoError(t, err)
		v2 := &pb.Value{Data: []byte("key2 test"), Metadata: "metadata for key2", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key2", Value: v2})
		require.NoError(t, err)
		listResp, err := keeperClient.List(ctx, &pb.ListReq{})
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, listResp.Keys)
	})
	testKeeperServer(t, "negative list without token", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := keeperClient.List(context.Background(), &pb.ListReq{})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.Unauthenticated, e.Code())
	})
}

func TestGetAll(t *testing.T) {
	testKeeperServer(t, "positive get all", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		v1 := &pb.Value{Data: []byte("key1 test"), Metadata: "metadata for key1", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key1", Value: v1})
		require.NoError(t, err)
		v2 := &pb.Value{Data: []byte("key2 test"), Metadata: "metadata for key2", Version: 1}
		_, err = keeperClient.Put(ctx, &pb.PutReq{Key: "key2", Value: v2})
		require.NoError(t, err)
		getAllResp, err := keeperClient.GetAll(ctx, &pb.GetAllReq{Version: 0})
		require.NoError(t, err)
		assert.Len(t, getAllResp.Result, 2)
		_, err = keeperClient.GetAll(ctx, &pb.GetAllReq{Version: getAllResp.Version})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.NotFound, e.Code())
	})
	testKeeperServer(t, "negative get all without data", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := userClient.Register(context.Background(), &pb.RegisterReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		respAuth, err := userClient.Auth(context.Background(), &pb.AuthReq{Login: "test", Password: "123"})
		require.NoError(t, err)
		md := metadata.New(map[string]string{common.AuthHeader: respAuth.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		_, err = keeperClient.GetAll(ctx, &pb.GetAllReq{Version: 0})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.NotFound, e.Code())
	})
	testKeeperServer(t, "negative get all without token", func(t *testing.T, userClient pb.UserClient, keeperClient pb.KeeperClient) {
		_, err := keeperClient.GetAll(context.Background(), &pb.GetAllReq{Version: 0})
		e, ok := status.FromError(err)
		assert.Equal(t, true, ok)
		assert.Equal(t, codes.Unauthenticated, e.Code())
	})
}
