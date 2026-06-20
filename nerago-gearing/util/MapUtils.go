package util

import (
	"maps"
)

func CombineMaps[K comparable, V any](parts ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, each := range parts {
		maps.Copy(result, each)
	}
	return result
}

func KeysToSlice[K comparable, V any](m map[K]V) []K {
	slice := make([]K, len(m))
	i := 0
	for k := range m {
		slice[i] = k
		i++
	}
	return slice
}

func MapFirstEntry[K comparable, V any](m map[K]V) (K, V) {
	for k, v := range m {
		return k, v
	}
	panic("empty map")
}
