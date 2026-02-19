package util

import (
	"fmt"
	"math/big"
	"time"
)

func PrintProgressInt(startTime time.Time, percent float64, index uint64) {
	if percent > 0 {
		timeTaken := time.Since(startTime)
		totalEstimate := time.Duration(float64(timeTaken) / percent)
		estimateRemain := totalEstimate - timeTaken
		fmt.Printf("%d %.1f%% %s\n", index, percent*100, estimateRemain.String())
	}
}

func PrintProgressBig(startTime time.Time, percent float64, index *big.Int) {
	if percent > 0 {
		timeTaken := time.Since(startTime)
		totalEstimate := time.Duration(float64(timeTaken) / percent)
		estimateRemain := totalEstimate - timeTaken
		fmt.Printf("%d %.1f%% %s\n", index, percent*100, estimateRemain.String())
	}
}
