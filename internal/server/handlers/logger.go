package handlers

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

func LogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("Got request: ", info.FullMethod)
	v, err := handler(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			log.Printf("Response code: %s, message: %s\n", e.Code(), e.Message())
		} else {
			log.Println("Unknown response code")
		}
	}
	return v, err
}

func LogWithUserID(ctx context.Context, f string, v ...any) {
	userID := GetUserIDFromContext(ctx)
	log.Printf("[%s] "+f+"\n", strconv.Itoa(int(userID)), v)
}

func LogErrorWithUserID(ctx context.Context, msg string, e error) {
	userID := GetUserIDFromContext(ctx)
	log.Printf("[ERROR] [%s] "+msg+": %s", strconv.Itoa(int(userID)), e.Error())
}
