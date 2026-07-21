package util

import "iter"

type BitSet []uint32

func BitSetMake(maxIndex int) BitSet {
	elementCount := (maxIndex + 32) / 32
	return make([]uint32, elementCount)
}

func (bs *BitSet) IsSet(index int) bool {
	i := index >> 5
	b := index & 0x1f
	element := (*bs)[i]

	value := (element >> b) & 0x1
	return value != 0
}

func (bs *BitSet) Set(index int) {
	i := index >> 5
	b := index & 0x1f

	mod := 1 << b
	(*bs)[i] |= mod
}

func (bs *BitSet) Clear(index int) {
	i := index >> 5
	b := index & 0x1f

	mod := ^(1 << b)
	(*bs)[i] &= mod
}

func (bs *BitSet) SeqIsSet() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i, value := range *bs {
			for b := range 32 {
				if value == 0 {
					break
				} else if (value & 0x1) != 0 {
					index := i*32 + b
					if !yield(index) {
						return
					}
				}
				value >>= 1
			}
		}
	}
}
