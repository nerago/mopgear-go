package util

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type TrackProgress struct {
	active bool
	nested bool

	lastPercent float64
	startTime   time.Time
	ringBuffer  RingBuffer[progressSnapshot]

	ctx    context.Context
	cancel context.CancelFunc

	nestedChildList    []*TrackProgress
	nestedProgressFunc func() float64

	mutex sync.RWMutex
}

type progressSnapshot struct {
	when    time.Time
	percent float64
}

func TrackProgress_Start() *TrackProgress {
	track := new(TrackProgress)
	track.active = true
	track.ctx, track.cancel = context.WithCancel(context.Background())
	track.startTime = time.Now()
	track.ringBuffer = RingBuffer_Create(10, progressSnapshot{track.startTime, 0.0})
	return track
}

func TrackProgress_Nop() *TrackProgress {
	return new(TrackProgress)
}

func (track *TrackProgress) MakeNested() *TrackProgress {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	nested := new(TrackProgress)
	nested.nested = true
	track.nestedChildList = append(track.nestedChildList, nested)

	return nested
}

func oneFunc() float64 {
	return 1.0
}

func (track *TrackProgress) Stop() {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	if track.active {
		track.active = false
		track.cancel()
	} else if track.nested {
		track.nested = false
		track.nestedProgressFunc = oneFunc
	}

	for _, nested := range track.nestedChildList {
		if nested != nil {
			nested.Stop()
		}
	}
	track.nestedChildList = nil
}

func (track *TrackProgress) IsRunning() bool {
	return track.active || track.nested
}

func (track *TrackProgress) run(getProgress func() float64) {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	if track.active {
		go func() {
			for {
				select {
				case <-track.ctx.Done():
					return
				case <-time.After(time.Second * 5):
					percent := getProgress()
					track.printProgress(percent)
				}
			}
		}()
	} else if track.nested {
		track.nestedProgressFunc = getProgress
	}
}

func (track *TrackProgress) RunFromInt(current *uint64, targetCount uint64) {
	track.run(func() float64 {
		percent := float64(*current) / float64(targetCount)
		return percent
	})
}

func (track *TrackProgress) RunFromAtomicInt(current *atomic.Uint64, targetCount uint64) {
	track.run(func() float64 {
		percent := float64(current.Load()) / float64(targetCount)
		return percent
	})
}

func (track *TrackProgress) RunFromArray(array *[]uint64, targetCount uint64) {
	track.run(func() float64 {
		var current uint64
		for _, value := range *array {
			current += value
		}

		percent := float64(current) / float64(targetCount)
		return percent
	})
}

func (track *TrackProgress) PrepareForPush() func(float64) {
	var latest float64

	track.run(func() float64 {
		return latest
	})

	return func(v float64) {
		latest = v
	}
}

func (track *TrackProgress) RunOuterTracking(expectedChildCount int) {
	track.nestedChildList = make([]*TrackProgress, 0, expectedChildCount)
	track.run(func() float64 {
		return track.sumNestedProgress(expectedChildCount)
	})
}

func (track *TrackProgress) sumNestedProgress(expectedChildCount int) float64 {
	track.mutex.RLock()
	defer track.mutex.RUnlock()

	var overallPercent float64 = 0
	for _, nested := range track.nestedChildList {
		if nested != nil {
			childFunc := nested.nestedProgressFunc
			if childFunc != nil {
				childRaw := childFunc()
				overallPercent += childRaw / float64(expectedChildCount)
			}
		}
	}

	return overallPercent
}

func (track *TrackProgress) printProgress(percent float64) {
	if percent > 0 {
		if percent != track.lastPercent {
			now := time.Now()

			estimateRemain := track.estimateRemain(now, percent)
			track.ringBuffer.Write(progressSnapshot{now, percent})

			fmt.Printf("%5.2f%% %s\n", percent*100, estimateRemain)
			track.lastPercent = percent
		}
	}
}

func (track *TrackProgress) estimateRemain(now time.Time, percent float64) string {
	est1 := track.estimateRemainFromRef(now, percent, track.ringBuffer.ReadOldest())
	est2 := track.estimateRemainFromRef(now, percent, track.ringBuffer.ReadNewest())
	est3 := track.estimateRemainFromStart(now, percent)

	averageRemain := (est1 + est2 + est3) / 3
	return compactDurationString(averageRemain)
}

func (track *TrackProgress) estimateRemainFromRef(now time.Time, percent float64, ref progressSnapshot) time.Duration {
	timeTakenSinceRef := float64(now.Sub(ref.when))
	percentIncreaseSinceRef := percent - ref.percent
	totalEstimateRef := timeTakenSinceRef / percentIncreaseSinceRef
	estimateRemain := (1 - percent) * totalEstimateRef
	return time.Duration(estimateRemain)
}

func (track *TrackProgress) estimateRemainFromStart(now time.Time, percent float64) time.Duration {
	timeTakenSince := float64(now.Sub(track.startTime))
	totalEstimate := timeTakenSince / percent
	estimateRemain := totalEstimate - timeTakenSince
	return time.Duration(estimateRemain)
}

func compactDurationString(duration time.Duration) string {
	if duration < time.Second {
		return "<1s"
	} else if duration < time.Minute {
		wholeSeconds := duration / time.Second

		buff := make([]byte, 0, 3)
		buff = strconv.AppendInt(buff, int64(wholeSeconds), 10)
		buff = append(buff, 's')
		return string(buff)
	} else if duration < time.Hour {
		wholeMinutes := duration / time.Minute
		duration -= wholeMinutes * time.Minute
		wholeSeconds := duration / time.Second

		buff := make([]byte, 0, 8)
		buff = strconv.AppendInt(buff, int64(wholeMinutes), 10)
		buff = append(buff, 'm', ' ')
		buff = strconv.AppendInt(buff, int64(wholeSeconds), 10)
		buff = append(buff, 's')
		return string(buff)
	} else {
		wholeHours := duration / time.Hour
		duration -= wholeHours * time.Hour
		wholeMinutes := duration / time.Minute
		duration -= wholeMinutes * time.Minute
		wholeSeconds := duration / time.Second

		buff := make([]byte, 0, 12)
		buff = strconv.AppendInt(buff, int64(wholeHours), 10)
		buff = append(buff, 'h', ' ')
		buff = strconv.AppendInt(buff, int64(wholeMinutes), 10)
		buff = append(buff, 'm', ' ')
		buff = strconv.AppendInt(buff, int64(wholeSeconds), 10)
		buff = append(buff, 's')
		return string(buff)
	}
}
