package channel_op

import (
	"iter"
	"sync"
)

func Map_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R) <-chan R {
	outputChannel := make(chan R)
	var waitGroup sync.WaitGroup

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				outputChannel <- mapper(value)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func MapMulti_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R)) <-chan R {
	outputChannel := make(chan R)
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
	return outputChannel
}

func Map_ChannelToSlice[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R) []R {
	tempChannel := make(chan R)
	var waitGroup sync.WaitGroup

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				tempChannel <- mapper(value)
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

func Map_ChannelToSlice_FutureCancellable[T any, R any](threadCount int, inputChannel <-chan T, isCancelled func() bool, mapper func(T) R) *FutureCancellable[[]R] {
	future := FutureCancellable_Make[[]R](func() {})
	tempChannel := make(chan R)
	var waitGroup sync.WaitGroup

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				if isCancelled() {
					break
				}

				tempChannel <- mapper(value)

				if isCancelled() {
					break
				}
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(tempChannel)
	}()

	go func() {
		outputSlice := make([]R, 0)
		for item := range tempChannel {
			outputSlice = append(outputSlice, item)
		}
		future.SetResult(outputSlice)
	}()

	return future
}

func Map_SliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T) R) <-chan R {
	outputChannel := make(chan R)
	indexChannel := make(chan int, threadCount)
	var waitGroup sync.WaitGroup

	go func() {
		for index := range inputSlice {
			indexChannel <- index
		}
		close(indexChannel)
	}()

	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				outputChannel <- mapper(&inputSlice[index])
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func MapOptional_SliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T) (R, bool)) <-chan R {
	outputChannel := make(chan R)
	indexChannel := make(chan int, threadCount)
	var waitGroup sync.WaitGroup

	go func() {
		for index := range inputSlice {
			indexChannel <- index
		}
		close(indexChannel)
	}()

	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				value, isValid := mapper(&inputSlice[index])
				if isValid {
					outputChannel <- value
				}
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Map_SliceToSlice[T any, R any](threadCount int, inputSlice []T, mapper func(*T) R) []R {
	indexChannel := make(chan int, threadCount)
	resultChannel := make(chan R, threadCount)
	var waitGroup sync.WaitGroup

	go func() {
		for index := range inputSlice {
			indexChannel <- index
		}
		close(indexChannel)
	}()

	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				resultChannel <- mapper(&inputSlice[index])
			}
		})
	}
	go func() {
		waitGroup.Wait()
		close(resultChannel)
	}()

	outputSlice := make([]R, 0, len(inputSlice))
	for item := range resultChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func ForEach_Slice[T any](threadCount int, inputSlice []T, process func(*T)) {
	indexChannel := make(chan int, threadCount)
	var waitGroup sync.WaitGroup

	go func() {
		for index := range inputSlice {
			indexChannel <- index
		}
		close(indexChannel)
	}()

	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
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

func GroupChannel_To_ManyChannel[T any, G comparable](threadCount int, bufferSizes int, inputChannel <-chan T, toGroup func(T) G) <-chan GroupChannelEntry[T, G] {
	nestedChannelMap := sync.Map{}

	var waitGroup sync.WaitGroup
	outputChannel := make(chan GroupChannelEntry[T, G], bufferSizes)

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
						make(chan T, bufferSizes),
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

func PeekChannel[T any](inputChannel <-chan T, apply func(*T)) <-chan T {
	outputChannel := make(chan T)

	go func() {
		for value := range inputChannel {
			apply(&value)
			outputChannel <- value
		}
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

func Channel_RemoveDuplicatesComparable[T comparable](inputChannel <-chan T) <-chan T {
	outputChannel := make(chan T)
	seen := make(map[T]bool)

	go func() {
		for next := range inputChannel {
			if !seen[next] {
				seen[next] = true
				outputChannel <- next
			}
		}
	}()

	return outputChannel
}

func Channel_RemoveDuplicatesFunc[T any](inputChannel <-chan T, equals func(a, b *T) bool) <-chan T {
	outputChannel := make(chan T)
	seen := make([]T, 0)

	go func() {
		for next := range inputChannel {
			found := false
			for checkIndex := range seen {
				if equals(&next, &seen[checkIndex]) {
					found = true
					break
				}
			}

			if !found {
				seen = append(seen, next)
				outputChannel <- next
			}
		}
	}()

	return outputChannel
}

func Channel_RemoveDuplicatesFuncNotify[T any](inputChannel <-chan T, equals func(a, b *T) bool, removedNotify func(x *T)) <-chan T {
	lock := sync.Mutex{}
	seen := make([]T, 0)

	return MapMulti_ChannelToChannel(4, inputChannel, func(next T, outputChannel chan<- T) {
		lock.Lock()

		found := false
		for checkIndex := range seen {
			if equals(&next, &seen[checkIndex]) {
				found = true
				break
			}
		}

		if !found {
			seen = append(seen, next)
			outputChannel <- next
			lock.Unlock()
		} else {
			lock.Unlock() // release lock early so notifies can take longer

			removedNotify(&next)
		}
	})
}

func MixChannels[T any](channelOne <-chan T, channelTwo <-chan T) <-chan T {
	outputChannel := make(chan T)

	go func() {
		for channelOne != nil || channelTwo != nil {
			select {
			case value, ok := <-channelOne:
				if ok {
					outputChannel <- value
				} else {
					channelOne = nil
				}
			case value, ok := <-channelTwo:
				if ok {
					outputChannel <- value
				} else {
					channelTwo = nil
				}
			}
		}
	}()

	return outputChannel
}

func ChannelWithPrependedValues[T any](inputChannel <-chan T, values ...T) <-chan T {
	outputChannel := make(chan T)

	go func() {
		for _, value := range values {
			outputChannel <- value
		}

		for value := range inputChannel {
			outputChannel <- value
		}
	}()

	return outputChannel
}

func ChannelCopy[T any](inputChannel <-chan T, outputChannel chan<- T) {
	go func() {
		for value := range inputChannel {
			outputChannel <- value
		}
	}()
}
