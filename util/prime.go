package util

import (
	"math/big"
)

var Int_Zero = big.NewInt(0)
var Int_One = big.NewInt(1)

func PrimeNextGreaterOrEqual(skip *big.Int) *big.Int {
	if skip.Cmp(Int_One) < 0 {
		skip.Set(Int_One)
		return skip
	}

	for !skip.ProbablyPrime(100) {
		skip.Add(skip, Int_One)
	}
	return skip
}

func PrimeNextGreater(skip *big.Int) *big.Int {
	if skip.Cmp(Int_One) < 0 {
		skip.Set(Int_One)
		return skip
	}

	skip.Add(skip, Int_One)
	for !skip.ProbablyPrime(100) {
		skip.Add(skip, Int_One)
	}
	return skip
}

func ChooseSkip_NextPrimeFromRatio(actualCombos, targetRunSize *big.Int) *big.Int {
	if actualCombos.Cmp(Int_Zero) == 0 || targetRunSize.Cmp(Int_Zero) == 0 {
		panic("unexpected zero")
	}

	skip := big.NewInt(0)
	if actualCombos.Cmp(targetRunSize) > 0 {
		skip.Div(actualCombos, targetRunSize)
		skip = PrimeNextGreaterOrEqual(skip)
	} else {
		skip = Int_One
	}
	return skip
}

var smallPrimeArray []int = []int{3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103,
	107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227, 229, 233, 239, 241, 251,
	257, 263, 269, 271, 277, 281, 283, 293, 307, 311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409}

// size is to line up with item set size
// could be more optimal but shouldn't be used in inner loops
func SmallPrimes(atLeast int) [16]int {
	result := [16]int{}
	resultIndex := 0
	for _, value := range smallPrimeArray {
		if value >= atLeast {
			result[resultIndex] = value
			resultIndex++
			if resultIndex == 16 {
				break
			}
		}
	}
	if resultIndex < 16 {
		panic("not enough primes")
	}
	return result
}
