package stats

import (
	"math/rand/v2"
	"testing"
)

const checkLoops = 100

func TestIncrementMutating(test *testing.T) {
	for range checkLoops {
		a0 := randStatBlock()
		b0 := randStatBlock()

		a1 := a0
		a2 := a0
		b1 := b0
		b2 := b0

		go_StatBlock_Increment_Mutating(&a1, &b1)
		StatBlock_Increment_Mutating(&a2, &b2)

		assertEquals(test, &a1, &a2, "should be same result")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &b2, "should be unchanged")
	}
}

func BenchmarkIncrementMutatingGo(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	for test.Loop() {
		go_StatBlock_Increment_Mutating(&a, &b)
	}
}

func BenchmarkIncrementMutatingAssem(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	for test.Loop() {
		StatBlock_Increment_Mutating(&a, &b)
	}
}

func TestAddInto(test *testing.T) {
	for range checkLoops {
		a0 := randStatBlock()
		b0 := randStatBlock()

		a1 := a0
		a2 := a0
		b1 := b0
		b2 := b0

		var r1, r2 StatBlock

		go_StatBlock_Add_Into(&a1, &b1, &r1)
		StatBlock_Add_Into(&a2, &b2, &r2)

		assertEquals(test, &r1, &r2, "should be same result")
		assertEquals(test, &a0, &a1, "should be unchanged")
		assertEquals(test, &a0, &a2, "should be unchanged")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &b2, "should be unchanged")
	}
}

func BenchmarkAddIntoGo(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	c := randStatBlockLimited(100)
	for test.Loop() {
		go_StatBlock_Add_Into(&a, &b, &c)
	}
}

func BenchmarkAddIntoAssem(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	c := randStatBlockLimited(100)
	for test.Loop() {
		StatBlock_Add_Into(&a, &b, &c)
	}
}

func TestAddAndSubtractInto(test *testing.T) {
	for range checkLoops {
		a0 := randStatBlock()
		b0 := randStatBlock()
		c0 := randStatBlock()

		a1 := a0
		a2 := a0
		b1 := b0
		b2 := b0
		c1 := c0
		c2 := c0

		var r1, r2 StatBlock

		go_StatBlock_AddAndSubtract_Into(&a1, &b1, &c1, &r1)
		StatBlock_AddAndSubtract_Into(&a2, &b2, &c2, &r2)

		assertEquals(test, &r1, &r2, "should be same result")
		assertEquals(test, &a0, &a1, "should be unchanged")
		assertEquals(test, &a0, &a2, "should be unchanged")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &b2, "should be unchanged")
		assertEquals(test, &c0, &c1, "should be unchanged")
		assertEquals(test, &c0, &c2, "should be unchanged")
	}
}

func BenchmarkAddSubtractIntoGo(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	c := randStatBlockLimited(100)
	d := randStatBlockLimited(100)
	for test.Loop() {
		go_StatBlock_AddAndSubtract_Into(&a, &b, &c, &d)
	}
}

func BenchmarkAddSubtractIntoAssem(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	c := randStatBlockLimited(100)
	d := randStatBlockLimited(100)
	for test.Loop() {
		StatBlock_AddAndSubtract_Into(&a, &b, &c, &d)
	}
}

func TestEquals(test *testing.T) {
	check := func(x, y *StatBlock) {
		assertBoolEqual(test, go_StatBlock_Equals(x, y), StatBlock_Equals(x, y))
	}

	for range checkLoops {
		a0 := randStatBlock()
		b0 := randStatBlock()
		c0 := b0

		a1 := a0
		b1 := b0
		c1 := c0

		check(&a1, &a1)
		check(&a1, &b1)
		check(&a1, &c1)
		check(&b1, &b1)
		check(&b1, &c1)
		check(&c1, &c1)

		assertEquals(test, &a0, &a1, "should be unchanged")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &c0, "should be unchanged")
		assertEquals(test, &b0, &c1, "should be unchanged")
	}
}

func BenchmarkEqualsGo(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	c := a
	var r uint64
	for test.Loop() {
		if go_StatBlock_Equals(&a, &b) {
			r++
		}
		if go_StatBlock_Equals(&a, &c) {
			r++
		}
	}
	resultInt = r
}

