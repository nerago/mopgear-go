package weightfind

import (
	"math"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/items"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/solve_build"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
)

const (
	grid_sim_max_steps = 6
	grid_sim_step      = 500
)

type incrementStat struct {
	stat  stats.StatType
	value int32
}

type incrementStatCombo map[stats.StatType]int32

func SimulateSteppedStatChangesForGrid(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, tracker *util.TrackProgress, label string, cancel util_async.CancelSignal, fixStatsMode weight_types.FixStatsRangeMode) []weight_types.WeightInput {
	var incrementStep int32 = grid_sim_step
	var incrementMax int32 = incrementStep * grid_sim_max_steps
	if len(requiredStats) == 8 {
		incrementMax = incrementStep * 2
	}

	initialBaseStats := InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementMax, fixStatsMode, true)

	incrementPermutations := makePermutationsForGridSim2(requiredStats, incrementStep, incrementMax, printer)
	tracker.RunOuterTracking(len(incrementPermutations))
	defer tracker.SetDone()

	printer.Printf("Running %d sims (part 1) for %s\n", len(incrementPermutations), label)
	inputList := util_async.Map_SliceToSlice_Cancellable(6, incrementPermutations, cancel, func(increments *incrementStatCombo) weight_types.WeightInput {
		bonusStat := initialBaseStats.Clone()
		str := util.StringBuild2{}
		str.WriteString("SIM ")
		str.WriteString(label)
		str.WriteRune(' ')
		for statType, valueIncrease := range *increments {
			bonusStat.Put(statType, bonusStat.GetOrNilValue(statType)+valueIncrease)

			str.WriteString(statType.Name())
			str.WriteRune('=')
			str.WriteInt32(bonusStat.GetOrNilValue(statType))
			str.WriteRune(' ')
		}

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), bonusStat, tracker.NewChild())

		str.WriteString("--> ")
		simResult.CompactStringGeneralBuilder(&str)
		printer.PrintlnFromBuild(str)

		return weight_types.WeightInput{
			TotalStat: addBonusStats(currentItemSet.Total(), bonusStat),
			SimResult: simResult,
		}
	})
	printer.Printf("Done sims (part 1) for %s\n", label)
	return inputList
}

func makePermutationsForGridSim2(statList []stats.StatType, incrementStep int32, incrementMax int32, printer *util.PrintRecorder) []incrementStatCombo {
	allCombos := make([]incrementStatCombo, 0)

	var incrementLo int32 = 0
	allCombos = append(allCombos, makeWithAllSameValue(statList, incrementLo))

	var incrementHi int32 = incrementStep
	for len(allCombos) < c_eachSimTargetGenerateDataCount && incrementHi <= incrementMax {
		// add the high version first so make sure it's more likely to make it when we cut off the list
		allCombos = append(allCombos, makeWithAllSameValue(statList, incrementHi))

		// add all exact mixes for the two levels
		levelCombos := makeAllMixesOf(statList, incrementLo, incrementHi)
		allCombos = append(allCombos, levelCombos...)

		incrementLo += incrementStep
		incrementHi += incrementStep
	}

	if len(allCombos) > c_eachSimTargetGenerateDataCount {
		allCombos = allCombos[0:c_eachSimTargetGenerateDataCount]
	}

	return allCombos
}

func makeAllMixesOf(statList []stats.StatType, valueOne int32, valueTwo int32) []incrementStatCombo {
	slotCount := uint64(len(statList))
	itemCount := 1 << slotCount
	levelCombos := make([]incrementStatCombo, 0, itemCount-2)

	// don't include the 00000 or 11111 entries since we want to add them specially
	for index := 1; index < itemCount-1; index++ {
		combo := make(incrementStatCombo)
		for statNum, statType := range statList {
			bitMask := 1 << statNum
			if (index & bitMask) == 0 {
				combo[statType] = valueOne
			} else {
				combo[statType] = valueTwo
			}
		}
		levelCombos = append(levelCombos, combo)
	}

	// shuffle so we get a good mix
	util_collection.Shuffle(levelCombos)
	return levelCombos
}

func makeWithAllSameValue(statList []stats.StatType, value int32) incrementStatCombo {
	combo := make(incrementStatCombo)
	for _, stat := range statList {
		combo[stat] = value
	}
	return combo
}

