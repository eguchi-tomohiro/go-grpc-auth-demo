package server

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PermissionHello  = "HELLO"
	PermissionSecret = "SECRET"
)

var routes = map[string][]string {
	"/pb.HelloService/Hello": {PermissionHello},
	"/pb.HelloService/TellMeSecret": {PermissionSecret},
}


type User struct {
	permissions []string
}

func AuthorizationUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// info.FullMethodにメソッドのフルパスが入っている。
		if canAccess(info.FullMethod, getUser(GetToken(ctx).Subject)) {
			return handler(ctx, req)
		}
		return nil, status.Error(
			codes.PermissionDenied,
			"could not access to specified method",
		)
	}
}

func getUser(id string) *User {
	// 本番ではDBから取得する。
	switch id {
	case "alice":
		return &User{permissions: []string{PermissionHello}}
	case "bob":
		return &User{permissions: []string{PermissionHello, PermissionSecret}}
	}
	return &User{}
}

func canAccess(method string, user *User) bool {
	r, ok := routes[method]
	if !ok {
		return false
	}

	// 検索しやすいように詰めなおす。
	permissions := map[string]bool{}
	for _, p := range user.permissions {
		permissions[p] = true
	}

	// 1つでも保有していない権限があればfalseとする。
	for _, p := range r {
		if !permissions[p] {
			return false
		}
	}

	return true
}
