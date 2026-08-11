package util_collection

import (
	"cmp"
	"iter"
	"math/rand"
	"slices"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func RemoveDuplicatesFunc_NewIfChanged[T any](slice []T, equals func(a, b *T) bool) []T {
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
	// TODO would prefer first instance
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

func RemoveDuplicatesFunc_InPlace[T any, S ~[]T](slice *S, equals func(a, b *T) bool) {
	if slice == nil || *slice == nil {
		return
	}

	var a, b int
	for a = 0; a < len(*slice); a++ {
		for b = a + 1; b < len(*slice); b++ {
			if equals(&(*slice)[a], &(*slice)[b]) {
				goto changed
			}
		}
	}
	return

changed:
	write := a
outerLoop:
	for a = a + 1; a < len(*slice); a++ {
		for b = a + 1; b < len(*slice); b++ {
			if equals(&(*slice)[a], &(*slice)[b]) {
				continue outerLoop
			}
		}
		(*slice)[write] = (*slice)[a]
		write++
	}
	*slice = (*slice)[0:write]
}

func RemoveDuplicatesFunc_AsNew_Notify[T any](slice []T, equals func(a, b *T) bool, removedNotify func(x *T)) []T {
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

func RemoveDuplicatesComparable_InPlace[T comparable, S ~[]T](slice *S) {
	if slice == nil || *slice == nil {
		return
	}
	if len(*slice) > 8 {
		removeDuplicatesComparableInPlaceLarge(slice)
	} else {
		removeDuplicatesComparableInPlaceSmall(slice)
	}
}

func removeDuplicatesComparableInPlaceLarge[T comparable, S ~[]T](slice *S) {
	mapSet := make(map[T]bool, len(*slice))
	for _, item := range *slice {
		mapSet[item] = true
	}

	index := 0
	for k := range mapSet {
		(*slice)[index] = k
		index++
	}

	*slice = (*slice)[:index]
}

func removeDuplicatesComparableInPlaceSmall[T comparable, S ~[]T](slice *S) {
	var a, b int
	for a = 0; a < len(*slice); a++ {
		for b = a + 1; b < len(*slice); b++ {
			if (*slice)[a] != (*slice)[b] {
				goto changed
			}
		}
	}
	return

changed:
	write := a
outerLoop:
	for a = a + 1; a < len(*slice); a++ {
		for b = a + 1; b < len(*slice); b++ {
			if (*slice)[a] == (*slice)[b] {
				continue outerLoop
			}
		}
		(*slice)[write] = (*slice)[a]
		write++
	}
	*slice = (*slice)[0:write]
}

func RemoveDuplicatesComparable_NewIfChanged[T comparable](slice []T) []T {
	if slice == nil {
		return slice
	}

	mapSet := make(map[T]bool, len(slice))
	var readIndex int
	for readIndex = range slice {
		item := slice[readIndex]
		if mapSet[item] {
			goto changed
		} else {
			mapSet[item] = true
		}
	}
	return slice

changed:
	result := make([]T, readIndex, len(slice)-1)
	copy(result, slice)

	readIndex++
	for readIndex < len(slice) {
		item := slice[readIndex]
		readIndex++

		if !mapSet[item] {
			result = append(result, item)
			mapSet[item] = true
		}

	}

	return result
}

func Shuffle[T any](slice []T) {
	rand.Shuffle(len(slice), func(a, b int) { slice[a], slice[b] = slice[b], slice[a] })
}

func DeleteIndexInPlace[T any](slice *[]T, index int) {
	if index < 0 || index >= len(*slice) {
		panic("invalid index")
	}

	var nilValue T
	if index == 0 {
		(*slice)[0] = nilValue
		*slice = (*slice)[1:]
	} else if index == len(*slice)-1 {
		(*slice)[index] = nilValue
		*slice = (*slice)[:index]
	} else {
		copy((*slice)[index:], (*slice)[index+1:])
		(*slice)[len(*slice)-1] = nilValue
		*slice = (*slice)[:len(*slice)-1]
	}
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

func MapSliceAsNew_NoPointer[T any, R any](slice []T, mapper func(x T) R) []R {
	if slice == nil {
		return nil
	}

	result := make([]R, len(slice))
	for i := range slice {
		result[i] = mapper(slice[i])
	}
	return result
}

func MapSliceAsSeq[T any, R any](slice []T, mapper func(x *T) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		for i := range slice {
			value := mapper(&slice[i])
			if !yield(value) {
				return
			}
		}
	}
}

func MapSeqAsSlice[T any, R any](seq iter.Seq[T], mapper func(x T) R) []R {
	if seq == nil {
		return nil
	}

	result := make([]R, 0)
	for input := range seq {
		value := mapper(input)
		result = append(result, value)
	}
	return result
}

func MapSeq2AsSlice[T any, U any, R any](seq iter.Seq2[T, U], mapper func(T, U) R) []R {
	if seq == nil {
		return nil
	}

	result := make([]R, 0)
	for inputA, inputB := range seq {
		value := mapper(inputA, inputB)
		result = append(result, value)
	}
	return result
}

func SliceToMap[T any, K comparable, V any](slice []T, toKey func(*T) K, toValue func(*T) V) map[K]V {
	if slice == nil {
		return nil
	}

	result := make(map[K]V, len(slice))
	for i := range slice {
		key := toKey(&slice[i])
		value := toValue(&slice[i])
		result[key] = value
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

func FilterSliceAsNew_NoPointer[T any](slice []T, filter func(x T) bool) []T {
	if slice == nil {
		return slice
	}

	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if filter(item) {
			result = append(result, item)
		}
	}
	return result
}

func FilterSliceInPlace[T any](slice *[]T, filter func(x *T) bool) {
	if slice == nil || *slice == nil {
		return
	}

	readIndex := 0
	for readIndex < len(*slice) {
		if !filter(&(*slice)[readIndex]) {
			goto change_part
		}
		readIndex++
	}
	return

change_part:
	writeIndex := readIndex
	readIndex++
	for readIndex < len(*slice) {
		if filter(&(*slice)[readIndex]) {
			(*slice)[writeIndex] = (*slice)[readIndex]
			writeIndex++
		}
		readIndex++
	}
	*slice = (*slice)[:writeIndex]
}

func FilterSliceInPlace_NoPointer[T any](slice *[]T, filter func(x T) bool) {
	if slice == nil || *slice == nil {
		return
	}

	readIndex := 0
	for readIndex < len(*slice) {
		if !filter((*slice)[readIndex]) {
			goto change_part
		}
		readIndex++
	}
	return

change_part:
	writeIndex := readIndex
	readIndex++
	for readIndex < len(*slice) {
		if filter((*slice)[readIndex]) {
			(*slice)[writeIndex] = (*slice)[readIndex]
			writeIndex++
		}
		readIndex++
	}
	*slice = (*slice)[:writeIndex]
}

func ContainsFunc_Pointer[T any](slice []T, predicate func(*T) bool) bool {
	for i := range slice {
		if predicate(&slice[i]) {
			return true
		}
	}
	return false
}

func EqualFunc_Pointer[T any](one []T, two []T, equal func(a, b *T) bool) bool {
	if len(one) != len(two) {
		return false
	}
	for i := range one {
		if !equal(&one[i], &two[i]) {
			return false
		}
	}
	return true
}

func EqualFunc_IgnoreOrder_Pointer[T any](one []T, two []T, equals func(a, b *T) bool) bool {
	if len(one) != len(two) {
		return false
	}

	twoConsumed := make([]bool, len(two))
outer:
	for a := range one {
		for b := range two {
			if !twoConsumed[b] && equals(&one[a], &two[b]) {
				twoConsumed[b] = true
				continue outer
			}
		}
		return false
	}

	return true
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
		return []T{item}
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

func FindElementWithMinFunc[X any, T *X, V Number](slice []T, toValue func(T) V) T {
	if len(slice) == 0 {
		panic("empty slice")
	}
	minElement := slice[0]
	minValue := toValue(minElement)
	for i := 1; i < len(slice); i++ {
		value := toValue(slice[i])
		if value < minValue {
			minValue = value
			minElement = slice[i]
		}
	}
	return minElement
}

func FindElementWithMaxFunc[X any, T *X, V Number](slice []T, toValue func(T) V) T {
	if len(slice) == 0 {
		panic("empty slice")
	}
	maxElement := slice[0]
	maxValue := toValue(maxElement)
	for i := 1; i < len(slice); i++ {
		value := toValue(slice[i])
		if value > maxValue {
			maxValue = value
			maxElement = slice[i]
		}
	}
	return maxElement
}

func FindMinFunc[T any, V Number](slice []T, toValue func(T) V) V {
	if len(slice) == 0 {
		panic("empty slice")
	}
	minValue := toValue(slice[0])
	for i := 1; i < len(slice); i++ {
		value := toValue(slice[i])
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

func FindMaxFunc[T any, V Number](slice []T, toValue func(T) V) V {
	if len(slice) == 0 {
		panic("empty slice")
	}
	maxValue := toValue(slice[0])
	for i := 1; i < len(slice); i++ {
		value := toValue(slice[i])
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func FindAverageFunc[T any, V Number](slice []T, toValue func(T) V) V {
	if len(slice) == 0 {
		panic("empty slice")
	}
	total := V(0)
	for i := range slice {
		value := toValue(slice[i])
		total += value
	}
	return total / V(len(slice))
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
		for i := range rankArray {
			entry := &rankArray[i]
			if !yield(entry.pointer, i) {
				return
			}
		}
	}
}

func SliceSampleFromStart[T any](slice []T, size int) []T {
	if len(slice) < size {
		return slice
	} else {
		return slice[0:size]
	}
}

func SliceSampleRandom[T any](slice []T, size int) []T {
	if len(slice) < size {
		return slice
	} else {
		sample := slices.Clone(slice)
		Shuffle(sample)
		return sample[0:size]
	}
}

func SliceSampleRandom_Seed[T any](slice []T, size int, seed int64) []T {
	if len(slice) < size {
		return slice
	} else {
		rng := rand.New(rand.NewSource(seed))

		sample := slices.Clone(slice)
		rng.Shuffle(len(sample), func(a, b int) { sample[a], sample[b] = sample[b], sample[a] })
		return sample[0:size]
	}
}