func BenchmarkEqualsAssem(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	c := a
	var r uint64
	for test.Loop() {
		if StatBlock_Equals(&a, &b) {
			r++
		}
		if StatBlock_Equals(&a, &c) {
			r++
		}
	}
	resultInt = r
}

func TestMultiplySum0(test *testing.T) {
	a := StatBlock{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 13, go_StatBlock_MultiplyForTotalSum_Float(&a, &b))
	assertFloatEqual(test, 13, StatBlock_MultiplyForTotalSum(&a, &b))
}

func TestMultiplySum1(test *testing.T) {
	a := StatBlock{1, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0}

	assertFloatEqual(test, 19, go_StatBlock_MultiplyForTotalSum_Float(&a, &b))
	assertFloatEqual(test, 19, StatBlock_MultiplyForTotalSum(&a, &b))
}

func TestMultiplySum2(test *testing.T) {
	a := StatBlock{1, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 38, go_StatBlock_MultiplyForTotalSum_Float(&a, &b))
	assertFloatEqual(test, 38, StatBlock_MultiplyForTotalSum(&a, &b))
}

func TestMultiplySum(test *testing.T) {
	for range checkLoops {
		a0 := randStatBlockLimited(0x70000000)
		b0 := randStatBlockLimited(0x70000000)

		a1 := a0
		a2 := a0
		b1 := b0
		b2 := b0

		test.Logf("%s\n%s", a1.CreateStringCSV(), b1.CreateStringCSV())
		assertFloatNearEqual(test, go_StatBlock_MultiplyForTotalSum_Float(&a1, &b1), StatBlock_MultiplyForTotalSum(&a2, &b2))

		assertEquals(test, &a0, &a1, "should be unchanged")
		assertEquals(test, &a0, &a2, "should be unchanged")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &b2, "should be unchanged")
	}
}

func TestMultiplySum0FloatA(test *testing.T) {
	a := StatBlockFloat{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 13, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 13, StatBlock_StatBlockFloat_MultiplyForTotalSum2(&a, &b))
}

func TestMultiplySum1FloatA(test *testing.T) {
	a := StatBlockFloat{1, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0}

	assertFloatEqual(test, 19, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 19, StatBlock_StatBlockFloat_MultiplyForTotalSum2(&a, &b))
}

func TestMultiplySum2FloatA(test *testing.T) {
	a := StatBlockFloat{1, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 38, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 38, StatBlock_StatBlockFloat_MultiplyForTotalSum2(&a, &b))
}

func TestMultiplySum3FloatA(test *testing.T) {
	a := StatBlockFloat{1.1, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 38.3, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 38.3, StatBlock_StatBlockFloat_MultiplyForTotalSum2(&a, &b))
}

func TestMultiplySumFloatA(test *testing.T) {
	for range checkLoops {
		a0 := StatBlockFloat_FromIntStatBlock(randStatBlockLimited(0x70000000), 1)
		b0 := randStatBlockLimited(0x70000000)

		a1 := a0
		a2 := a0
		b1 := b0
		b2 := b0

		test.Logf("%s\n%s", a1.CreateStringCSV(1), b1.CreateStringCSV())
		assertFloatNearEqual(test, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a1, &b1), StatBlock_StatBlockFloat_MultiplyForTotalSum2(&a2, &b2))

		assertEqualsFloatBlock(test, &a0, &a1, "should be unchanged")
		assertEqualsFloatBlock(test, &a0, &a2, "should be unchanged")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &b2, "should be unchanged")
	}
}

func TestMultiplySum0FloatB(test *testing.T) {
	a := StatBlockFloat{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 13, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 13, StatBlock_StatBlockFloat_MultiplyForTotalSum3(&a, &b))
}

func TestMultiplySum1FloatB(test *testing.T) {
	a := StatBlockFloat{1, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0}

	assertFloatEqual(test, 19, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 19, StatBlock_StatBlockFloat_MultiplyForTotalSum3(&a, &b))
}

func TestMultiplySum2FloatB(test *testing.T) {
	a := StatBlockFloat{1, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 38, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 38, StatBlock_StatBlockFloat_MultiplyForTotalSum3(&a, &b))
}

func TestMultiplySum3FloatB(test *testing.T) {
	a := StatBlockFloat{1.1, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0}
	b := StatBlock{3, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 0}

	assertFloatEqual(test, 38.3, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a, &b))
	assertFloatEqual(test, 38.3, StatBlock_StatBlockFloat_MultiplyForTotalSum3(&a, &b))
}

