package util

// type Equatable[T any] interface {
// 	Equals(other T)
// }

func RemoveDuplicatesFunc[T any](slice []T, equals func(a, b *T) bool) []T {
	if slice == nil {
		return nil
	}

	result := make([]T, 0, len(slice))
outer:
	for outerIndex := range slice {
		next := &slice[outerIndex]
		for checkIndex := range result {
			if equals(next, &result[checkIndex]) {
				continue outer
			}
		}
		result = append(result, *next)
	}
	return result
}

func RemoveDuplicatesFuncNotify[T any](slice []T, equals func(a, b *T) bool, removedNotify func(x *T)) []T {
	if slice == nil {
		return nil
	}

	result := make([]T, 0, len(slice))
outer:
	for outerIndex := range slice {
		next := &slice[outerIndex]
		for checkIndex := range result {
			if equals(next, &result[checkIndex]) {
				removedNotify(next)
				continue outer
			}
		}
		result = append(result, *next)
	}
	return result
}

func RemoveDuplicatesComparable[T comparable](slice []T) []T {
	if slice == nil {
		return slice
	}

	mapSet := make(map[T]bool, len(slice))
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

func MapSliceAsNew[T any](slice []T, mapper func(x *T) T) []T {
	if slice == nil {
		return nil
	}

	result := make([]T, len(slice))
	for i := range slice {
		result[i] = mapper(&slice[i])
	}
	return result
}

func CastSliceAsNew[T any, R any](slice []T, mapper func(x *T) R) []R {
	if slice == nil {
		return nil
	}

	result := make([]R, len(slice))
	for i := range slice {
		result[i] = mapper(&slice[i])
	}
	return result
}

func FilterSliceAsNew[T any](slice []T, filter func(x *T) bool) []T {
	if slice == nil {
		return slice
	}

	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if filter(&item) {
			result = append(result, item)
		}
	}
	return result
}

func FilterSliceInPlace[T any](slice []T, filter func(x *T) bool) []T {
	if slice == nil {
		return nil
	}

	readIndex := 0
	for readIndex < len(slice) {
		if !filter(&slice[readIndex]) {
			goto change_part
		}
		readIndex++
	}
	return slice

change_part:
	writeIndex := readIndex
	readIndex++
	for readIndex < len(slice) {
		if filter(&slice[readIndex]) {
			slice[writeIndex] = slice[readIndex]
			writeIndex++
		}
		readIndex++
	}
	return slice[:writeIndex]
}

func RepeatValue[T any](value T, count int) []T {
	result := make([]T, count)
	for i := range count {
		result[i] = value
	}
	return result
}

func CopyAndAppend[T any](curr []T, item T) []T {
	if curr != nil {
		list := make([]T, len(curr)+1)
		copy(list, curr)
		list[len(curr)] = item
		return list
	} else {
		list := make([]T, 1)
		list[0] = item
		return list
	}
}