func GenerateRandomSets(gearFile string, substituteItems []items.ItemId, model *gear_model.SpecModel, makeSetCount int, printer *util.PrintRecorder, label string, includeBestWorst bool) ([]items.FullItemSet, items.FullOptionsMap) {
	itemOptions := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItems {
		// TODO support for random suffix items
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, model, printer)
		for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
			if itemOptions.Has(slotEquip) {
				itemOptions.AddSeveralOptionsSpecific(slotEquip, opts)
			}
		}
	}
	itemOptions.RemoveDuplicates()

	setList := solve_build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, model, makeSetCount, label)
	if includeBestWorst {
		setList = append(setList, solve_build.SolverBuildBestWorst(&itemOptions, model)...)
	}
	return setList, itemOptions
}

func SimulateRealRandomSets(gearFile string, substituteItems []items.ItemId, model *gear_model.SpecModel, makeSetCount int, simSize simulate.WowSim_RunSize, fixStatsMode weight_types.FixStatsRangeMode, printer *util.PrintRecorder, track *util.TrackProgress, label string, cancel util_async.CancelSignal) []weight_types.WeightInput {
	setList, _ := GenerateRandomSets(gearFile, substituteItems, model, makeSetCount, printer, label, true)

	track.RunOuterTracking(len(setList))
	defer track.SetDone()

	printer.Printf("Running %d sims (part 2) for %s\n", len(setList), label)
	weightInputs := util_async.Map_SliceToSlice_Cancellable(6, setList, cancel, func(itemSet *items.FullItemSet) weight_types.WeightInput {
		bonusStats := InitialBonusStatMap_fixRanges(printer, *itemSet, 0, fixStatsMode, false)

		var total stats.StatBlock
		if bonusStats != nil {
			total = addBonusStats(itemSet.Total(), bonusStats)
		} else {
			total = *itemSet.Total()
		}

		simResult := simulate.WowSim_Execute_UseModel(simSize, model, itemSet.Items(), bonusStats, track.NewChild())

		str := util.StringBuild2{}
		str.WriteString("SIM ")
		str.WriteString(label)
		str.WriteRune(' ')
		total.AppendString(&str)
		str.WriteString(" --> ")
		simResult.CompactStringGeneralBuilder(&str)
		printer.PrintlnFromBuild(str)

		return weight_types.WeightInput{TotalStat: total, SimResult: simResult}
	})

	printer.Printf("Done sims (part 2) for %s\n", label)
	return weightInputs
}

func addBonusStats(base *stats.StatBlock, bonusStat *stats.StatTypeMap[int32]) stats.StatBlock {
	resultBlock := *base
	for stat, add := range bonusStat.SeqKeyValue() {
		value := int64(resultBlock[stat]) + int64(add)
		if value < 0 || value > math.MaxUint32 {
			panic("out of range")
		}
		resultBlock[stat] = uint32(value)
	}
	return resultBlock
}

const c_hasteDiscontinuityStart = 10500
const c_hasteDiscontinuityEnd = 14000

func hasteInDiscontinuityRange(value uint32) bool {
	return value >= c_hasteDiscontinuityStart && value <= c_hasteDiscontinuityEnd
}

func checkBadHasteRange(printer *util.PrintRecorder, currentHaste uint32, incrementBaseHaste int32, plannedIncrementTestRange int32) bool {
	printer.Printf("Current gear haste %d\n", currentHaste)
	minValue := int32(currentHaste) + incrementBaseHaste
	maxValue := int32(currentHaste) + incrementBaseHaste + plannedIncrementTestRange
	printer.Printf("Planned simulated gear haste %d-%d\n", minValue, maxValue)

	return maxValue > c_hasteDiscontinuityStart && minValue < c_hasteDiscontinuityEnd
}

func checkBadExpertRange(printer *util.PrintRecorder, current uint32, decrementBase int32, plannedIncrementTestRange int32) bool {
	printer.Printf("Current gear expertise %d\n", current)
	minValue := int32(current) - decrementBase
	maxValue := int32(current) - decrementBase + plannedIncrementTestRange
	printer.Printf("Planned simulated gear expertise %d-%d\n", minValue, maxValue)
	return maxValue > int32(requirements.TARGET_RATING_TANK)
}

