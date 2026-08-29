package util

import (
	"reflect"
	"sync"
)

func IsNil(a any) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}

func NilSafeEqual[T any](one *T, two *T, equals func(T, T) bool) bool {
	if one != nil && two != nil {
		return equals(*one, *two)
	} else if one == nil && two == nil {
		return true
	} else {
		return false
	}
}

func NilSafeEqualPointers[T any](one *T, two *T, equals func(*T, *T) bool) bool {
	if one != nil && two != nil {
		return equals(one, two)
	} else if one == nil && two == nil {
		return true
	} else {
		return false
	}
}

func NilSafeEqualComparable[T comparable](one *T, two *T) bool {
	if one != nil && two != nil {
		return *one == *two
	} else if one == nil && two == nil {
		return true
	} else {
		return false
	}
}

type TypedPool[T any] struct {
	pool sync.Pool
}

func (p *TypedPool[T]) Get() *T {
	value := p.pool.Get()
	if value != nil {
		return value.(*T)
	} else {
		return new(T)
	}
}

func (p *TypedPool[T]) Put(instance *T) {
	p.pool.Put(instance)
}
