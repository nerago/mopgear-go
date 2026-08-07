package util_async

import (
	"iter"
	"paladin_gearing_go/util/util_collection"
	"reflect"
	"sync"
)

func makeOutputChannel[R any]() chan R {
	return make(chan R, 12)
}
func makeOutputChannelUnbuffered[R any]() chan R {
	return make(chan R)
}

func Map_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R) <-chan R {
	outputChannel := makeOutputChannel[R]()
	if threadCount > 1 {
		waitGroup := makeThreadsMapValuesChannelToChannel(threadCount, inputChannel, mapper, outputChannel)
		closeChannelOnGroupFinished(waitGroup, outputChannel)
	} else {
		go func() {
			for value := range inputChannel {
				outputChannel <- mapper(value)
			}
			close(outputChannel)
		}()
	}
	return outputChannel
}

func MapMulti_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R)) <-chan R {
	outputChannel := makeOutputChannel[R]()
	if threadCount > 1 {
		waitGroup := makeThreadsMapMultiChannelToChannel(threadCount, inputChannel, mapper, outputChannel)
		closeChannelOnGroupFinished(waitGroup, outputChannel)
	} else {
		go func() {
			for value := range inputChannel {
				mapper(value, outputChannel)
			}
			close(outputChannel)
		}()
	}
	return outputChannel
}

func MapMulti_ChannelToChannel_Cancellable[T any, R any](threadCount int, inputChannel <-chan T, cancel CancelSignal, mapper func(T, chan<- R)) <-chan R {
	outputChannel := makeOutputChannel[R]()
	waitGroup := new(sync.WaitGroup)
	makeThreadsMapMultiChannelToChannelCancellable(threadCount, inputChannel, cancel, mapper, waitGroup, outputChannel)
	closeChannelOnGroupFinished(waitGroup, outputChannel)
	return outputChannel
}

func MapMulti_SliceToChannel_Cancellable[T any, R any](threadCount int, inputSlice []T, cancel CancelSignal, mapper func(*T, chan<- R)) <-chan R {
	indexChannel := makeIndexChannelCancellable(inputSlice, cancel)
	outputChannel := makeOutputChannel[R]()
	waitGroup := makeThreadsMapMultiSliceToChannelCancellable(threadCount, inputSlice, cancel, mapper, indexChannel, outputChannel)
	closeChannelOnGroupFinished(waitGroup, outputChannel)
	return outputChannel
}

func Map_ChannelToSlice[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R) []R {
	tempChannel := make(chan R)
	waitGroup := makeThreadsMapValuesChannelToChannel(threadCount, inputChannel, mapper, tempChannel)
	closeChannelOnGroupFinished(waitGroup, tempChannel)
	return channelToSlice(tempChannel)
}

func Map_ChannelToSlice_FutureCancellable[T any, R any](threadCount int, inputChannel <-chan T, onComplete func(), mapper func(T) R) *FutureCancellable[[]R] {
	future := FutureCancellable_Make[[]R]()
	tempChannel := make(chan R)
	waitGroup := makeThreadsMapValuesChannelToChannelCancellable(threadCount, inputChannel, future, tempChannel, mapper)
	closeChannelOnGroupFinished(waitGroup, tempChannel)

	go func() {
		outputSlice := channelToSlice(tempChannel)
		future.SetResult(outputSlice)
		onComplete() // TODO maybe onComplete should be a feature of Futures
	}()

	return future
}

func Map_SliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T) R) <-chan R {
	outputChannel := makeOutputChannel[R]()
	if threadCount > 1 {
		indexChannel := makeIndexChannel(inputSlice)
		waitGroup := makeThreadsMapSliceToChannel(threadCount, inputSlice, mapper, indexChannel, outputChannel)
		closeChannelOnGroupFinished(waitGroup, outputChannel)
	} else {
		go func() {
			for i := range inputSlice {
				outputChannel <- mapper(&inputSlice[i])
			}
			close(outputChannel)
		}()
	}
	return outputChannel
}

func MapOptional_SliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T) (R, bool)) <-chan R {
	outputChannel := makeOutputChannel[R]()
	indexChannel := makeIndexChannel(inputSlice)
	waitGroup := makeThreadsMapOptionalSliceToChannel(threadCount, inputSlice, mapper, indexChannel, outputChannel)
	closeChannelOnGroupFinished(waitGroup, outputChannel)
	return outputChannel
}

func MapOptional_SliceToChannel_Cancellable[T any, R any](threadCount int, inputSlice []T, cancel CancelSignal, mapper func(*T) (R, bool)) <-chan R {
	indexChannel := makeIndexChannelCancellable(inputSlice, cancel)
	outputChannel := makeOutputChannel[R]()
	waitGroup := makeThreadsMapOptionalSliceToChannelCancellable(threadCount, inputSlice, cancel, mapper, indexChannel, outputChannel)
	closeChannelOnGroupFinished(waitGroup, outputChannel)
	return outputChannel
}