func fixBadHasteRange(printer *util.PrintRecorder, currentHaste uint32, plannedIncrementTestRange int32, preferHigh bool) int32 {
	minStat := int32(currentHaste)
	maxStat := int32(currentHaste) + plannedIncrementTestRange

	var fix int32
	if minStat >= c_hasteDiscontinuityEnd {
		// above trouble range
		fix = 0
	} else if maxStat <= c_hasteDiscontinuityStart {
		// below trouble range
		fix = 0
	} else if preferHigh || (maxStat-c_hasteDiscontinuityStart) >= (c_hasteDiscontinuityEnd-minStat) {
		// positive fix to be above range
		fix = c_hasteDiscontinuityEnd - minStat
	} else {
		// negative fix to be below range
		fix = c_hasteDiscontinuityStart - maxStat
	}

	if fix != 0 {
		printer.Printf("Current gear haste %d, planned simulation %d-%d, corrected %d-%d\n", currentHaste, minStat, maxStat, minStat+fix, maxStat+fix)
	}

	if !(minStat+fix >= c_hasteDiscontinuityEnd || maxStat+fix <= c_hasteDiscontinuityStart) {
		panic("didn't fix range")
	}

	return fix
}

func fixBadExpertRange(printer *util.PrintRecorder, currentExpert uint32, plannedIncrementTestRange int32) int32 {
	minValue := int32(currentExpert)
	maxValue := int32(currentExpert) + plannedIncrementTestRange

	var fix int32
	if maxValue >= int32(requirements.TARGET_RATING_TANK) {
		fix = int32(requirements.TARGET_RATING_TANK) - maxValue
	}

	if fix != 0 {
		printer.Printf("Current gear expertise %d, planned simulation %d-%d, corrected %d-%d\n", currentExpert, minValue, maxValue, minValue+fix, maxValue+fix)
	}
	return fix
}

func InitialBonusStatMap(printer *util.PrintRecorder, currentItemSet items.FullItemSet, incrementBaseHaste int32, decrementBaseExpertise int32, incrementMax int32) map[stats.StatType]int32 {
	if checkBadHasteRange(printer, currentItemSet.Total().GetUInt(stats.Stat_Haste), incrementBaseHaste, incrementMax) {
		panic("haste in discontinuity range")
	}
	if checkBadExpertRange(printer, currentItemSet.Total().Expertise(), decrementBaseExpertise, incrementMax) {
		panic("simulate will overcap expertise")
	}
	initialBaseStats := make(map[stats.StatType]int32)
	initialBaseStats[stats.Stat_Haste] += incrementBaseHaste
	initialBaseStats[stats.Stat_Expertise] -= decrementBaseExpertise
	return initialBaseStats
}

func InitialBonusStatMap_fixRanges(printer *util.PrintRecorder, currentItemSet items.FullItemSet, plannedIncrementTestRange int32, fixMode weight_types.FixStatsRangeMode, isForGrid bool) *stats.StatTypeMap[int32] {
	initialBaseStats := &stats.StatTypeMap[int32]{}

	if fixMode == weight_types.FixStatsRangeMode_NotSet {
		panic("fixMode not specified")
	}

	if fixMode&weight_types.FixStatsRangeMode_ExpertiseAlways != 0 {
		incrementBaseExpertise := fixBadExpertRange(printer, currentItemSet.Total().Expertise(), plannedIncrementTestRange)
		initialBaseStats.Put(stats.Stat_Expertise, incrementBaseExpertise)
	}

	hasteValue := currentItemSet.Total().GetUInt(stats.Stat_Haste)
	switch {
	case fixMode&weight_types.FixStatsRangeMode_HasteAlways != 0:
		incrementBaseHaste := fixBadHasteRange(printer, hasteValue, plannedIncrementTestRange, false)
		initialBaseStats.Put(stats.Stat_Haste, incrementBaseHaste)

	case fixMode&(weight_types.FixStatsRangeMode_HasteHigherOnly|weight_types.FixStatsRangeMode_HasteGridOnly) != 0:
		if isForGrid {
			incrementBaseHaste := fixBadHasteRange(printer, hasteValue, plannedIncrementTestRange, true)
			initialBaseStats.Put(stats.Stat_Haste, incrementBaseHaste)
		}

	case fixMode&weight_types.FixStatsRangeMode_HasteHigherOnly != 0:
		incrementBaseHaste := fixBadHasteRange(printer, hasteValue, plannedIncrementTestRange, true)
		initialBaseStats.Put(stats.Stat_Haste, incrementBaseHaste)

	case fixMode&weight_types.FixStatsRangeMode_HasteGridOnly != 0:
		if isForGrid {
			incrementBaseHaste := fixBadHasteRange(printer, hasteValue, plannedIncrementTestRange, false)
			initialBaseStats.Put(stats.Stat_Haste, incrementBaseHaste)
		}
	}

	return initialBaseStats
}
