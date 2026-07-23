package util_collection

import (
	"iter"
	"math"
)

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

func (bs *BitSet) IsAnySetInRange(minIndex, maxIndex uint32) bool {
	minElementNum := minIndex >> 5
	maxElementNum := maxIndex >> 5
	minBitNum := minIndex & 0x1f
	maxBitNum := maxIndex & 0x1f
	minMask := uint32(0xffffffff) << minBitNum
	maxMask := uint32(0xffffffff) >> (31 - maxBitNum)

	if minElementNum == maxElementNum {
		value := (*bs)[minElementNum]
		return (value & minMask & maxMask) != 0
	}

	valueMin := (*bs)[minElementNum]
	if (valueMin & minMask) != 0 {
		return true
	}

	for i := minElementNum + 1; i < maxElementNum; i++ {
		if (*bs)[i] != 0 {
			return true
		}
	}

	valueMax := (*bs)[maxElementNum]
	return (valueMax & maxMask) != 0
}

func (bs *BitSet) ClearInRange(minIndex, maxIndex uint32) {
	minElementNum := minIndex >> 5
	maxElementNum := maxIndex >> 5
	minBitNum := minIndex & 0x1f
	maxBitNum := maxIndex & 0x1f
	minMask := uint32(0xffffffff) >> (32 - minBitNum)
	maxMask := uint32(0xffffffff) << (maxBitNum + 1)

	if minElementNum == maxElementNum {
		(*bs)[minElementNum] &= minMask | maxMask
		return
	}

	(*bs)[minElementNum] &= minMask

	for i := minElementNum + 1; i < maxElementNum; i++ {
		(*bs)[i] = 0
	}

	(*bs)[maxElementNum] &= maxMask
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

func (bs *BitSet) ClearAll() {
	for i := range *bs {
		(*bs)[i] = 0
	}
}

func (bs *BitSet) FirstIndexIsSet() (uint32, bool) {
	for i, value := range *bs {
		for b := range 32 {
			if value == 0 {
				break
			} else if (value & 0x1) != 0 {
				index := i*32 + b
				return uint32(index), true
			}
			value >>= 1
		}
	}
	return math.MaxUint32, false
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
