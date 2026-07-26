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
