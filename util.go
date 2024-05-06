package main

import (
	"crypto/rand"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func contains[T comparable](s []T, e T) bool {
	for _, elem := range s {
		if elem == e {
			return true
		}
	}
	return false
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func recoveryHandler(p any) error {
	slog.Error("PANIC", "p", p)
	return status.Error(codes.Internal, "internal panic")
}
