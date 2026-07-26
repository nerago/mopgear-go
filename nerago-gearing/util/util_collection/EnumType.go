package util_collection

type EnumBaseType interface {
	~uint8
	Name() string
	EnumNumValues() uint8
}

type EnumType[E EnumBaseType] struct {
}

func EnumTypeMake[E EnumBaseType](values []E) EnumType[E] {
	for i := 0; i < len(values); i++ {
		if values[i] != E(i) {
			panic("expected iota enum in order")
		}
	}
	return EnumType[E]{}
}

func (et EnumType[E]) NumValues() uint8 {
	return E(0).EnumNumValues()
}

func (et EnumType[E]) First() E {
	return E(0)
}

func (et EnumType[E]) Last() E {
	return E(et.NumValues() - 1)
}

func (et EnumType[E]) WithName(name string) (E, bool) {
	for i := range et.NumValues() {
		value := E(i)
		if value.Name() == name {
			return value, true
		}
	}
	return E(0), false
}