func TestMultiplySumFloatB(test *testing.T) {
	for range checkLoops {
		a0 := StatBlockFloat_FromIntStatBlock(randStatBlockLimited(0x70000000), 1)
		b0 := randStatBlockLimited(0x70000000)

		a1 := a0
		a2 := a0
		b1 := b0
		b2 := b0

		test.Logf("%s\n%s", a1.CreateStringCSV(1), b1.CreateStringCSV())
		assertFloatNearEqual(test, go_StatBlock_StatBlockFloat_MultiplyForTotalSum(&a1, &b1), StatBlock_StatBlockFloat_MultiplyForTotalSum3(&a2, &b2))

		assertEqualsFloatBlock(test, &a0, &a1, "should be unchanged")
		assertEqualsFloatBlock(test, &a0, &a2, "should be unchanged")
		assertEquals(test, &b0, &b1, "should be unchanged")
		assertEquals(test, &b0, &b2, "should be unchanged")
	}
}

var resultFloat float64
var resultInt uint64

func BenchmarkMultiplysGoFloat(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	var t float64
	for test.Loop() {
		t += go_StatBlock_MultiplyForTotalSum_Float(&a, &b)
	}
	resultFloat = t
}

func BenchmarkMultiplysAssemFloat(test *testing.B) {
	a := randStatBlockLimited(100)
	b := randStatBlockLimited(100)
	var t float64
	for test.Loop() {
		t += StatBlock_MultiplyForTotalSum(&a, &b)
	}
	resultFloat = t
}

func assertEquals(test *testing.T, a, b *StatBlock, failMessage string) {
	if *a != *b {
		test.Fatalf("FAIL expect=%s actual=%s message=%s", a.CreateString(), b.CreateString(), failMessage)
	}
}

func assertEqualsFloatBlock(test *testing.T, a, b *StatBlockFloat, failMessage string) {
	if *a != *b {
		test.Fatalf("FAIL expect=%s actual=%s message=%s", a.CreateString(1), b.CreateString(1), failMessage)
	}
}

func assertBoolEqual(test *testing.T, a, b bool) {
	if a != b {
		test.Fatalf("FAIL unequal bools")
	}
}

func assertFloatEqual(test *testing.T, a, b float64) {
	if a != b {
		test.Fatalf("FAIL expect=%f actual=%f", a, b)
	}
}

func assertFloatNearEqual(test *testing.T, a, b float64) {
	diff := a - b
	if diff == 0.0 {
		return
	}

	if a != 0 {
		relative := diff / a
		if relative >= -0.000001 && relative <= 0.000001 {
			return
		}
	}

	test.Fatalf("FAIL expect=%f actual=%f", a, b)
}

func assertIntEqual(test *testing.T, a, b uint64) {
	if a != b {
		test.Fatalf("FAIL expect=%d actual=%d", a, b)
	}
}

func randStatBlock() StatBlock {
	block := StatBlock{}
	for i := range block {
		block[i] = rand.Uint32()
	}
	return block
}

func randStatBlockLimited(max uint32) StatBlock {
	block := StatBlock{}
	for i := range block {
		block[i] = rand.Uint32N(max)
	}
	return block
}

func go_StatBlock_Add_Into(a, b, out *StatBlock) {
	for i := range a {
		out[i] = a[i] + b[i]
	}
}

func go_StatBlock_Increment_Mutating(mutate *StatBlock, other *StatBlock) {
	for i := range mutate {
		mutate[i] += other[i]
	}
}

func go_StatBlock_Equals(a, b *StatBlock) (ret bool) {
	return *a == *b
}

func go_StatBlock_AddAndSubtract_Into(add1, add2, subtract, out *StatBlock) {
	for i := range add1 {
		out[i] = add1[i] + add2[i] - subtract[i]
	}
}

func go_StatBlock_MultiplyForTotalSum_Float(a, b *StatBlock) float64 {
	var result float64 = 0
	for i := range a {
		result += float64(a[i]) * float64(b[i])
	}
	return result
}

func go_StatBlock_StatBlockFloat_MultiplyForTotalSum(a *StatBlockFloat, b *StatBlock) float64 {
	var result float64 = 0
	for i := range a {
		result += a[i] * float64(b[i])
	}
	return result
}
