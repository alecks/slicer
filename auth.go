package main

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/alecks/slicer/proto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/uptrace/bun"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var authExcludes = map[string][]string{
	pb.AuthService_ServiceDesc.ServiceName: {"Authenticate"},
	pb.MetaService_ServiceDesc.ServiceName: {"Info"},
}

var jwtValidMethods = []string{"HS256"}

type userClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type serviceContext struct {
	db *bun.DB
}

type authService struct {
	pb.UnimplementedAuthServiceServer
	c *serviceContext
}

func (s *authService) Authenticate(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {
	if in.Password != "test" {
		return nil, status.Error(codes.Unauthenticated, "username or password incorrect") // TODO: make errors more predictable
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims{
		UserID: in.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // TODO: lower this time / make it configurable
		},
	})

	tokenText, err := token.SignedString(jwtSecret)
	if err != nil {
		slog.Error("failed to sign token", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.AuthResponse{
		Token: tokenText,
	}, nil
}

func authInterceptor(ctx context.Context) (context.Context, error) {
	rawToken, err := auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "could not get bearer token in metadata")
	}

	claims := userClaims{}
	_, err = jwt.ParseWithClaims(rawToken, &claims, getJwtSecret, jwt.WithValidMethods(jwtValidMethods))
	if err != nil {
		slog.Debug("failed to parse token", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return context.WithValue(ctx, ctxClaims, claims), nil
}

func requireAuth(ctx context.Context, callMeta interceptors.CallMeta) bool {
	auth := true
	for service, methods := range authExcludes {
		if callMeta.Service == service && contains(methods, callMeta.Method) {
			auth = false
		}
	}
	return auth
}

func getJwtSecret(t *jwt.Token) (any, error) {
	return jwtSecret, nil
}
