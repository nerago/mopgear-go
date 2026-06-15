package util

import (
	"cmp"
	"iter"
	"paladin_gearing_go/util/channel_op"
	"slices"
	"sync"
)

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

func RemoveDuplicatesFunc2[T any](slice []T, equals func(a, b *T) bool) []T {
	if slice == nil {
		return nil
	}

	var a, b int
	for a = 0; a < len(slice); a++ {
		for b = a + 1; b < len(slice); b++ {
			if equals(&slice[a], &slice[b]) {
				goto changed
			}
		}
	}
	return slice

changed:
	result := make([]T, 0, len(slice)-1)
	result = append(result, slice[0:a]...)
outerLoop:
	for a = a + 1; a < len(slice); a++ {
		for b = a + 1; b < len(slice); b++ {
			if equals(&slice[a], &slice[b]) {
				continue outerLoop
			}
		}
		result = append(result, slice[a])
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

func RemoveDuplicatesFunc_Channels[T any](inputChannel <-chan T, equals func(a, b *T) bool) <-chan T {
	lock := sync.Mutex{}
	seen := make([]T, 0)

	return channel_op.Map_ChannelToChannel(2, inputChannel, func(next T, outputChannel chan<- T) {
		lock.Lock()
		defer lock.Unlock()

		for checkIndex := range seen {
			if equals(&next, &seen[checkIndex]) {
				return
			}
		}

		seen = append(seen, next)
		outputChannel <- next
	})
}

func RemoveDuplicatesFuncNotify_Channels[T any](inputChannel <-chan T, equals func(a, b *T) bool, removedNotify func(x *T)) <-chan T {
	lock := sync.Mutex{}
	seen := make([]T, 0)

	return channel_op.Map_ChannelToChannel(2, inputChannel, func(next T, outputChannel chan<- T) {
		lock.Lock()
		defer lock.Unlock()

		for checkIndex := range seen {
			if equals(&next, &seen[checkIndex]) {
				removedNotify(&next)
				return
			}
		}

		seen = append(seen, next)
		outputChannel <- next
	})
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

func DeleteIndexInPlace[T any](slice []T, index int) []T {
	if index == 0 {
		clear(slice[0:1])
		return slice[1:]
	} else if len := len(slice); index == len-1 {
		clear(slice[index:len])
		return slice[:index]
	} else {
		copy(slice[index:], slice[index+1:])
		clear(slice[len-1 : len])
		return slice[:len-1]
	}
	// return slices.Delete(slice, index, index+1)
}

func MapSliceAsNew[T any, R any](slice []T, mapper func(x *T) R) []R {
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

// func FilterSliceInPlace[T any](slice []T, filter func(x *T) bool) []T {
// 	if slice == nil {
// 		return nil
// 	}

// 	readIndex := 0
// 	for readIndex < len(slice) {
// 		if !filter(&slice[readIndex]) {
// 			goto change_part
// 		}
// 		readIndex++
// 	}
// 	return slice

// change_part:
// 	writeIndex := readIndex
// 	readIndex++
// 	for readIndex < len(slice) {
// 		if filter(&slice[readIndex]) {
// 			slice[writeIndex] = slice[readIndex]
// 			writeIndex++
// 		}
// 		readIndex++
// 	}
// 	return slice[:writeIndex]
// }

func ContainsFunc_Pointer[T any](slice []T, predicate func(*T) bool) bool {
	for i := range slice {
		if predicate(&slice[i]) {
			return true
		}
	}
	return false
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

func PermuteAll_Slice[T any](listsOfOptions [][]T) [][]T {
	return permuteRecur_Slice(listsOfOptions, nil, make([][]T, 0))
}

func permuteRecur_Slice[T any](listsOfOptions [][]T, curr []T, slice [][]T) [][]T {
	if len(listsOfOptions) == 0 {
		slice = append(slice, curr)
	} else {
		for _, opt := range listsOfOptions[0] {
			next := CopyAndAppend(curr, opt)
			slice = permuteRecur_Slice(listsOfOptions[1:], next, slice)
		}
	}
	return slice
}

func PermuteAll_Seq[T any](listsOfOptions [][]T) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		permuteRecur_Seq(listsOfOptions, nil, yield)
	}
}

func permuteRecur_Seq[T any](listsOfOptions [][]T, curr []T, yield func([]T) bool) bool {
	if len(listsOfOptions) == 0 {
		return yield(curr)
	} else {
		for _, opt := range listsOfOptions[0] {
			next := CopyAndAppend(curr, opt)
			more := permuteRecur_Seq(listsOfOptions[1:], next, yield)
			if !more {
				return false
			}
		}
		return true
	}
}

func ForPointer[T any](slice []T) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for i := range slice {
			if !yield(&slice[i]) {
				return
			}
		}
	}
}

func FindWith[T any](slice []T, check func(T) bool) T {
	for _, item := range slice {
		if check(item) {
			return item
		}
	}
	panic("not found")
}

// guarantees that each ranking number is used in range, even given duplicate numbers
func CalculateRanking[T any](highGood bool, inputData []T, toScore func(*T) float64) iter.Seq2[*T, int] {
	type internalEntry struct {
		score   float64
		pointer *T
	}

	rankArray := make([]internalEntry, len(inputData))
	for i := range len(inputData) {
		pointer := &inputData[i]
		rankArray[i] = internalEntry{
			toScore(pointer),
			pointer,
		}
	}

	if highGood {
		slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(a.score, b.score) })
	} else {
		slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(b.score, a.score) })
	}

	return func(yield func(*T, int) bool) {
		for i := range len(inputData) {
			entry := &rankArray[i]
			if !yield(entry.pointer, i) {
				return
			}
		}
	}
}

type HiLoInt struct {
	Lo int
	Hi int
}

func (hilo HiLoInt) Mid() int {
	return (hilo.Hi + hilo.Lo) / 2
}

func (hilo HiLoInt) Overlap(other HiLoInt) bool {
	return hilo.Between(other.Lo) || hilo.Between(other.Hi) || other.Between(hilo.Lo) || other.Between(hilo.Hi)
}

func (hilo HiLoInt) Between(check int) bool {
	return hilo.Lo <= check && check <= hilo.Hi
}

// guarantees that each ranking number is used in range, even given duplicate numbers
func CalculateRankingRanges[T any](highGood bool, inputData []T, toScore func(*T) float64) iter.Seq2[*T, HiLoInt] {
	if len(inputData) == 0 {
		return func(yield func(*T, HiLoInt) bool) {}
	}

	type internalEntry struct {
		score   float64
		pointer *T
	}

	rankArray := make([]internalEntry, len(inputData))
	for i := range len(inputData) {
		pointer := &inputData[i]
		rankArray[i] = internalEntry{
			toScore(pointer),
			pointer,
		}
	}

	if highGood {
		slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(a.score, b.score) })
	} else {
		slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(b.score, a.score) })
	}

	grouped := make(map[float64]HiLoInt)
	for rankIndex := range rankArray {
		entryScore := rankArray[rankIndex].score
		if hiLo, hasExisting := grouped[entryScore]; hasExisting {
			hiLo.Hi = rankIndex
			grouped[entryScore] = hiLo
		} else {
			grouped[entryScore] = HiLoInt{Lo: rankIndex, Hi: rankIndex}
		}
	}

	return func(yield func(*T, HiLoInt) bool) {
		for i := range rankArray {
			entry := &rankArray[i]
			score := entry.score
			hiLo := grouped[score]
			if !yield(entry.pointer, hiLo) {
				return
			}
		}
	}
}
