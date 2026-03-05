package util

// type Equatable[T any] interface {
// 	Equals(other T)
// }

func RemoveDuplicatesFunc[T any](slice []T, equals func(a, b *T) bool) []T {
	result := make([]T, 0, len(slice))
outer:
	for _, a := range slice {
		for _, b := range result {
			if equals(&a, &b) {
				continue outer
			}
		}
		result = append(result, a)
	}
	return result
}

func RemoveDuplicatesComparable[T comparable](slice []T) []T {
	if slice == nil {
		return slice
	}

	mapSet := make(map[T]bool)
	for _, item := range slice {
		mapSet[item] = true
	}

	index := 0
	for k := range mapSet {
		slice[index] = k
		index++
	}

	return slice[:index]
}

func FilterSlice[T any](slice []T, filter func(x *T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if filter(&item) {
			result = append(result, item)
		}
	}
	return result
}
