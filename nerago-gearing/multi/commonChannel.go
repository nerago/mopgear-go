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

const (
	additionalSetEach     uint64 = 32
	additionalThreads     uint64 = 2
	defaultThreadSetCount        = 8
)

type comboType int8

const (
	comboType_random             comboType = iota
	comboType_overflow           comboType = iota
	comboType_baselineAndFillOut comboType = iota
	comboType_equippedAndFillOut comboType = iota
	comboType_equippedExact      comboType = iota
	comboType_highs      comboType = iota
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

type forceItemMode int8

const (
	Force_Unknown             forceItemMode = iota
	Force_Optional            forceItemMode = iota
	Force_Forbidden           forceItemMode = iota
	Force_FixedWhereAvailable forceItemMode = iota
	Force_RequiredAlways      forceItemMode = iota
)

type commonComboEntry struct {
	Item      *items.FullItem
	forceMode forceItemMode
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

func (combo *commonCombo) addItem(itemId items.ItemId, item *items.FullItem, forceMode forceItemMode) {
	combo.entryMap[itemId] = commonComboEntry{Item: item, forceMode: forceMode}
}

func (combo *commonCombo) hasItem(itemId items.ItemId) bool {
	_, has := combo.entryMap[itemId]
	return has
}

func (combo *commonCombo) setAllow(itemId items.ItemId, choiceAlternate bool, force forceItemMode) {
	combo.allowChoices[itemId] = choiceAlternate
	if existing, inMap := combo.entryMap[itemId]; inMap {
		existing.forceMode = force
		combo.entryMap[itemId] = existing
	} else if force != Force_Forbidden {
		panic("trying to set allow/force for item not yet included in combo")
	}
}

func (combo *commonCombo) getAnySpecifications(itemId items.ItemId) (forceMode forceItemMode, item *items.FullItem) {
	entry, hasEntry := combo.entryMap[itemId]
	if hasEntry {
		if entry.forceMode == Force_Unknown {
			panic("force not set")
		}
		return entry.forceMode, entry.Item
	} else {
		return Force_Unknown, nil
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

func (job *MultiSetJob) makeCommonChannel(commonOptions CommonComboOptions, targetCount uint64) (<-chan commonCombo, uint64) {
	additionalCount := additionalSetEach * additionalThreads * uint64(len(job.params))

	var calculatedEachThreadCount int64 = (int64(targetCount) - int64(additionalCount)) / generateThreadCount
	var eachThreadCount uint64
	if calculatedEachThreadCount > defaultThreadSetCount {
		eachThreadCount = uint64(calculatedEachThreadCount)
	} else {
		eachThreadCount = defaultThreadSetCount
	}
	actualExpectedCount := eachThreadCount*((generateThreadCount/2)*2) + additionalCount

	job.printer.Printf("MAKE COMMON total=%d additional=%d eachThread=%d\n", targetCount, additionalCount, eachThreadCount)

	var waitGroup sync.WaitGroup
	comboChannel := make(chan commonCombo)
	waitGroup.Go(func() { makeBaselineWorker(job.params, commonOptions, comboChannel) })
	waitGroup.Go(func() { makeEquippedWorker(job.params, commonOptions, comboChannel) })

	makeRandomThreads(&waitGroup, commonOptions, generateThreadCount/2, eachThreadCount, comboChannel)
	makeOverflowThreads(&waitGroup, commonOptions, generateThreadCount/2, eachThreadCount, comboChannel)

	go func() {
		waitGroup.Wait()
		close(comboChannel)
	}()

	var resultChannel <-chan commonCombo = comboChannel
	if len(job.specificAllowRates) > 0 {
		resultChannel = applyAllowRate(job.specificAllowRates, comboChannel)
	}
	return resultChannel, actualExpectedCount
}

func makeBaselineWorker(params []multiSetParamInternal, commonOptions CommonComboOptions, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xBA5E))
	for paramIndex := range params {
		param := &params[paramIndex]
		for range additionalSetEach {
			combo := commonCombo_Make(len(commonOptions)+len(param.baselineResult.FullSet.Items()), comboType_baselineAndFillOut)

			// copy what items are in baseline set
			for item := range param.baselineResult.FullSet.Items().AllItemSeq() {
				combo.addItem(item.ItemId(), item, Force_Optional)
			}

			fillOutRemainingOptions(commonOptions, &combo, rng)

			comboChannel <- combo
		}
	}
}

func makeEquippedWorker(params []multiSetParamInternal, commonOptions CommonComboOptions, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xE819))
	for paramIndex := range params {
		param := &params[paramIndex]
		for range additionalSetEach {
			combo := commonCombo_Make(len(commonOptions)+len(param.exactEquippedGear), comboType_equippedAndFillOut)

			// copy what items are in equipped set
			for item := range param.exactEquippedGear.AllItemSeq() {
				combo.addItem(item.ItemId(), item, Force_Optional)
			}

			fillOutRemainingOptions(commonOptions, &combo, rng)

			comboChannel <- combo
		}
	}
}

func fillOutRemainingOptions(commonOptions CommonComboOptions, combo *commonCombo, rng *rand.Rand) {
	for itemId, options := range commonOptions {
		if !combo.hasItem(itemId) {
			index := rng.Intn(len(options))
			combo.addItem(itemId, &options[index], Force_Optional)
		}
	}
}

func combinationCount(options CommonComboOptions) *big.Int {
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

func applyAllowRate(specificAllowRates map[items.ItemId]specificAllowEntry, comboChannel chan commonCombo) <-chan commonCombo {
	// TODO consider special handling if it came from an equipped set etc
	if len(specificAllowRates) == 1 {
		itemId, entry := util.MapFirstEntry(specificAllowRates)
		return channel_op.TransformAll_ChannelToChannel(generateThreadCount, comboChannel,
			func(threadNum int, inChan <-chan commonCombo, outChan chan<- commonCombo) {
				rng := rand.New(rand.NewSource(int64(threadNum)))
				for combo := range inChan {
					applyAllowEntry(itemId, entry, &combo, rng)
					outChan <- combo
				}
			})
	} else {
		return channel_op.TransformAll_ChannelToChannel(generateThreadCount, comboChannel,
			func(threadNum int, inChan <-chan commonCombo, outChan chan<- commonCombo) {
				rng := rand.New(rand.NewSource(int64(threadNum)))
				for combo := range inChan {
					for itemId, entry := range specificAllowRates {
						applyAllowEntry(itemId, entry, &combo, rng)
					}
					outChan <- combo
				}
			})
	}
}

func applyAllowEntry(itemId items.ItemId, entry specificAllowEntry, combo *commonCombo, rng *rand.Rand) {
	choiceAlternate := rng.Float32() < entry.proportion
	if choiceAlternate {
		combo.setAllow(itemId, true, entry.modeOn)
	} else {
		combo.setAllow(itemId, false, entry.modeOff)
	}
}
