package main

func contains[T comparable](s []T, e T) bool {
	for _, elem := range s {
		if elem == e {
			return true
		}
	}
	return false
}
