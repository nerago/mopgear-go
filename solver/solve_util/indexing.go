package solve_util

import (
	"math/big"
)

func IndexSplitsBig(max, skip *big.Int, threadCount int64) []*big.Int {
	indexPerThread := big.NewInt(0)
	indexPerThread.Div(max, skip)
	indexPerThread.Div(indexPerThread, big.NewInt(threadCount))
	indexPerThread.Mul(indexPerThread, skip)

	splitArray := make([]*big.Int, 0)
	start := big.NewInt(0)
	for range threadCount {
		splitArray = append(splitArray, start)
		start = big.NewInt(0).Add(start, indexPerThread)
	}
	splitArray = append(splitArray, max)

	return splitArray
}

func IndexSplitsInt(max, skip uint64, threadCount uint64) []uint64 {
	indexPerThread := max / skip
	indexPerThread /= threadCount
	indexPerThread *= skip

	splitArray := make([]uint64, 0)
	var start uint64 = 0
	for range threadCount {
		splitArray = append(splitArray, start)
		start += indexPerThread
	}
	splitArray = append(splitArray, max)

	return splitArray
}
