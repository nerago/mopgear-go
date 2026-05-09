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

func StatBlock_MultiplyForTotalSum(a, b *StatBlock) float64 {
	// doesn't exactly match order of operations in assembly version anymore, but shouldn't mind floating precision anyway
	var result float64 = 0
	for i := range a {
		result += float64(a[i]) * float64(b[i])
	}
	return result
}