func Map_SliceToSlice[T any, R any](threadCount int, inputSlice []T, mapper func(*T) R) []R {
	if threadCount > 1 {
		resultChannel := make(chan R, threadCount)
		indexChannel := makeIndexChannel(inputSlice)
		waitGroup := makeThreadsMapSliceToChannel(threadCount, inputSlice, mapper, indexChannel, resultChannel)
		closeChannelOnGroupFinished(waitGroup, resultChannel)
		return channelToSliceKnownSize(resultChannel, len(inputSlice))
	} else {
		return util_collection.MapSliceAsNew(inputSlice, mapper)
	}
}

func Map_SliceToSlice_Cancellable[T any, R any](threadCount int, inputSlice []T, cancel CancelSignal, mapper func(*T) R) []R {
	indexChannel := makeIndexChannelCancellable(inputSlice, cancel)
	tempChannel := make(chan R, threadCount)
	waitGroup := makeThreadsSliceToChannelCancellable(threadCount, inputSlice, cancel, mapper, indexChannel, tempChannel)
	closeChannelOnGroupFinished(waitGroup, tempChannel)
	return channelToSliceKnownSize(tempChannel, len(inputSlice))
}

func ForEach_Slice[T any](threadCount int, inputSlice []T, process func(*T)) {
	if threadCount > 1 {
		indexChannel := makeIndexChannel(inputSlice)
		waitGroup := makeThreadsForEachSlice(threadCount, inputSlice, process, indexChannel)
		waitGroup.Wait()
	} else {
		for i := range inputSlice {
			process(&inputSlice[i])
		}
	}
}

func ForEach_Slice_Cancellable[T any](threadCount int, inputSlice []T, cancel CancelSignal, process func(*T)) {
	indexChannel := makeIndexChannelCancellable(inputSlice, cancel)
	waitGroup := makeThreadsForEachSlice(threadCount, inputSlice, process, indexChannel)
	waitGroup.Wait()
}

func ForEach_Channel[T any](threadCount int, inputChannel <-chan T, process func(T)) {
	waitGroup := makeThreadsForEachChannel(threadCount, inputChannel, process)
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

	waitGroup := new(sync.WaitGroup)
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
	outputChannel := makeOutputChannel[T]()
	go func() {
		for value := range seq {
			outputChannel <- value
		}
		close(outputChannel)
	}()
	return outputChannel
}

func SeqToChannel_Cancellable[T any](seq iter.Seq[T], cancel CancelSignal) <-chan T {
	outputChannel := makeOutputChannelUnbuffered[T]()
	go func() {
	outer:
		for value := range seq {
			select {
			case outputChannel <- value:
			case <-cancel.CancelSignalChannel():
				break outer
			}
		}
		close(outputChannel)
	}()
	return outputChannel
}

func PeekChannel[T any](inputChannel <-chan T, apply func(*T)) <-chan T {
	outputChannel := makeOutputChannel[T]()

	go func() {
		for value := range inputChannel {
			apply(&value)
			outputChannel <- value
		}
		close(outputChannel)
	}()

	return outputChannel
}

func TeeChannelToSlice[T any](inputChannel <-chan T) (<-chan T, *Future[[]T]) {
	future := Future_Make[[]T]()
	outputChannel := makeOutputChannel[T]()

	go func() {
		slice := make([]T, 0)

		for value := range inputChannel {
			slice = append(slice, value)
			outputChannel <- value
		}
		close(outputChannel)

		future.SetResult(slice)
	}()

	return outputChannel, future
}

func Channel_RemoveDuplicatesComparable[T comparable](inputChannel <-chan T) <-chan T {
	outputChannel := makeOutputChannel[T]()
	seen := make(map[T]bool)

	go func() {
		for next := range inputChannel {
			if !seen[next] {
				seen[next] = true
				outputChannel <- next
			}
		}
		close(outputChannel)
	}()

	return outputChannel
}

func Channel_RemoveDuplicatesFunc[T any](inputChannel <-chan T, equals func(a, b *T) bool) <-chan T {
	outputChannel := makeOutputChannel[T]()
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
		close(outputChannel)
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
	outputChannel := makeOutputChannel[T]()

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
		close(outputChannel)
	}()

	return outputChannel
}

func MixChannelsMany[T any](inputChannels []<-chan T) <-chan T {
	outputChannel := makeOutputChannel[T]()

	cases := make([]reflect.SelectCase, len(inputChannels))
	for i := range inputChannels {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(inputChannels[i]),
		}
	}

	go func() {
		for len(cases) > 0 {
			chosen, reflectValue, hasValue := reflect.Select(cases)
			if hasValue {
				value := reflectValue.Interface().(T)
				outputChannel <- value
			} else {
				util_collection.DeleteIndexInPlace(&cases, chosen)
			}
		}
		close(outputChannel)
	}()

	return outputChannel
}

