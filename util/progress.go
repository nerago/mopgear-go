package util

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

func TrackProgressInt(index *uint64, max uint64) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		percent := float64(*index) / float64(max)

		PrintProgressInt(startTime, percent, *index)
	}

	// TODO never stops
}

func TrackProgressBig(index, max *big.Int) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		PrintProgressBig(startTime, percent, index)
	}

	// TODO never stops
}

func TrackProgressIntThreaded(threadCounters *[]uint64, targetCount uint64, ctx context.Context) {
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			var totalCount uint64 = 0
			for _, value := range *threadCounters {
				totalCount += value
			}
			percent := float64(totalCount) / float64(targetCount)
			PrintProgressInt(startTime, percent, totalCount)
		}
	}
}

func TrackProgressBigThreaded(threadCounters *[12]uint64, skip, max *big.Int, ctx context.Context) {
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second * 5)
		}

		var totalCount uint64 = 0
		for _, value := range threadCounters {
			totalCount += value
		}
		totalCountBig := big.NewInt(int64(totalCount))
		index := big.NewInt(0).Mul(totalCountBig, skip)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		PrintProgressBig(startTime, percent, index)
	}
}

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
