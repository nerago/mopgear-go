package util_collection

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"iter"
	"slices"
)

type EnumMap[E EnumBaseType, V any] struct {
	content  []V
	isSet    BitSet
	len      uint8
	enumType EnumType[E]
}

func EnumMapMake[E EnumBaseType, V any](enumType EnumType[E]) EnumMap[E, V] {
	return EnumMap[E, V]{
		make([]V, enumType.NumValues()),
		BitSetMake(uint32(enumType.NumValues() - 1)),
		0,
		enumType,
	}
}

func (em *EnumMap[E, V]) initInternal() {
	if em.content == nil || em.isSet == nil {
		numValues := E(0).EnumNumValues()
		em.content = make([]V, numValues)
		em.isSet = BitSetMake(uint32(numValues - 1))
	}
}

func (em *EnumMap[E, V]) Clone() *EnumMap[E, V] {
	return &EnumMap[E, V]{
		slices.Clone(em.content),
		em.isSet.Clone(),
		em.len,
		em.enumType,
	}
}

func (em *EnumMap[E, V]) Clear() {
	var nilValue V
	for index := range em.isSet.SeqIsSet() {
		em.content[index] = nilValue
	}
	em.isSet.ClearAll()
}

func (em *EnumMap[E, V]) Size() int {
	return int(em.len)
}

func (em *EnumMap[E, V]) IsEmpty() bool {
	return em.len == 0
}

func (em *EnumMap[E, V]) Equals(other *EnumMap[E, V], elementEqual func(*V, *V) bool) bool {
	if em.len != other.len {
		return false
	} else if em.len == 0 && other.len == 0 {
		return true
	}

	for i := range em.isSet {
		if em.isSet[i] != other.isSet[i] {
			return false
		}
	}

	for i := range em.content {
		if !elementEqual(&em.content[i], &other.content[i]) {
			return false
		}
	}
	return true
}

func (em *EnumMap[E, V]) EqualsInterface(other IMap[E, V], elementEqual func(*V, *V) bool) bool {
	if asType, isType := other.(*EnumMap[E, V]); isType {
		return em.Equals(asType, elementEqual)
	} else {
		return IMapEquals(em, other, elementEqual)
	}
}

func (em *EnumMap[E, V]) Has(key E) bool {
	if em.isSet == nil {
		return false
	}
	return em.isSet.IsSet(uint32(key))
}

func (em *EnumMap[E, V]) Get(key E) (V, bool) {
	if em.isSet != nil && em.isSet.IsSet(uint32(key)) {
		return em.content[key], true
	} else {
		var nilValue V
		return nilValue, false
	}
}

func (em *EnumMap[E, V]) GetOrPanic(key E) V {
	if em.isSet != nil && em.isSet.IsSet(uint32(key)) {
		return em.content[key]
	} else {
		panic("key not set")
	}
}

func (em *EnumMap[E, V]) GetOrNilValue(key E) V {
	if em.isSet != nil && em.isSet.IsSet(uint32(key)) {
		return em.content[key]
	} else {
		var nilValue V
		return nilValue
	}
}

func (em *EnumMap[E, V]) GetOrDefault(key E, defaultValue V) V {
	if em.isSet != nil && em.isSet.IsSet(uint32(key)) {
		return em.content[key]
	} else {
		return defaultValue
	}
}

func (em *EnumMap[E, V]) Put(key E, value V) {
	em.initInternal()
	if !em.isSet.SetReturningOld(uint32(key)) {
		em.len++
	}
	em.content[key] = value
}

func (em *EnumMap[E, V]) Compute(key E, apply func(V) V) {
	em.initInternal()
	if em.isSet.SetReturningOld(uint32(key)) {
		em.content[key] = apply(em.content[key])
	} else {
		var nilValue V
		em.content[key] = apply(nilValue)
		em.len++
	}
}

func (em *EnumMap[E, V]) Delete(key E) {
	if em.isSet != nil && em.isSet.ClearReturningOld(uint32(key)) {
		var nilValue V
		em.content[key] = nilValue
		em.len--
	}
}