func ChannelWithPrependedValues[T any](inputChannel <-chan T, values ...T) <-chan T {
	outputChannel := makeOutputChannel[T]()

	go func() {
		for _, value := range values {
			outputChannel <- value
		}

		for value := range inputChannel {
			outputChannel <- value
		}

		close(outputChannel)
	}()

	return outputChannel
}

func ChannelCopy[T any](inputChannel <-chan T, outputChannel chan<- T, closeOutputOnDone bool) {
	go func() {
		for value := range inputChannel {
			outputChannel <- value
		}
		if closeOutputOnDone {
			close(outputChannel)
		}
	}()
}

func closeChannelOnGroupFinished[R any](waitGroup *sync.WaitGroup, outputChannel chan R) {
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
}

func channelToSlice[R any](tempChannel chan R) []R {
	outputSlice := make([]R, 0)
	for item := range tempChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func channelToSliceKnownSize[R any](tempChannel chan R, size int) []R {
	outputSlice := make([]R, 0, size)
	for item := range tempChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func makeIndexChannel[T any](inputSlice []T) chan int {
	indexChannel := make(chan int)
	go func() {
		for index := range inputSlice {
			indexChannel <- index
		}
		close(indexChannel)
	}()
	return indexChannel
}

func makeIndexChannelCancellable[T any](inputSlice []T, cancel CancelSignal) chan int {
	indexChannel := make(chan int)
	loopCancelChannel := cancel.CancelSignalChannel()

	go func() {
	indexLoop:
		for index := range inputSlice {
			select {
			case indexChannel <- index:
				// ok
			case <-loopCancelChannel:
				break indexLoop
			}
		}
		close(indexChannel)
	}()
	return indexChannel
}

func makeThreadsMapValuesChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R, outputChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				outputChannel <- mapper(value)
			}
		})
	}
	return waitGroup
}

func makeThreadsMapValuesChannelToChannelCancellable[T any, R any](threadCount int, inputChannel <-chan T, future *FutureCancellable[[]R], tempChannel chan R, mapper func(T) R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				if future.ShouldFinish() {
					break
				}

				tempChannel <- mapper(value)

				if future.ShouldFinish() {
					break
				}
			}
		})
	}
	return waitGroup
}

func makeThreadsMapMultiChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R), outputChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				mapper(value, outputChannel)
			}
		})
	}
	return waitGroup
}

func makeThreadsMapMultiChannelToChannelCancellable[T any, R any](threadCount int, inputChannel <-chan T, cancel CancelSignal, mapper func(T, chan<- R), waitGroup *sync.WaitGroup, outputChannel chan R) {
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				if cancel.ShouldFinish() {
					break
				}

				mapper(value, outputChannel)

				if cancel.ShouldFinish() {
					break
				}
			}
		})
	}
}

func makeThreadsMapSliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T) R, indexChannel chan int, outputChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				outputChannel <- mapper(&inputSlice[index])
			}
		})
	}
	return waitGroup
}

func makeThreadsSliceToChannelCancellable[T any, R any](threadCount int, inputSlice []T, cancel CancelSignal, mapper func(*T) R, indexChannel chan int, tempChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				if cancel.ShouldContinue() {
					tempChannel <- mapper(&inputSlice[index])
				}
			}
		})
	}
	return waitGroup
}

func makeThreadsMapMultiSliceToChannelCancellable[T any, R any](threadCount int, inputSlice []T, cancel CancelSignal, mapper func(*T, chan<- R), indexChannel chan int, outputChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				if cancel.ShouldContinue() {
					mapper(&inputSlice[index], outputChannel)
				}
			}
		})
	}
	return waitGroup
}

func makeThreadsMapOptionalSliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T) (R, bool), indexChannel chan int, outputChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
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
	return waitGroup
}

func makeThreadsMapOptionalSliceToChannelCancellable[T any, R any](threadCount int, inputSlice []T, cancel CancelSignal, mapper func(*T) (R, bool), indexChannel chan int, outputChannel chan R) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				if cancel.ShouldContinue() {
					value, isValid := mapper(&inputSlice[index])
					if isValid {
						outputChannel <- value
					}
				}
			}
		})
	}
	return waitGroup
}

func makeThreadsForEachSlice[T any](threadCount int, inputSlice []T, process func(*T), indexChannel chan int) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				process(&inputSlice[index])
			}
		})
	}
	return waitGroup
}

func makeThreadsForEachChannel[T any](threadCount int, inputChannel <-chan T, process func(T)) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				process(value)
			}
		})
	}
	return waitGroup
}
