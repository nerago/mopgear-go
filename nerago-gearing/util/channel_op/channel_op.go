package channel_op

import (
	"iter"
	"sync"
)

func Map_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := make(chan R)

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				mapper(value, outputChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Map_ChannelToChannel_Provided[T any, R any](threadCount int, inputChannel <-chan T, outputChannel chan<- R, mapper func(T, chan<- R)) {
	var waitGroup sync.WaitGroup

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				mapper(value, outputChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
}

func Map_ChannelToSlice[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R)) []R {
	var waitGroup sync.WaitGroup
	tempChannel := make(chan R)

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				mapper(value, tempChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(tempChannel)
	}()

	outputSlice := make([]R, 0)
	for item := range tempChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func Map_SliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	outputChannel := make(chan R)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				mapper(&inputSlice[index], outputChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Map_SliceToChannel_Provided[T any, R any](threadCount int, inputSlice []T, outputChannel chan<- R, mapper func(*T, chan<- R)) {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				mapper(&inputSlice[index], outputChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
}

func Map_SliceToSlice_LowOverhead[T any, R any](threadCount int, inputSlice []T, mapper func(*T, chan<- R)) []R {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	tempChannel := make(chan R)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				mapper(&inputSlice[index], tempChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(tempChannel)
	}()

	outputSlice := make([]R, 0, inputLength)
	for item := range tempChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func Map_SliceToSlice[T any, R any](threadCount int, inputSlice []T, mapper func(*T, chan<- R)) []R {
	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	var waitGroupIndexes sync.WaitGroup
	indexChannel := make(chan int, 8)
	for threadNum := range threadCount {
		waitGroupIndexes.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				indexChannel <- index
			}
		})
	}
	go func() {
		waitGroupIndexes.Wait()
		close(indexChannel)
	}()

	var waitGroupResult sync.WaitGroup
	resultChannel := make(chan R, 8)
	for range threadCount {
		waitGroupResult.Go(func() {
			for index := range indexChannel {
				mapper(&inputSlice[index], resultChannel)
			}
		})
	}
	go func() {
		waitGroupResult.Wait()
		close(resultChannel)
	}()

	outputSlice := make([]R, 0, inputLength)
	for item := range resultChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func ForEach_Slice[T any](threadCount int, inputSlice []T, process func(*T)) {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				process(&inputSlice[index])
			}
		})
	}

	waitGroup.Wait()
}

func ForEach_Channel[T any](threadCount int, inputChannel <-chan T, process func(T)) {
	var waitGroup sync.WaitGroup

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				process(value)
			}
		})
	}

	waitGroup.Wait()
}

func indexSplitsInt(sliceLength int, threadCount int) []int {
	indexPerThread := sliceLength / threadCount

	splitArray := make([]int, 0, threadCount+1)
	start := 0
	for range threadCount {
		splitArray = append(splitArray, start)
		start += indexPerThread
	}
	splitArray = append(splitArray, sliceLength)

	return splitArray
}

type GroupChannelEntry[T any, G comparable] struct {
	groupKey G
	channel  chan T
}

func (entry GroupChannelEntry[T, G]) GroupKey() G {
	return entry.groupKey
}

func (entry GroupChannelEntry[T, G]) Channel() <-chan T {
	return entry.channel
}

func GroupChannel_To_ManyChannel[T any, G comparable](threadCount int, inputChannel <-chan T, toGroup func(T) G) <-chan GroupChannelEntry[T, G] {
	nestedChannelMap := sync.Map{}

	var waitGroup sync.WaitGroup
	outputChannel := make(chan GroupChannelEntry[T, G])

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				groupKey := toGroup(value)

				existing, loadedExisting := nestedChannelMap.Load(groupKey)
				if loadedExisting {
					entry := existing.(GroupChannelEntry[T, G])
					entry.channel <- value
				} else {
					possibleNewEntry := &GroupChannelEntry[T, G]{
						groupKey,
						make(chan T),
					}
					oldOrNew, loadedExisting := nestedChannelMap.LoadOrStore(groupKey, possibleNewEntry)
					if loadedExisting {
						close(possibleNewEntry.channel)
					}

					entry := oldOrNew.(GroupChannelEntry[T, G])
					entry.channel <- value
				}
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
		nestedChannelMap.Range(func(key, value any) bool {
			entry := value.(GroupChannelEntry[T, G])
			close(entry.channel)
			return true
		})
	}()
	return outputChannel
}

func SeqToChannel[T any](seq iter.Seq[T]) <-chan T {
	outputChannel := make(chan T)
	go func() {
		for value := range seq {
			outputChannel <- value
		}
		close(outputChannel)
	}()
	return outputChannel
}

func PeekChannel[T any](threadCount int, inputChannel <-chan T, apply func(T)) <-chan T {
	var waitGroup sync.WaitGroup
	outputChannel := make(chan T)

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				apply(value)
				outputChannel <- value
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func TeeChannelToSlice[T any](inputChannel <-chan T, slicePointer *[]T) <-chan T {
	outputChannel := make(chan T)

	go func() {
		for value := range inputChannel {
			*slicePointer = append(*slicePointer, value)
			outputChannel <- value
		}
		close(outputChannel)
	}()

	return outputChannel
}
