package util

func PrimesSmall(requestCount int64) []int64 {
	var max int64 = 10000
	var prime int64 = 2

	sieve := make([]bool, max)
	for prime*prime < max {
		for i := prime * prime; i < max; i += prime {
			sieve[i] = true
		}

		for i := prime + 1; i < max; i++ {
			if !sieve[i] {
				prime = i
				break
			}
		}
	}

	primeList := make([]int64, requestCount)
	for prime = 2; prime < max && int64(len(primeList)) < requestCount; prime++ {
		if !sieve[prime] {
			primeList = append(primeList, prime)
		}
	}
	return primeList
}
