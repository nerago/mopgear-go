package util

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
