package util

func PrimesSmall(requestCount int64) []int64 {
	var maxCheck int64 = 10000
	var prime int64 = 2

	sieve := make([]bool, maxCheck)
	for prime*prime < maxCheck {
		for i := prime * prime; i < maxCheck; i += prime {
			sieve[i] = true
		}

		for i := prime + 1; i < maxCheck; i++ {
			if !sieve[i] {
				prime = i
				break
			}
		}
	}

	primeList := make([]int64, 0, requestCount)
	for prime = 2; prime < maxCheck && int64(len(primeList)) < requestCount; prime++ {
		if !sieve[prime] {
			primeList = append(primeList, prime)
		}
	}
	return primeList
}
