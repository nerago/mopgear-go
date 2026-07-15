package util

import (
	"slices"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

type StringBuild2 []byte

func (sb *StringBuild2) Reset() {
	*sb = nil
}

func (sb *StringBuild2) String() string {
	if len(*sb) > 0 {
		return unsafe.String(unsafe.SliceData(*sb), len(*sb))
	} else {
		return ""
	}
}

func (sb *StringBuild2) Len() int {
	return len(*sb)
}

func (sb *StringBuild2) WriteBuilder(other StringBuild2) {
	*sb = append(*sb, other...)
}

func (sb *StringBuild2) WriteString(value string) {
	*sb = append(*sb, value...)
}

func (sb *StringBuild2) WriteBytes(bytes []byte) {
	*sb = append(*sb, bytes...)
}

func (sb *StringBuild2) WriteRune(value rune) {
	*sb = utf8.AppendRune(*sb, value)
}

func (sb *StringBuild2) WriteInt64(value int64) {
	*sb = strconv.AppendInt(*sb, value, 10)
}

func (sb *StringBuild2) WriteUint64(value uint64) {
	*sb = strconv.AppendUint(*sb, value, 10)
}

func (sb *StringBuild2) WriteUint32(value uint32) {
	*sb = strconv.AppendUint(*sb, uint64(value), 10)
}

func (sb *StringBuild2) WriteUint16(value uint16) {
	*sb = strconv.AppendUint(*sb, uint64(value), 10)
}

func (sb *StringBuild2) WriteInt(value int) {
	*sb = strconv.AppendInt(*sb, int64(value), 10)
}

func (sb *StringBuild2) WriteInt32(value int32) {
	*sb = strconv.AppendInt(*sb, int64(value), 10)
}

func (sb *StringBuild2) WriteFloat64(value float64, decimalPlaces int) {
	*sb = strconv.AppendFloat(*sb, value, 'f', decimalPlaces, 64)
}

func (sb *StringBuild2) WriteFloat32(value float32, decimalPlaces int) {
	*sb = strconv.AppendFloat(*sb, float64(value), 'f', decimalPlaces, 32)
}

func (sb *StringBuild2) WriteFloat64_RightPadded(value float64, decimalPlaces int, pad int) {
	*sb = slices.Grow(*sb, pad)

	oldSize := len(*sb)
	*sb = strconv.AppendFloat(*sb, value, 'f', decimalPlaces, 64)
	written := len(*sb) - oldSize

	remainingPad := pad - written
	for remainingPad > 0 {
		*sb = append(*sb, ' ')
		remainingPad--
	}
}

func (sb *StringBuild2) Rewind(removeFromEnd int) {
	if len(*sb) < removeFromEnd {
		panic("can't rewind bytes that don't exist")
	}
	*sb = (*sb)[0 : len(*sb)-removeFromEnd]
}
