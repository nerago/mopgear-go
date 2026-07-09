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
	root *trackProgressRoot

	progressForParent  func() float64
	childList          []*TrackProgress
	expectedChildCount int

	active        bool
	cancelHandler func()

	mutex sync.RWMutex
}

func TrackProgress_Start() *TrackProgress {
	track := new(TrackProgress)
	track.root = trackProgressMakeRoot()
	return track
}

func TrackProgress_Nop() *TrackProgress {
	return new(TrackProgress)
}

func (track *TrackProgress) NewChild() *TrackProgress {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	child := new(TrackProgress)
	track.childList = append(track.childList, child)
	return child
}

func oneFunc() float64 {
	return 1.0
}

func (track *TrackProgress) SetDone() {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	if track.root != nil {
		track.root.stopLoop()
	} else {
		track.progressForParent = oneFunc
	}

	track.active = false
}

func (track *TrackProgress) IsActive() bool {
	return track.active
}

func (track *TrackProgress) run(getProgress func() float64) {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	if track.root != nil {
		track.root.startLoop(getProgress)
	} else if track.progressForParent == nil {
		track.progressForParent = getProgress
	} else if track.active {
		panic("TrackProgress run already called")
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
	track.mutex.Lock()
	defer track.mutex.Unlock()

	if expectedChildCount < 0 {
		expectedChildCount = 0
	}
	track.expectedChildCount = expectedChildCount
	track.childList = make([]*TrackProgress, 0, expectedChildCount)
	track.run(track.sumNestedProgress)
}

func (track *TrackProgress) UpdateExpectedChildCount(expectedChildCount int) {
	track.mutex.Lock()
	defer track.mutex.Unlock()

	track.expectedChildCount = expectedChildCount
}

func (track *TrackProgress) sumNestedProgress() float64 {
	track.mutex.RLock()
	defer track.mutex.RUnlock()

	currentCount := track.expectedChildCount
	if currentCount == 0 {
		currentCount = len(track.childList)
	}

	var overallPercent float64 = 0
	for _, nested := range track.childList {
		if nested != nil {
			childFunc := nested.progressForParent
			if childFunc != nil {
				childRaw := childFunc()
				overallPercent += childRaw / float64(currentCount)
			}
		}
	}

	return overallPercent
}

type progressSnapshot struct {
	when    time.Time
	percent float64
}

type trackProgressRoot struct {
	lastPercent float64
	startTime   time.Time
	ringBuffer  RingBufferCache[progressSnapshot]

	activeLoopRunning  bool
	activeLoopEndCheck context.Context
	activeLoopEndNow   context.CancelFunc
}

func trackProgressMakeRoot() *trackProgressRoot {
	root := new(trackProgressRoot)
	root.activeLoopEndCheck, root.activeLoopEndNow = context.WithCancel(context.Background())
	root.startTime = time.Now()
	root.ringBuffer = RingBufferCache_Create(10, progressSnapshot{root.startTime, 0.0})
	return root
}

func (root *trackProgressRoot) startLoop(getProgress func() float64) {
	if root.activeLoopRunning {
		panic("TrackProgress run already called")
	}

	go func() {
		root.activeLoopRunning = true

	loop:
		for {
			select {
			case <-root.activeLoopEndCheck.Done():
				break loop
			case <-time.After(time.Second * 5):
				percent := getProgress()
				root.printProgress(percent)
			}
		}

		root.activeLoopRunning = false
	}()
}

func (root *trackProgressRoot) stopLoop() {
	root.activeLoopEndNow()
}

func (root *trackProgressRoot) printProgress(percent float64) {
	if percent > 0 {
		if percent != root.lastPercent {
			now := time.Now()

			estimateRemain := root.estimateRemain(now, percent)
			root.ringBuffer.Write(progressSnapshot{now, percent})

			fmt.Printf("%5.2f%% %s\n", percent*100, estimateRemain)
			root.lastPercent = percent
		}
	}
}

func (root *trackProgressRoot) estimateRemain(now time.Time, percent float64) string {
	est1 := root.estimateRemainFromRef(now, percent, root.ringBuffer.ReadOldest())
	est2 := root.estimateRemainFromRef(now, percent, root.ringBuffer.ReadNewest())
	est3 := root.estimateRemainFromStart(now, percent)

	averageRemain := (est1 + est2 + est3) / 3
	return compactDurationString(averageRemain)
}

func (root *trackProgressRoot) estimateRemainFromRef(now time.Time, percent float64, ref progressSnapshot) time.Duration {
	timeTakenSinceRef := float64(now.Sub(ref.when))
	percentIncreaseSinceRef := percent - ref.percent
	totalEstimateRef := timeTakenSinceRef / percentIncreaseSinceRef
	estimateRemain := (1 - percent) * totalEstimateRef
	return time.Duration(estimateRemain)
}

func (root *trackProgressRoot) estimateRemainFromStart(now time.Time, percent float64) time.Duration {
	timeTakenSince := float64(now.Sub(root.startTime))
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
