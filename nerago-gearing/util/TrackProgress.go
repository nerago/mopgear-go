package util

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

type TrackProgress struct {
	active bool
	nested bool

	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc

	nestedChildList    []*TrackProgress
	nestedProgressFunc func() float64
}

func TrackProgress_Start() *TrackProgress {
	track := new(TrackProgress)
	track.active = true
	track.startTime = time.Now()
	track.ctx, track.cancel = context.WithCancel(context.Background())
	return track
}

func TrackProgress_Nop() *TrackProgress {
	return new(TrackProgress)
}

func (track *TrackProgress) MakeNested() *TrackProgress {
	nested := new(TrackProgress)
	nested.nested = true
	track.nestedChildList = append(track.nestedChildList, nested)
	return nested
}

func (track *TrackProgress) Stop() {
	if track.active {
		track.active = false
		track.cancel()
	} else if track.nested {
		track.nested = false
		track.nestedProgressFunc = func() float64 { return 1.0 }
	}
}

func (track *TrackProgress) run(getProgress func() (float64, uint64)) {
	if track.active {
		go func() {
			for {
				select {
				case <-track.ctx.Done():
					return
				case <-time.After(time.Second * 5):
					percent, index := getProgress()
					PrintProgressInt(track.startTime, percent, index)
				}
			}
		}()
	} else if track.nested {
		track.nestedProgressFunc = func() float64 {
			percent, _ := getProgress()
			return percent
		}
	}
}

func (track *TrackProgress) RunFromBigInt(current *big.Int, targetCount *big.Int) {
	if track.active {
		go func() {
			for {
				select {
				case <-track.ctx.Done():
					return
				case <-time.After(time.Second * 5):
					var ratio big.Rat
					ratio.SetFrac(current, targetCount)
					percent, _ := ratio.Float64()
					PrintProgressBig(track.startTime, percent, current)
				}
			}
		}()
	} else if track.nested {
		track.nestedProgressFunc = func() float64 {
			var ratio big.Rat
			ratio.SetFrac(current, targetCount)
			percent, _ := ratio.Float64()
			return percent
		}
	}
}

func (track *TrackProgress) RunFromInt(current *uint64, targetCount uint64) {
	track.run(func() (float64, uint64) {
		percent := float64(*current) / float64(targetCount)
		return percent, *current
	})
}

func (track *TrackProgress) RunFromArray(array *[]uint64, targetCount uint64) {
	track.run(func() (float64, uint64) {
		var current uint64
		for _, value := range *array {
			current += value
		}

		percent := float64(current) / float64(targetCount)
		return percent, current
	})
}

func (track *TrackProgress) RunOuterTracking(expectedChildCount int) {
	track.nestedChildList = make([]*TrackProgress, 0, expectedChildCount)
	go func() {
		for {
			select {
			case <-track.ctx.Done():
				return
			case <-time.After(time.Second * 5):
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
				PrintProgressBasic(track.startTime, overallPercent)
			}
		}
	}()
}

func PrintProgressBasic(startTime time.Time, percent float64) {
	if percent > 0 {
		estimateRemain := estimateRemain(startTime, percent)
		fmt.Printf("%5.2f%% %s\n", percent*100, estimateRemain)
	}
}

func PrintProgressInt(startTime time.Time, percent float64, index uint64) {
	if percent > 0 {
		estimateRemain := estimateRemain(startTime, percent)
		fmt.Printf("%d %4.1f%% %s\n", index, percent*100, estimateRemain)
	}
}

func PrintProgressBig(startTime time.Time, percent float64, index *big.Int) {
	if percent > 0 {
		estimateRemain := estimateRemain(startTime, percent)
		fmt.Printf("%d %4.1f%% %s\n", index, percent*100, estimateRemain)
	}
}

func estimateRemain(startTime time.Time, percent float64) string {
	timeTaken := time.Since(startTime)
	totalEstimate := time.Duration(float64(timeTaken) / percent)
	estimateRemain := totalEstimate - timeTaken
	return compactDurationString(estimateRemain)
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
