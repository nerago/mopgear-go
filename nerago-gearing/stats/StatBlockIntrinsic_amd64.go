package stats

import (
	"simd/archsimd"
	"unsafe"
)

func StatBlock_Add_Into(a, b, out *StatBlock)

func StatBlock_Increment_Mutating(mutate *StatBlock, other *StatBlock)

func StatBlock_Equals(a, b *StatBlock) bool

func StatBlock_AddAndSubtract_Into(add1, add2, subtract, out *StatBlock)

func StatBlock_MultiplyForTotalSum(a, b *StatBlock) float64

func StatBlock_StatBlockFloat_MultiplyForTotalSum3(a *StatBlockFloat, b *StatBlock) float64

func StatBlock_Add_Into_Experiment(a, b, out *StatBlock) {
	a1 := archsimd.LoadUint32x8Slice(a[0:8])
	a2 := archsimd.LoadUint32x4Slice(a[8:12])
	b1 := archsimd.LoadUint32x8Slice(b[0:8])
	b2 := archsimd.LoadUint32x4Slice(b[8:12])
	s1 := a1.Add(b1)
	s2 := a2.Add(b2)
	s1.StoreSlice(out[0:8])
	s2.StoreSlice(out[8:12])
}

func StatBlock_AddAndSubtract_Into_Experiment(add1, add2, subtract, out *StatBlock) {
	a1 := archsimd.LoadUint32x8Slice(add1[0:8])
	b1 := archsimd.LoadUint32x8Slice(add2[0:8])
	c1 := archsimd.LoadUint32x8Slice(subtract[0:8])
	s1 := a1.Add(b1).Sub(c1)
	s1.StoreSlice(out[0:8])

	a2 := archsimd.LoadUint32x4Slice(add1[8:12])
	b2 := archsimd.LoadUint32x4Slice(add2[8:12])
	c2 := archsimd.LoadUint32x4Slice(subtract[8:12])
	s2 := a2.Add(b2).Sub(c2)
	s2.StoreSlice(out[8:12])
}

func StatBlock_Increment_Mutating_Experiment(mutate *StatBlock, other *StatBlock) {
	m1 := archsimd.LoadUint32x8Slice(mutate[0:8])
	m2 := archsimd.LoadUint32x4Slice(mutate[8:12])
	o1 := archsimd.LoadUint32x8Slice(other[0:8])
	o2 := archsimd.LoadUint32x4Slice(other[8:12])
	s1 := m1.Add(o1)
	s2 := m2.Add(o2)
	s1.StoreSlice(mutate[0:8])
	s2.StoreSlice(mutate[8:12])
}

func StatBlock_Equals_Experiment(a, b *StatBlock) bool {
	a1 := archsimd.LoadUint32x8Slice(a[0:8])
	a2 := archsimd.LoadUint32x4Slice(a[8:12])
	b1 := archsimd.LoadUint32x8Slice(b[0:8])
	b2 := archsimd.LoadUint32x4Slice(b[8:12])

	e1 := a1.Equal(b1)
	e2 := a2.Equal(b2)

	f1 := e1.ToBits()
	f2 := e2.ToBits() | 0xf0
	g := f1 & f2

	return g == 0xff
}

func StatBlock_StatBlockFloat_MultiplyForTotalSum2(a *StatBlockFloat, b *StatBlock) float64 {
	a1 := archsimd.LoadFloat64x4Slice(a[0:4])
	a2 := archsimd.LoadFloat64x4Slice(a[4:8])
	a3 := archsimd.LoadFloat64x4Slice(a[8:12])

	p1 := (*[4]int32)(unsafe.Pointer((*[4]uint32)(b[0:4])))
	p2 := (*[4]int32)(unsafe.Pointer((*[4]uint32)(b[4:8])))
	p3 := (*[4]int32)(unsafe.Pointer((*[4]uint32)(b[8:12])))
	b1 := archsimd.LoadInt32x4(p1).ConvertToFloat64()
	b2 := archsimd.LoadInt32x4(p2).ConvertToFloat64()
	b3 := archsimd.LoadInt32x4(p3).ConvertToFloat64()

	total := a1.Mul(b1)
	total = a2.MulAdd(b2, total)
	total = a3.MulAdd(b3, total)

	parts := [4]float64{}
	total.Store(&parts)

	return parts[0] + parts[1] + parts[2] + parts[3]
}

func StatBlock_StatBlockFloat_MultiplyForTotalSumZZZ(a *StatBlockFloat, b *StatBlock) float64 {
	a1 := archsimd.LoadFloat64x4Slice(a[0:4]).ConvertToFloat32()
	a2 := archsimd.LoadFloat64x4Slice(a[4:8]).ConvertToFloat32()
	a12 := archsimd.Float32x8{}
	a12.SetLo(a1)
	a12.SetHi(a2)
	a3 := archsimd.LoadFloat64x4Slice(a[8:12]).ConvertToFloat32()

	//b1 := archsimd.LoadUint32x8Slice(b[0:8]).ConvertToFloat32()
	//b2 := archsimd.LoadUint32x4Slice(b[8:12]).ConvertToFloat32()
	//b1 := archsimd.LoadUint32x8((*[8]uint32)(b[0:8])).ConvertToFloat32()
	//b2 := archsimd.LoadUint32x4((*[4]uint32)(b[8:12])).ConvertToFloat32()
	ptrB1 := (*[8]uint32)(b[0:8])
	ptrB2 := (*[4]uint32)(b[8:12])
	ptrB1Signed := (*[8]int32)(unsafe.Pointer(ptrB1))
	ptrB2Signed := (*[4]int32)(unsafe.Pointer(ptrB2))
	b1 := archsimd.LoadInt32x8(ptrB1Signed).ConvertToFloat32()
	b2 := archsimd.LoadInt32x4(ptrB2Signed).ConvertToFloat32()

	total1 := a12.Mul(b1)
	total2 := archsimd.Float32x8{}
	total2.SetLo(a3.Mul(b2))

	total := total1.Add(total2)

	parts := [8]float32{}
	total.Store(&parts)

	sum := parts[0] + parts[1] + parts[2] + parts[3] + parts[4] + parts[5] + parts[6] + parts[7]
	return float64(sum)
}
