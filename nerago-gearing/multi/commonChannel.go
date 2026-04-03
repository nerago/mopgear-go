package multi

import (
	"maps"
	"math/big"
	"math/rand"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"sync"
)

const additionalSetEach uint64 = 32
const additionalThreads uint64 = 2

type comboType int8

const (
	comboType_random             comboType = iota
	comboType_overflow           comboType = iota
	comboType_baselineAndFillOut comboType = iota
	comboType_equippedAndFillOut comboType = iota
	comboType_equippedExact      comboType = iota
)

func (comboType comboType) name() string {
	switch comboType {
	case comboType_random:
		return "random"
	case comboType_overflow:
		return "overflow"
	case comboType_baselineAndFillOut:
		return "baseline and fillout"
	case comboType_equippedAndFillOut:
		return "equipped and fillout"
	case comboType_equippedExact:
		return "equipped exact"
	default:
		return "unknown"
	}
}

type commonComboEntry struct {
	Item      *items.FullItem
	Forbidden bool
}

type commonCombo struct {
	entryMap     map[items.ItemId]commonComboEntry
	allowChoices map[items.ItemId]bool
	comboType    comboType
	revised      bool
}

func commonCombo_Make(len int, comboType comboType) commonCombo {
	return commonCombo{
		make(map[items.ItemId]commonComboEntry, len),
		make(map[items.ItemId]bool, len),
		comboType,
		false,
	}
}

func (combo *commonCombo) addItem(itemId items.ItemId, item *items.FullItem) {
	combo.entryMap[itemId] = commonComboEntry{Item: item}
}

func (combo *commonCombo) hasItem(itemId items.ItemId) bool {
	_, has := combo.entryMap[itemId]
	return has
}

func (combo *commonCombo) setAllow(itemId items.ItemId, allow bool) {
	combo.allowChoices[itemId] = allow
	if !allow {
		combo.entryMap[itemId] = commonComboEntry{Forbidden: true}
	}
}

func (combo *commonCombo) getValues(itemId items.ItemId) (bool, bool, *items.FullItem) {
	entry, hasEntry := combo.entryMap[itemId]
	if hasEntry {
		return true, !entry.Forbidden, entry.Item
	} else {
		return false, false, nil
	}
}

func (combo *commonCombo) logString() string {
	build := util.StringBuild2{}
	build.WriteString(combo.comboType.name())
	if combo.revised {
		build.WriteString(" REVISED")
	}
	for itemId, allow := range combo.allowChoices {
		build.WriteRune(' ')
		build.WriteUint32(uint32(itemId))
		if allow {
			build.WriteString("=true")
		} else {
			build.WriteString("=false")
		}
		build.WriteRune(' ')
	}
	return build.String()
}

func (combo *commonCombo) clone() commonCombo {
	return commonCombo{
		maps.Clone(combo.entryMap),
		maps.Clone(combo.allowChoices),
		combo.comboType,
		true,
	}
}

func (job *MultiSetJob) makeCommonChannel(commonOptions commonComboOptions, targetCount uint64, trackProgress *util.TrackProgress) <-chan commonCombo {
	counters := make([]uint64, generateThreadCount+additionalThreads)
	additionalCount := additionalSetEach * additionalThreads

	var eachThreadCount uint64
	if targetCount > additionalCount {
		eachThreadCount = max((targetCount-additionalCount)/generateThreadCount, 1)
	} else {
		eachThreadCount = additionalSetEach
	}

	job.printer.Printf("MAKE COMMON total=%d additional=%d eachThread=%d\n", targetCount, additionalCount, eachThreadCount)

	trackProgress.RunFromArray(&counters, targetCount)

	var waitGroup sync.WaitGroup
	comboChannel := make(chan commonCombo)
	// waitGroup.Go(func() { makeBaselineWorker(job.params, commonOptions, &counters[0], comboChannel) })
	// waitGroup.Go(func() { makeEquippedWorker(job.params, commonOptions, &counters[1], comboChannel) })

	makeRandomThreads(&waitGroup, commonOptions, generateThreadCount/2, eachThreadCount, counters[2:2+generateThreadCount/2], comboChannel)
	makeOverflowThreads(&waitGroup, commonOptions, generateThreadCount/2, eachThreadCount, counters[2+generateThreadCount/2:], comboChannel)

	go func() {
		waitGroup.Wait()
		close(comboChannel)
	}()

	if len(job.specificAllowRates) > 0 {
		return applyAllowRate(job.specificAllowRates, comboChannel)
	} else {
		return comboChannel
	}
}

