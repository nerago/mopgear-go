package util

import (
	"iter"
)

// type Equatable[T any] interface {
// 	Equals(other T)
// }

func RemoveDuplicatesFunc[T any](slice []T, equals func(a, b *T) bool) []T {
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
	result := make([]T, len(slice))
	for i := range slice {
		result[i] = mapper(&slice[i])
	}
	return result
}

func FilterSliceAsNew[T any](slice []T, filter func(x *T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if filter(&item) {
			result = append(result, item)
		}
	}
	return result
}

func FilterSliceInPlace[T any](slice []T, filter func(x *T) bool) []T {
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

func PermuteAll[T any](sliceOfSlices [][]T) iter.Seq[[]T] {
	index := 0
	progress := make([][]T, 0, len(sliceOfSlices[index]))
	for _, item := range sliceOfSlices[index] {
		progress = append(progress, []T{item})
	}
	index++

	for index < len(sliceOfSlices)-1 {
		next := make([][]T, 0, len(sliceOfSlices[index])*len(progress))
		for _, item := range sliceOfSlices[index] {
			for _, curr := range progress {
				list := copyAndAppend(curr, item)
				next = append(next, list)
			}
		}
		index++
		progress = next
	}

	return func(yield func([]T) bool) {
		for _, item := range sliceOfSlices[index] {
			for _, curr := range progress {
				list := copyAndAppend(curr, item)
				if !yield(list) {
					return
				}
			}
		}
	}
}

func copyAndAppend[T any](curr []T, item T) []T {
	list := make([]T, len(curr)+1)
	copy(list, curr)
	list[len(curr)] = item
	return list
}
