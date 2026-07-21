package util

import "iter"

type BitSet []uint32

func BitSetMake(maxIndex uint32) BitSet {
	elementCount := (maxIndex + 32) / 32
	return make([]uint32, elementCount)
}

func (bs *BitSet) IsSet(index uint32) bool {
	i := index >> 5
	b := index & 0x1f
	element := (*bs)[i]

	value := (element >> b) & 0x1
	return value != 0
}

func (bs *BitSet) Set(index uint32) {
	i := index >> 5
	b := index & 0x1f

	var mod uint32 = 1 << b
	(*bs)[i] |= mod
}

func (bs *BitSet) Clear(index uint32) {
	i := index >> 5
	b := index & 0x1f

	var mod uint32 = ^(1 << b)
	(*bs)[i] &= mod
}

func (bs *BitSet) SeqIsSet() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		for i, value := range *bs {
			for b := range 32 {
				if value == 0 {
					break
				} else if (value & 0x1) != 0 {
					index := i*32 + b
					if !yield(uint32(index)) {
						return
					}
				}
				value >>= 1
			}
		}
	}
}
