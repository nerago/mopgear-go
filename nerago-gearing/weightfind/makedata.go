package weightfind

import (
	"maps"
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
	"paladin_gearing_go/weightfind/weight_highs"
)

const max_grid_sim_count = 600
const grid_sim_steps = 2

func SimulateSteppedStatChangesForGrid(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, tracker *util.TrackProgress) []weight_highs.WeightInput {
	var incrementMin int32 = 0
	var incrementStep int32 = 250
	var incrementMax int32 = incrementStep * grid_sim_steps

	initialBaseStats := InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementMax)

	statCheckList := requiredStats
	type incrementStat struct {
		stat  stats.StatType
		value int32
	}

	incrementOptions := make([][]incrementStat, 0)
	for _, stat := range statCheckList {
		optionArray := make([]incrementStat, 0)
		// TODO this was maybe intended to be a less or equal? could greatly increase combos if done now
		for value := incrementMin; value < incrementMax; value += incrementStep {
			entry := incrementStat{stat, value}
			optionArray = append(optionArray, entry)
		}
		incrementOptions = append(incrementOptions, optionArray)
	}
	incrementPermutations := util.PermuteAll_Slice(incrementOptions)

	printer.Printf("SimulateSteppedStatChangesForGrid incrementPermutations=%d\n", len(incrementPermutations))

	// crude cut down to size
	if len(incrementPermutations) > max_grid_sim_count {
		incrementPermutations = incrementPermutations[0:max_grid_sim_count]
	}

	tracker.RunOuterTracking(len(incrementPermutations))
	defer tracker.SetDone()

	inputList := util_async.Map_SliceToSlice(6, incrementPermutations, func(increments *[]incrementStat) weight_highs.WeightInput {
		bonusStat := maps.Clone(initialBaseStats)
		str := util.StringBuild2{}
		str.WriteString("SIM ")
		for _, inc := range *increments {
			bonusStat[inc.stat] += inc.value

			str.WriteString(inc.stat.Name())
			str.WriteRune('=')
			str.WriteInt32(bonusStat[inc.stat])
			str.WriteRune(' ')
		}

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), &bonusStat, tracker.NewChild())

		str.WriteString("--> ")
		simResult.CompactStringGeneralBuilder(&str)
		printer.PrintlnFromBuild(str)

		return weight_highs.WeightInput{
			TotalStat: addBonusStats(currentItemSet.Total(), bonusStat),
			SimResult: simResult,
		}
	})
	return inputList
}

func SimulateRealRandomSets(gearFile string, substituteItems []items.ItemId, model *gear_model.SpecModel, makeSetCount int, simSize simulate.WowSim_RunSize, doFixRanges bool, printer *util.PrintRecorder, track *util.TrackProgress) []weight_highs.WeightInput {
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

	setList := solve_build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, model, makeSetCount, 0)

	track.RunOuterTracking(len(setList))
	defer track.SetDone()

	weightInputs := util_async.Map_SliceToSlice(6, setList, func(itemSet *items.FullItemSet) weight_highs.WeightInput {
		var bonusStats *map[stats.StatType]int32 = nil
		if doFixRanges {
			bonusFix := InitialBonusStatMap_fixRanges(printer, *itemSet, 0)
			bonusStats = &bonusFix
		}

		var total stats.StatBlock
		if bonusStats != nil {
			total = addBonusStats(itemSet.Total(), *bonusStats)
		} else {
			total = *itemSet.Total()
		}

		simResult := simulate.WowSim_Execute_UseModel(simSize, model, itemSet.Items(), bonusStats, track.NewChild())

		str := util.StringBuild2{}
		str.WriteString("SIM ")
		total.AppendString(&str)
		str.WriteString(" --> ")
		simResult.CompactStringGeneralBuilder(&str)
		printer.PrintlnFromBuild(str)

		return weight_highs.WeightInput{TotalStat: total, SimResult: simResult}
	})

	return weightInputs
}

func addBonusStats(base *stats.StatBlock, bonusStat map[stats.StatType]int32) stats.StatBlock {
	resultBlock := *base
	for stat, add := range bonusStat {
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
	min := int32(currentHaste) + incrementBaseHaste
	max := int32(currentHaste) + incrementBaseHaste + plannedIncrementTestRange
	printer.Printf("Planned simulated gear haste %d-%d\n", min, max)

	return max > c_hasteDiscontinuityStart && min < c_hasteDiscontinuityEnd
}

func checkBadExpertRange(printer *util.PrintRecorder, current uint32, decrementBase int32, plannedIncrementTestRange int32) bool {
	printer.Printf("Current gear expertise %d\n", current)
	min := int32(current) - decrementBase
	max := int32(current) - decrementBase + plannedIncrementTestRange
	printer.Printf("Planned simulated gear expertise %d-%d\n", min, max)
	return max > int32(requirements.TARGET_RATING_TANK)
}

func fixBadHasteRange(printer *util.PrintRecorder, currentHaste uint32, plannedIncrementTestRange int32) int32 {
	min := int32(currentHaste)
	max := int32(currentHaste) + plannedIncrementTestRange

	var fix int32
	if min >= c_hasteDiscontinuityEnd {
		fix = 0
	} else if max <= c_hasteDiscontinuityStart {
		fix = 0
	} else if (max - c_hasteDiscontinuityStart) < (c_hasteDiscontinuityEnd - min) {
		fix = c_hasteDiscontinuityStart - max
	} else {
		fix = c_hasteDiscontinuityEnd - min
	}

	if fix != 0 {
		printer.Printf("Current gear haste %d, planned simulation %d-%d, corrected %d-%d\n", currentHaste, min, max, min+fix, max+fix)
	}

	if !(min+fix >= c_hasteDiscontinuityEnd || max+fix <= c_hasteDiscontinuityStart) {
		panic("didn't fix range")
	}

	return fix
}

func fixBadExpertRange(printer *util.PrintRecorder, currentExpert uint32, plannedIncrementTestRange int32) int32 {
	min := int32(currentExpert)
	max := int32(currentExpert) + plannedIncrementTestRange

	var fix int32
	if max >= int32(requirements.TARGET_RATING_TANK) {
		fix = int32(requirements.TARGET_RATING_TANK) - max
	}

	if fix != 0 {
		printer.Printf("Current gear expertise %d, planned simulation %d-%d, corrected %d-%d\n", currentExpert, min, max, min+fix, max+fix)
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

func InitialBonusStatMap_fixRanges(printer *util.PrintRecorder, currentItemSet items.FullItemSet, plannedIncrementTestRange int32) map[stats.StatType]int32 {
	incrementBaseHaste := fixBadHasteRange(printer, currentItemSet.Total().GetUInt(stats.Stat_Haste), plannedIncrementTestRange)
	incrementBaseExpertise := fixBadExpertRange(printer, currentItemSet.Total().Expertise(), plannedIncrementTestRange)
	initialBaseStats := make(map[stats.StatType]int32)
	initialBaseStats[stats.Stat_Haste] += incrementBaseHaste
	initialBaseStats[stats.Stat_Expertise] += incrementBaseExpertise
	return initialBaseStats
}
