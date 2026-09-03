package weightfind

import (
	"math"

	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/requirements"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/solver/solve_build"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

const (
	grid_sim_max_steps = 6
	grid_sim_step      = 500

	//lets say we do about another 200, about 25 per stat
	//range of our current stats is about 5000-10000
	//lets say can plausibly get them to 20000, +10000
	//20205,0,34261,0,0,2634,3883,12048,4847,3233,3254,6947
	fit_sim_step_down_count = 10
	fit_sim_step_down_inc   = 500
	fit_sim_step_up_count   = 10
	fit_sim_step_up_inc     = 750
)

type incrementStatCombo stats.StatTypeMap[int32]

func SimulateSteppedStatChangesForFitting(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, tracker *util.TrackProgress, label string, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	initialBaseStats := stats.StatTypeMap[int32]{}
	for _, statType := range requiredStats {
		initialBaseStats.Put(statType, 0)
	}

	incrementCombos := makeCombosForFittingSim(requiredStats)
	tracker.RunOuterTracking(len(incrementCombos))
	defer tracker.SetDone()

	printer.Printf("Running %d sims (part 3) for %s\n", len(incrementCombos), label)
	inputList, err := util_async.MapOptional_SliceToSlice_Cancellable_PassError(6, incrementCombos, cancel, func(increments *incrementStatCombo) (weight_types.WeightInput, bool, error) {
		return simForCombo(increments, &initialBaseStats, label, simSpeed, speedUp, spec, goal, fight, profession, currentItemSet, tracker, printer)
	})
	printer.Printf("Done sims (part 3) for %s\n", label)
	return inputList, err
}

func makeCombosForFittingSim(statList []stats.StatType) []incrementStatCombo {
	allCombos := make([]incrementStatCombo, 0)

	for _, statType := range statList {
		up := int32(fit_sim_step_up_inc)
		for range fit_sim_step_up_count {
			combo := incrementStatCombo{}
			combo.Put(statType, up)
			allCombos = append(allCombos, combo)
			up += fit_sim_step_up_inc
		}

		down := int32(-fit_sim_step_down_inc)
		for range fit_sim_step_down_count {
			combo := incrementStatCombo{}
			combo.Put(statType, down)
			allCombos = append(allCombos, combo)
			down -= fit_sim_step_down_inc
		}
	}

	return allCombos
}

func SimulateSteppedStatChangesForGrid(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, tracker *util.TrackProgress, label string, cancel util_async.CancelSignal, fixStatsMode weight_types.FixStatsRangeMode, targetMaxGenerateCount int) ([]weight_types.WeightInput, error) {
	var incrementStep int32 = grid_sim_step
	var incrementMax int32 = incrementStep * grid_sim_max_steps
	if len(requiredStats) == 8 {
		incrementMax = incrementStep * 2
	}

	initialBaseStats := InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementMax, fixStatsMode, true)

	incrementPermutations := makeCombosForGridSim2(requiredStats, incrementStep, incrementMax, targetMaxGenerateCount)
	tracker.RunOuterTracking(len(incrementPermutations))
	defer tracker.SetDone()

	printer.Printf("Running %d sims (part 1) for %s\n", len(incrementPermutations), label)
	inputList, err := util_async.MapOptional_SliceToSlice_Cancellable_PassError(6, incrementPermutations, cancel, func(increments *incrementStatCombo) (weight_types.WeightInput, bool, error) {
		return simForCombo(increments, initialBaseStats, label, simSpeed, speedUp, spec, goal, fight, profession, currentItemSet, tracker, printer)
	})
	printer.Printf("Done sims (part 1) for %s\n", label)
	return inputList, err
}

func simForCombo(increments *incrementStatCombo, initialBaseStats *stats.StatTypeMap[int32], label string, simSpeed simulate.WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, currentItemSet items.FullItemSet, tracker *util.TrackProgress, printer *util.PrintRecorder) (weight_types.WeightInput, bool, error) {
	bonusStat := initialBaseStats.Clone()
	str := util.StringBuild2{}
	str.WriteString("SIM ")
	str.WriteString(label)
	str.WriteRune(' ')
	for statType, valueIncrease := range increments.SeqKeyValue() {
		bonusStat.Compute(statType, func(v int32) int32 { return v + valueIncrease })
		str.WriteString(statType.Name())
		str.WriteRune('=')
		str.WriteInt32(bonusStat.GetOrNilValue(statType))
		str.WriteRune(' ')
	}

	totalStats, isValid := addBonusStats(currentItemSet.Total(), bonusStat)
	if !isValid {
		return weight_types.WeightInput{}, false, nil
	}

	simResult, err2 := simulate.ExecuteSpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), bonusStat, tracker.NewChild())
	if err2 != nil {
		return weight_types.WeightInput{}, false, err2
	}

	str.WriteString("--> ")
	simResult.CompactStringGeneralAppend(&str)
	printer.PrintlnFromBuild(str)

	input := weight_types.WeightInput{
		TotalStat: totalStats,
		SimResult: simResult,
	}
	return input, true, nil
}

