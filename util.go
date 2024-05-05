package main

import "crypto/rand"

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
