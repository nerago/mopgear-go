package util

func CombineMaps[K comparable, V any](parts ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, each := range parts {
		for k, v := range each {
			result[k] = v
		}
	}
	return result
}
