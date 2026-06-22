package weightfind

import (
	"maps"
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/model/requirements"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
)

func SimulateSteppedStatChangesForGrid(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo, tracker *util.TrackProgress) []stathighs.WeightInput {
	var incrementMin int32 = 0
	var incrementMax int32 = 500
	var incrementStep int32 = 250

	initialBaseStats := InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementMax)

	statCheckList := stathighs.G_RequiredStats
	type incrementStat struct {
		stat  stats.StatType
		value int32
	}

	incrementOptions := make([][]incrementStat, 0)
	for _, stat := range statCheckList {
		optionArray := make([]incrementStat, 0)
		for value := incrementMin; value < incrementMax; value += incrementStep {
			entry := incrementStat{stat, value}
			optionArray = append(optionArray, entry)
		}
		incrementOptions = append(incrementOptions, optionArray)
	}

	incrementPermutations := util.PermuteAll_Slice(incrementOptions)

	tracker.RunOuterTracking(len(incrementPermutations))
	defer tracker.SetDone()

	inputList := channel_op.Map_SliceToSlice(6, incrementPermutations, func(increments *[]incrementStat) stathighs.WeightInput {
		innerPrint := util.PrintRecorder_HoldAll()

		bonusStat := maps.Clone(initialBaseStats)
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		for _, inc := range *increments {
			bonusStat[inc.stat] += inc.value

			str.WriteString(inc.stat.Name())
			str.WriteRune('=')
			str.WriteInt32(bonusStat[inc.stat])
			str.WriteRune(' ')
		}

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), &bonusStat, tracker.NewChild())

		innerPrint.PrintlnFromBuild(str)
		innerPrint.Println("   --> " + simResult.CompactStringGeneral())

		printer.AppendOther(innerPrint)

		return stathighs.WeightInput{
			TotalStat: addBonusStats(currentItemSet.Total(), bonusStat),
			SimResult: simResult,
		}
	})
	return inputList
}

func SimulateRealRandomSets(gearFile string, substituteItems []items.ItemId, model *model.Model, makeSetCount int, simSize simulate.WowSim_RunSize, doFixRanges bool, printer *util.PrintRecorder, track *util.TrackProgress) []stathighs.WeightInput {
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

	setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, model, makeSetCount, printer, 0)

	track.RunOuterTracking(len(setList))
	defer track.SetDone()

	weightInputs := channel_op.Map_SliceToSlice(6, setList, func(itemSet *items.FullItemSet) stathighs.WeightInput {
		var bonusStats *map[stats.StatType]int32 = nil
		if doFixRanges {
			bonusFix := InitialBonusStatMap_fixRanges(printer, *itemSet, 0)
			bonusStats = &bonusFix
		}

		simResult := simulate.WowSim_Execute_UseModel(simSize, model, itemSet.Items(), bonusStats, track.NewChild())
		return stathighs.WeightInput{TotalStat: *itemSet.Total(), SimResult: simResult}
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
	printer.Printf("Current gear haste %d\n", currentHaste)
	min := int32(currentHaste)
	max := int32(currentHaste) + plannedIncrementTestRange
	printer.Printf("Planned simulated gear haste %d-%d\n", min, max)

	var fix int32
	if max > c_hasteDiscontinuityStart {
		fix = c_hasteDiscontinuityStart - max
	} else if min < c_hasteDiscontinuityEnd {
		fix = c_hasteDiscontinuityEnd - min
	}

	if fix != 0 {
		printer.Printf("Corrected simulated gear haste %d-%d\n", min+fix, max+fix)
	}
	return fix
}

func fixBadExpertRange(printer *util.PrintRecorder, currentExpert uint32, plannedIncrementTestRange int32) int32 {
	printer.Printf("Current gear expertise %d\n", currentExpert)
	min := int32(currentExpert)
	max := int32(currentExpert) + plannedIncrementTestRange
	printer.Printf("Planned simulated gear expertise %d-%d\n", min, max)

	var fix int32
	if max >= int32(requirements.TARGET_RATING_TANK) {
		fix = int32(requirements.TARGET_RATING_TANK) - max
	}

	if fix != 0 {
		printer.Printf("Corrected simulated gear expertise %d-%d\n", min+fix, max+fix)
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