func makeBaselineWorker(params []MultiSetParam, commonOptions commonComboOptions, doneCounter *uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xBA5E))
	for paramIndex := range params {
		param := &params[paramIndex]
		for range additionalSetEach {
			combo := commonCombo_Make(len(commonOptions)+len(param.baselineResult.FullSet.Items()), comboType_baselineAndFillOut)

			// copy what items are in baseline set
			for item := range param.baselineResult.FullSet.Items().AllItemSeq() {
				combo.addItem(item.ItemId(), item)
			}

			fillOutRemainingOptions(commonOptions, &combo, rng)

			comboChannel <- combo
			*doneCounter++
		}
	}
}

func makeEquippedWorker(params []MultiSetParam, commonOptions commonComboOptions, doneCounter *uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xE819))
	for paramIndex := range params {
		param := &params[paramIndex]
		for range additionalSetEach {
			combo := commonCombo_Make(len(commonOptions)+len(param.exactEquippedGear), comboType_equippedAndFillOut)

			// copy what items are in equipped set
			for item := range param.exactEquippedGear.AllItemSeq() {
				combo.addItem(item.ItemId(), item)
			}

			fillOutRemainingOptions(commonOptions, &combo, rng)

			comboChannel <- combo
			*doneCounter++
		}
	}
}

func fillOutRemainingOptions(commonOptions commonComboOptions, combo *commonCombo, rng *rand.Rand) {
	for itemId, options := range commonOptions {
		if !combo.hasItem(itemId) {
			index := rng.Intn(len(options))
			combo.addItem(itemId, &options[index])
		}
	}
}

func combinationCount(options commonComboOptions) *big.Int {
	valueCount := 0
	total := big.NewInt(1)
	for _, slotArray := range options {
		slotSize := int64(len(slotArray))
		if slotSize > 0 {
			total.Mul(total, big.NewInt(slotSize))
			valueCount++
		}
	}
	if valueCount == 0 {
		panic("empty options")
	}
	return total
}

func applyAllowRate(specificAllowRates map[items.ItemId]float32, comboChannel chan commonCombo) <-chan commonCombo {
	if len(specificAllowRates) == 1 {
		itemId, rate := util.MapFirstEntry(specificAllowRates)
		return channel_op.TransformAll_ChannelToChannel(generateThreadCount, comboChannel,
			func(threadNum int, inChan <-chan commonCombo, outChan chan<- commonCombo) {
				rng := rand.New(rand.NewSource(int64(threadNum)))
				for combo := range inChan {
					applyAllowEntry(itemId, rate, &combo, rng)
					outChan <- combo
				}
			})
	} else {
		return channel_op.TransformAll_ChannelToChannel(generateThreadCount, comboChannel,
			func(threadNum int, inChan <-chan commonCombo, outChan chan<- commonCombo) {
				rng := rand.New(rand.NewSource(int64(threadNum)))
				for combo := range inChan {
					for itemId, rate := range specificAllowRates {
						applyAllowEntry(itemId, rate, &combo, rng)
					}
					outChan <- combo
				}
			})
	}
}

func applyAllowEntry(itemId items.ItemId, rate float32, combo *commonCombo, rng *rand.Rand) {
	allow := rng.Float32() < rate
	combo.setAllow(itemId, allow)
}