func (em *EnumMap[E, V]) Foreach(apply func(key1 E, value V)) {
	for index := range em.isSet.SeqIsSet() {
		apply(E(index), em.content[index])
	}
}

func (em *EnumMap[E, V]) SeqKeyValue() iter.Seq2[E, V] {
	return func(yield func(E, V) bool) {
		for index := range em.isSet.SeqIsSet() {
			if !yield(E(index), em.content[index]) {
				return
			}
		}
	}
}

func (em *EnumMap[E, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for index := range em.isSet.SeqIsSet() {
			if !yield(em.content[index]) {
				return
			}
		}
	}
}

func (em *EnumMap[E, V]) SeqKey() iter.Seq[E] {
	return func(yield func(E) bool) {
		for index := range em.isSet.SeqIsSet() {
			if !yield(E(index)) {
				return
			}
		}
	}
}

func (em *EnumMap[E, V]) FirstKey() E {
	for index := range em.isSet.SeqIsSet() {
		return E(index)
	}
	panic("no key found")
}

func (em *EnumMap[E, V]) KeySlice() []E {
	slice := make([]E, em.len)
	write := 0
	for index := range em.isSet.SeqIsSet() {
		slice[write] = E(index)
		write++
	}
	return slice[:write]
}

func (em *EnumMap[E, V]) ValueSlice() []V {
	slice := make([]V, em.len)
	write := 0
	for index := range em.isSet.SeqIsSet() {
		slice[write] = em.content[index]
		write++
	}
	return slice[:write]
}

func (em *EnumMap[E, V]) MarshalJSONTo(outputEncoder *jsontext.Encoder) error {
	//buffer := bytes.Buffer{}
	//innerEncoder := jsontext.NewEncoder(&buffer)
	//err := em.marshalToEncoder(innerEncoder)
	//if err != nil {
	//	return err
	//}
	//return outputEncoder.WriteValue(buffer.Bytes())

	return em.marshalToEncoder(outputEncoder)
}

func (em *EnumMap[E, V]) UnmarshalJSONFrom(inputDecoder *jsontext.Decoder) error {
	//value, err := inputDecoder.ReadValue()
	//if err != nil {
	//	return err
	//}
	//buffer := bytes.Buffer{}
	//buffer.Write(value)
	//decoder := jsontext.NewDecoder(&buffer)
	//return em.unmarshalFromDecoder(decoder)

	return em.unmarshalFromDecoder(inputDecoder)
}

func (em *EnumMap[E, V]) marshalToEncoder(encoder *jsontext.Encoder) error {
	err := encoder.WriteToken(jsontext.BeginObject)
	if err != nil {
		return err
	}

	for index := range em.isSet.SeqIsSet() {
		key := E(index)
		value := em.content[index]

		err = encoder.WriteToken(jsontext.String(key.Name()))
		if err != nil {
			return err
		}

		err = json.MarshalEncode(encoder, value)
		if err != nil {
			return err
		}
	}

	err = encoder.WriteToken(jsontext.EndObject)
	if err != nil {
		return err
	}

	return nil
}

func (em *EnumMap[E, V]) unmarshalFromDecoder(decoder *jsontext.Decoder) error {
	em.initInternal()

	token, err := decoder.ReadToken()
	if err != nil {
		return err
	}
	if token.Kind() != jsontext.KindBeginObject {
		return errors.New("expected object start")
	}

	for {
		token, err = decoder.ReadToken()
		if err != nil {
			return err
		}

		var key E
		if token.Kind() == jsontext.KindEndObject {
			break
		} else if token.Kind() == jsontext.KindString {
			enumName := token.String()
			enumValue, foundValue := em.enumType.WithName(enumName)
			if !foundValue {
				return errors.New("invalid enum value name")
			}
			key = enumValue
		} else {
			return errors.New("expected key or object end")
		}

		var value V
		err = json.UnmarshalDecode(decoder, &value)
		if err != nil {
			return err
		}

		em.Put(key, value)
	}

	return nil
}
