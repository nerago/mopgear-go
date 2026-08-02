package util_collection

type sample uint8

func (s sample) Name() string {
	switch s {
	case Sample_One:
		return "one"
	case Sample_Two:
		return "two"
	case Sample_Three:
		return "three"
	}
	return ""
}

func (s sample) EnumNumValues() uint8 {
	return uint8(len(sampleList))
}

const (
	Sample_One      sample = iota
	Sample_Two      sample = iota
	Sample_Three    sample = iota
	Sample_EnumSize        = 3
)

var sampleList = []sample{Sample_One, Sample_Two, Sample_Three}
var sampleEnum = EnumTypeMake(sampleList)

type enumMap2Sample[V any] struct {
	EnumMap2[sample, V, [Sample_EnumSize]V, [Sample_EnumSize]bool]
}

type EnumMap2[E EnumBaseType, V any, A ArrayTinyParam[V], B ArrayTinyParam[bool]] struct {
	content  A
	isSet    B
	len      uint8
	enumType EnumType[E]
}

func (em *EnumMap2[E, V, A, B]) Size() int {
	return int(em.len)
}

func (em *EnumMap2[E, V, A, B]) GetOrPanic(key E) V {
	if em.isSet[key] {
		return em.content[key]
	} else {
		panic("key not set")
	}
}

func (em *EnumMap2[E, V, A, B]) Put(key E, value V) {
	if !em.isSet[key] {
		em.len++
	}
	em.content[key] = value
	em.isSet[key] = true
}

func (em *EnumMap2[E, V, A, B]) Delete(key E) {
	if em.isSet[key] {
		var nilValue V
		em.content[key] = nilValue
		em.isSet[key] = false
		em.len--
	}
}
