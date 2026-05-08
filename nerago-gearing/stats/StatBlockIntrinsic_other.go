//go:build !amd64
// +build !amd64

package stats

func StatBlock_Add_Into(a, b, out *StatBlock) {
	for i := range a {
		out[i] = a[i] + b[i]
	}
}

func StatBlock_Increment_Mutating(mutate *StatBlock, other *StatBlock) {
	for i := range mutate {
		mutate[i] += other[i]
	}
}

func StatBlock_Equals(a, b *StatBlock) (ret bool) {
	return *a == *b
}

func StatBlock_AddAndSubtract_Into(add1, add2, subtract, out *StatBlock) {
	for i := range add1 {
		out[i] = add1[i] + add2[i] - subtract[i]
	}
}

func StatBlock_MultiplyForTotalSum_Float(a, b *StatBlock) float32 {
	var result float32 = 0
	for i := range a {
		result += float32(a[i]) * float32(b[i])
	}
	return result
}

func StatBlock_MultiplyForTotalSum_Int(a, b *StatBlock) uint64 {
	var result uint64 = 0
	for i := range a {
		result += uint64(a[i]) * uint64(b[i])
	}
	return result
}
