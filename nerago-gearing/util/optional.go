package util

type Optional[T any] struct {
	exists bool
	value  T
}

func Optional_Empty[T any]() Optional[T] {
	return Optional[T]{}
}

func Optional_OfValue[T any](value T) Optional[T] {
	return Optional[T]{true, value}
}

func Optional_OfPointer[T any](value *T) Optional[T] {
	if value != nil {
		return Optional[T]{true, *value}
	} else {
		return Optional[T]{}
	}
}

func (opt Optional[T]) IsEmpty() bool {
	return !opt.exists
}

func (opt Optional[T]) HasValue() bool {
	return opt.exists
}

func (opt Optional[T]) GetWithFlag() (T, bool) {
	return opt.value, opt.exists
}

func (opt Optional[T]) GetAsNillable() *T {
	if opt.exists {
		return &opt.value
	} else {
		return nil
	}
}

func (opt Optional[T]) GetOrPanic() T {
	if !opt.exists {
		panic("optional is empty")
	}
	return opt.value
}

func (opt Optional[T]) GetOrDefault(defaultValue T) T {
	if opt.exists {
		return opt.value
	} else {
		return defaultValue
	}
}
