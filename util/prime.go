package util

import (
	"math/big"
)

var int_one = big.NewInt(1)

func NextPrime(skip *big.Int) *big.Int {
	if skip.Cmp(int_one) <= 0 {
		return int_one
	}

	for !skip.ProbablyPrime(100) {
		skip.Add(skip, int_one)
	}
	return skip
}

var smallPrimeArray []int = []int{3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103,
	107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227, 229, 233, 239, 241, 251,
	257, 263, 269, 271, 277, 281, 283, 293, 307, 311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409}

// size is to line up with item set size
// could be more optimal but shouldn't be used in inner loops
func SmallPrimes(greaterThan int) [16]int {
	result := [16]int{}
	resultIndex := 0
	for _, value := range smallPrimeArray {
		if value > greaterThan {
			result[resultIndex] = value
			if resultIndex == 15 {
				break
			}
			resultIndex++
		}
	}
	return result
}
