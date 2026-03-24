package util

import "maps"

func CombineMaps[K comparable, V any](parts ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, each := range parts {
		maps.Copy(result, each)
	}
	return result
}