func makeCombosForGridSim2(statList []stats.StatType, incrementStep int32, incrementMax int32, targetMaxGenerateCount int) []incrementStatCombo {
	allCombos := make([]incrementStatCombo, 0)

	// TODO add negative increments too
	var incrementLo int32 = 0
	allCombos = append(allCombos, makeWithAllSameValue(statList, incrementLo))

	var incrementHi int32 = incrementStep
	for len(allCombos) < targetMaxGenerateCount && incrementHi <= incrementMax {
		// add the high version first so make sure it's more likely to make it when we cut off the list
		allCombos = append(allCombos, makeWithAllSameValue(statList, incrementHi))

		// add all exact mixes for the two levels
		levelCombos := makeAllMixesOf(statList, incrementLo, incrementHi)
		allCombos = append(allCombos, levelCombos...)

		// add all mixes for negative versions of the increments
		levelCombos = makeAllMixesOf(statList, -incrementLo, -incrementHi)
		allCombos = append(allCombos, levelCombos...)

		incrementLo += incrementStep
		incrementHi += incrementStep
	}

	if len(allCombos) > targetMaxGenerateCount {
		allCombos = allCombos[0:targetMaxGenerateCount]
	}

	return allCombos
}

func makeAllMixesOf(statList []stats.StatType, valueOne int32, valueTwo int32) []incrementStatCombo {
	slotCount := uint64(len(statList))
	itemCount := 1 << slotCount
	levelCombos := make([]incrementStatCombo, 0, itemCount-2)

	// don't include the 00000 or 11111 entries since we want to add them specially
	for index := 1; index < itemCount-1; index++ {
		combo := incrementStatCombo{}
		for statNum, statType := range statList {
			bitMask := 1 << statNum
			if (index & bitMask) == 0 {
				combo.Put(statType, valueOne)
			} else {
				combo.Put(statType, valueTwo)
			}
		}
		levelCombos = append(levelCombos, combo)
	}

	// shuffle so we get a good mix
	util_collection.Shuffle(levelCombos)
	return levelCombos
}

func makeWithAllSameValue(statList []stats.StatType, value int32) incrementStatCombo {
	combo := incrementStatCombo{}
	for _, stat := range statList {
		combo.Put(stat, value)
	}
	return combo
}

func GenerateRandomSets(gearFile string, substituteItems []items.ItemId, model *gear_model.SpecModel, makeSetCount int, printer *util.PrintRecorder, label string, includeBestWorst bool) ([]items.FullItemSet, items.FullOptionsMap, error) {
	itemOptions, err := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)
	if err != nil {
		return nil, items.FullOptionsMap{}, err
	}

	for _, itemId := range substituteItems {
		opts, example, err := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, model, printer)
		if err != nil {
			return nil, items.FullOptionsMap{}, err
		}

		example.SlotItem().ForEachEquip(func(slotEquip items.SlotEquip) {
			if itemOptions.Has(slotEquip) {
				itemOptions.AddSeveralOptionsSpecific(slotEquip, opts)
			}
		})
	}
	itemOptions.RemoveDuplicates()

	setList := solve_build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, model, makeSetCount, label)
	if includeBestWorst {
		setList = append(setList, solve_build.SolverBuildBestWorst(&itemOptions, model)...)
	}
	return setList, itemOptions, nil
}

func SimulateRealRandomSets(gearFile string, substituteItems []items.ItemId, model *gear_model.SpecModel, makeSetCount int, simSize simulate.WowSim_RunSize, fixStatsMode weight_types.FixStatsRangeMode, printer *util.PrintRecorder, track *util.TrackProgress, label string, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	setList, _, err := GenerateRandomSets(gearFile, substituteItems, model, makeSetCount, printer, label, true)
	if err != nil {
		return nil, err
	}

	track.RunOuterTracking(len(setList))
	defer track.SetDone()

	printer.Printf("Running %d sims (part 2) for %s\n", len(setList), label)
	weightInputs, err := util_async.MapOptional_SliceToSlice_Cancellable_PassError(6, setList, cancel, func(itemSet *items.FullItemSet) (weight_types.WeightInput, bool, error) {
		bonusStats := InitialBonusStatMap_fixRanges(printer, *itemSet, 0, fixStatsMode, false)

		var total stats.StatBlock
		if bonusStats != nil {
			sum, isValid := addBonusStats(itemSet.Total(), bonusStats)
			if !isValid {
				return weight_types.WeightInput{}, false, nil
			}
			total = sum
		} else {
			total = *itemSet.Total()
		}

		simResult, err2 := simulate.ExecuteUseModel(simSize, model, itemSet.Items(), bonusStats, track.NewChild())
		if err2 != nil {
			return weight_types.WeightInput{}, false, err2
		}

		str := util.StringBuild2{}
		str.WriteString("SIM ")
		str.WriteString(label)
		str.WriteRune(' ')
		total.AppendString(&str)
		str.WriteString(" --> ")
		simResult.CompactStringGeneralAppend(&str)
		printer.PrintlnFromBuild(str)

		return weight_types.WeightInput{TotalStat: total, SimResult: simResult}, true, nil
	})

	printer.Printf("Done sims (part 2) for %s\n", label)
	return weightInputs, err
}

func addBonusStats(base *stats.StatBlock, bonusStat *stats.StatTypeMap[int32]) (stats.StatBlock, bool) {
	resultBlock := *base
	for stat, add := range bonusStat.SeqKeyValue() {
		value := int64(resultBlock[stat]) + int64(add)
		if value < 0 || value > math.MaxUint32 {
			return stats.StatBlock{}, false
		}
		resultBlock[stat] = uint32(value)
	}
	return resultBlock, true
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
