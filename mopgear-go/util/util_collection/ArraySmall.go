package util_collection

import (
	"reflect"
)

type ArraySmall[T any] struct {
	internal *[0x10000]T
}

func ArraySmallMake[T any](size uint16) ArraySmall[T] {
	arrayType := reflect.ArrayOf(int(size), reflect.TypeFor[T]())
	arrayPointer := reflect.New(arrayType)
	internal := (*[0x10000]T)(arrayPointer.UnsafePointer())
	return ArraySmall[T]{internal}
}

func (a *ArraySmall[T]) Get(index uint16) T {
	return a.internal[index]
}

func (a *ArraySmall[T]) Put(index uint16, value T) {
	a.internal[index] = value
}

type ArrayTinyParam[T any] interface {
	[1]T | [2]T | [3]T | [4]T | [5]T | [6]T | [7]T | [8]T | [9]T | [10]T | [11]T | [12]T | [13]T | [14]T | [15]T | [16]T
}

type ArrayTiny[T any, A ArrayTinyParam[T]] struct {
	array A
}

func (a *ArrayTiny[T, A]) Get(index uint8) T {
	return a.array[index]
}

func (a *ArrayTiny[T, A]) Put(index uint8, value T) {
	a.array[index] = value
}
