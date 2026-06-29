package main

import (
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/model/ratings"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/weightfind"
	"slices"
	"sync/atomic"
)

func findT5BIS(printer *util.PrintRecorder) {
	// model := model.Model_PallyProtMitigation_NoSet()
	// itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationNoSet, &model, setup.MissingEnchant_Panic, printer)

	// model := model.Model_PallyProtMitigation_WithSet()
	// itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationSet, &model, setup.MissingEnchant_Panic, printer)

	model := model.Model_PallyProtCompromise()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtCompromise, &model, setup.MissingEnchant_Panic, printer)

	addExtrasT5Dumb(&itemOptions, &model, printer)

	fullItemSet := solver.Solver_Lite(&itemOptions, &model, printer)

	tools.ReportSetFewerParams(&model, &fullItemSet, printer)
	simpleItemList(fullItemSet, printer)
}

func simpleItemList(fullItemSet items.FullItemSet, printer *util.PrintRecorder) {
	printer.Println("--------------------")
	for item := range fullItemSet.Items().AllItemSeq() {
		printer.Println(item.BaseName())
	}
	printer.Println("--------------------")
	for item := range fullItemSet.Items().AllItemSeq() {
		boss := db.BossItemData_BossForItem(item)
		printer.Println(item.BaseName() + " " + boss)
	}
	printer.Println("--------------------")
}

var throneTrinkets = []items.ItemId{
	94519, // crit prim rage
	87063, // none vial dragon
	95779, // none vial sang
	96793, // none fort zand
	96555, // none soul barrier
	87172, // none darkmist
	94529, // none gaze twins
	94527, // ji-kun
	96398, // spark of zandalar
}

func addExtrasT5Dumb(itemOptions *items.FullOptionsMap, model *model.Model, printer *util.PrintRecorder) []items.ItemId {
	allTrinkets := slices.Concat(throneTrinkets, loaders.G_siegeStrengthTrinkets, loaders.G_seigeTankTrinkets)
	allTrinkets = util.RemoveDuplicatesComparable(allTrinkets)

	extraItemsCombined := slices.Concat(
		util.MapSliceAsNew(loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic), func(x **items.FullItem) items.ItemId { return (*x).ItemId() }),
		util.MapSliceAsNew(loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic), func(x **items.FullItem) items.ItemId { return (*x).ItemId() }),
		allTrinkets,
	)

	for _, itemId := range extraItemsCombined {
		// if !itemOptions.IncludesItemId(itemId) || slices.Contains(allTrinkets, itemId) {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, model, printer)
		for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
			if itemOptions.CouldAddUpgrade_EquipSlot(slotEquip, example, printer) != items.CanUpgrade_InvalidAlways {
				itemOptions.AddSeveralOptionsSpecific(slotEquip, opts)
			}
		}
		// }
	}

	itemOptions.RemoveDuplicates()

	return allTrinkets
}

func allT5stuff(model *model.Model, gearFile string, printer *util.PrintRecorder) ([]items.ItemId, items.FullOptionsMap) {
	itemOptions := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)

	allTrinkets := addExtrasT5Dumb(&itemOptions, model, printer)

	itemOptions[items.Equip_Trinket1] = util.RemoveDuplicatesFunc(itemOptions[items.Equip_Trinket1], (*items.FullItem).Equals)
	itemOptions[items.Equip_Trinket2] = util.RemoveDuplicatesFunc(itemOptions[items.Equip_Trinket2], (*items.FullItem).Equals)
	return allTrinkets, itemOptions
}

func findT5TrinketPermutations(printer *util.PrintRecorder) {
	model := model.Model_PallyProtMitigation_NoSet()
	allTrinkets, itemOptions := allT5stuff(&model, files.GearFileProtMitigationNoSet, printer)

	baseItemSet := solver.Solver_Lite(&itemOptions, &model, printer)
	printer.Println("BASELINE SET")
	tools.ReportSetFewerParams(&model, &baseItemSet, printer)

	trinkCombos := make([][2]items.ItemId, 0)
	for _, trinkA := range allTrinkets {
		for _, trinkB := range allTrinkets {
			if trinkA < trinkB {
				trinkCombos = append(trinkCombos, [2]items.ItemId{trinkA, trinkB})
			}
		}
	}

	type trinkResult struct {
		combo [2]items.ItemId
		sim   stats.SimData
	}

	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(trinkCombos))
	results := channel_op.Map_SliceToSlice(6, trinkCombos, func(combo *[2]items.ItemId) trinkResult {
		thisOptions := itemOptions.Clone()
		thisOptions.ForceSlotOnlySpecifiedItemId(items.Equip_Trinket1, combo[0])
		thisOptions.ForceSlotOnlySpecifiedItemId(items.Equip_Trinket2, combo[1])

		thisItemSet := solver.Solver_Lite(&thisOptions, &model, printer)

		thisEquip := thisItemSet.Items()
		sim := simulate.WowSim_Execute_UseModel(simulate.RunSize_QuickDirty, &model, thisEquip, nil, progress.NewChild())
		printer.Printf("%s %s %s\n", thisEquip[items.Equip_Trinket1].CreateFullName(), thisEquip[items.Equip_Trinket2].CreateFullName(), sim.CompactStringGeneral())

		return trinkResult{*combo, sim}
	})

	csv := util.CSVOutputByColumn{}
	csv.InitRows(6)
	csv.AddStringMany("trink1", "trink2", "dps", "taken", "tmi", "death")
	csv.FinishColumn()

	for _, res := range results {
		name0 := db.WowSimDB_ByIdAndUpgrade(res.combo[0], 0).BaseName()
		csv.AddString(name0)

		name1 := db.WowSimDB_ByIdAndUpgrade(res.combo[1], 0).BaseName()
		csv.AddString(name1)

		csv.AddFloat64(res.sim.DPS(), 0)
		csv.AddFloat64(res.sim.DTPS(), 0)
		csv.AddFloat64(res.sim.TMI(), 3)
		csv.AddFloat64(res.sim.DEATH()*100, 3)
		csv.FinishColumn()
	}

	csv.Write(printer)
}

func findT5WeightPermutations(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	// simRunSize := simulate.RunSize_TestOnly

	loadModel := model.Model_PallyProtMitigation_NoSet()
	_, itemOptions := allT5stuff(&loadModel, files.GearFileProtMitigationNoSet, printer)

	// ALL
	statListAll := []stats.StatType{
		stats.Stat_Strength, stats.Stat_Stamina,
		stats.Stat_Crit, stats.Stat_Haste,
		stats.Stat_Expertise, stats.Stat_Dodge,
		stats.Stat_Parry, stats.Stat_Mastery}

	// DPS: dump dodge, stam
	// statList := []stats.StatType{
	// 	stats.Stat_Strength, stats.Stat_Crit, stats.Stat_Haste,
	// 	stats.Stat_Expertise, stats.Stat_Parry, stats.Stat_Mastery}

	// MITI: dump crit, expertise
	statListTest := []stats.StatType{
		stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Haste,
		stats.Stat_Dodge, stats.Stat_Parry, stats.Stat_Mastery}

	allPossibleOrders := generateAllOrders(statListTest)

	type orderIntermediate struct {
		order      []stats.StatType
		orderText  string
		itemSet    items.FullItemSet
		alterModel model.Model
	}

	type orderIntermediateGrouped struct {
		itemSet       items.FullItemSet
		intermediates []orderIntermediate
	}

	type orderResult struct {
		group orderIntermediateGrouped
		sim   stats.SimData
	}

	progress1 := util.TrackProgress_Start()
	progress1Atom := atomic.Uint64{}
	progress1.RunFromAtomicInt(&progress1Atom, uint64(len(allPossibleOrders)))
	intermediates := channel_op.Map_SliceToSlice(10, allPossibleOrders, func(order *[]stats.StatType) orderIntermediate {
		alterModel := model.Model_PallyProtMitigation_NoSet()
		alterModel.StatRatings = ratings.StatRatingsWeights_FromPriorities(*order)

		thisItemSet := solver.Solver_Lite(&itemOptions, &alterModel, printer)

		orderText := util.StringBuild2{}
		for _, stat := range *order {
			orderText.WriteString(stat.Name())
			orderText.WriteRune(' ')
		}

		orderTextStr := orderText.String()
		printer.Printf("%s %s\n", orderTextStr, thisItemSet.Total().CreateString())

		progress1Atom.Add(1)
		return orderIntermediate{*order, orderTextStr, thisItemSet, alterModel}
	})
	progress1.SetDone()

	intermediatesGrouped := make([]*orderIntermediateGrouped, 0)
outerIntermediateLoop:
	for _, inter := range intermediates {
		for _, group := range intermediatesGrouped {
			if group.itemSet.Equals(&inter.itemSet) {
				group.intermediates = append(group.intermediates, inter)
				continue outerIntermediateLoop
			}
		}
		intermediatesGrouped = append(intermediatesGrouped, &orderIntermediateGrouped{
			inter.itemSet,
			[]orderIntermediate{inter},
		})
	}

	progress2 := util.TrackProgress_Start()
	progress2.RunOuterTracking(len(intermediatesGrouped))
	results := channel_op.Map_SliceToSlice(10, intermediatesGrouped, func(group **orderIntermediateGrouped) orderResult {
		thisEquip := (*group).itemSet.Items()
		alterModel := (*group).intermediates[0].alterModel // all models should be the same for passing basic info to sim

		sim := simulate.WowSim_Execute_UseModel(simRunSize, &alterModel, thisEquip, nil, progress2.NewChild())

		printer.Printf("%s %s\n", (*group).itemSet.Total().CreateString(), sim.CompactStringGeneral())

		return orderResult{**group, sim}
	})
	progress2.SetDone()

	line := util.StringBuild2{}
	for i := range statListTest {
		line.WriteString("order")
		line.WriteInt(i)
		line.WriteRune(',')
	}
	for _, v := range statListAll {
		line.WriteString(v.Name())
		line.WriteRune(',')
	}
	for _, simType := range stats.SimTypeList {
		line.WriteString(simType.Name())
		line.WriteRune(',')
	}
	printer.PrintlnFromBuild(line)
	line.Reset()

	for _, res := range results {
		for _, inter := range res.group.intermediates {
			for i := range inter.order {
				line.WriteString(inter.order[i].Name())
				line.WriteRune(',')
			}
			for _, v := range statListAll {
				line.WriteUint32(res.group.itemSet.Total().GetUInt(v))
				line.WriteRune(',')
			}
			for _, simType := range stats.SimTypeList {
				line.WriteFloat64(res.sim.GetFriendly(simType), 3)
				line.WriteRune(',')
			}
			printer.PrintlnFromBuild(line)
			line.Reset()
		}
	}
}

func generateAllOrders(statList []stats.StatType) [][]stats.StatType {
	result := make([][]stats.StatType, 0)
	result = generateAllOrders_recur(result, statList, nil)
	return result
}

func generateAllOrders_recur(result [][]stats.StatType, statList []stats.StatType, progress []stats.StatType) [][]stats.StatType {
	if len(statList) == 0 {
		return append(result, progress)
	} else {
		for i, st := range statList {
			opt := slices.Clone(progress)
			opt = append(opt, st)
			minus := withoutIndex(statList, i)
			result = generateAllOrders_recur(result, minus, opt)
		}
		return result
	}
}

func withoutIndex(slice []stats.StatType, remove int) []stats.StatType {
	if len(slice) == 1 {
		return nil
	}

	trimmed := make([]stats.StatType, 0, len(slice)-1)
	trimmed = append(trimmed, slice[0:remove]...)
	trimmed = append(trimmed, slice[remove+1:]...)
	return trimmed
}

var initialPriorityDps = []stats.StatType{
	// indended DPS weights based on spreadsheet "weights-summary"
	stats.Stat_Strength,
	stats.Stat_Crit,
	stats.Stat_Expertise, // lower so we might have some ratings room, could be #1
	stats.Stat_Haste,
	stats.Stat_Parry,
	stats.Stat_Dodge,
	stats.Stat_Mastery,
	stats.Stat_Stamina,
}
var initialPriorityTaken = []stats.StatType{
	// indended DTPS weights based on spreadsheet "weights-summary"
	stats.Stat_Mastery,
	stats.Stat_Strength,
	stats.Stat_Parry,
	stats.Stat_Dodge,
	stats.Stat_Haste,
	stats.Stat_Expertise,
	stats.Stat_Stamina,
	stats.Stat_Crit,
}
var initialPriorityCompromise = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Mastery,
	stats.Stat_Stamina,
	stats.Stat_Crit,
	stats.Stat_Haste,
	stats.Stat_Expertise,
	stats.Stat_Dodge,
	stats.Stat_Parry,
}
var initialPriorityDeath = []stats.StatType{
	// indended DEATH weights based on spreadsheet "weights-summary"
	stats.Stat_Stamina,
	stats.Stat_Parry,
	stats.Stat_Mastery,
	stats.Stat_Haste,
	stats.Stat_Dodge,
	stats.Stat_Expertise, // i'd kind of like expertise higher
	stats.Stat_Strength,
	stats.Stat_Crit,
}

func statWeightsGridFromInitialT5(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_TestOnly
	simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_Medium
	fight := stats.Fight_Horridon_LowHeal

	weightFileOut := files.WeightMitiNoSetFile
	gearFile := files.GearFileProtMitigationNoSet
	gearModel := model.Model_PallyProtMitigation_NoSet()
	priority := initialPriorityTaken
	trinkets := [2]items.ItemId{trinketFortZand, trinketThokTailHeroic}
	statWeightsGridFromInitialT5_inner(gearModel, priority, gearFile, trinkets, fight, gearModel.SimRatioWeighting, weightFileOut, printer, simSpeed)

	weightFileOut = files.WeightMitiWithSetFile
	gearFile = files.GearFileProtMitigationWithSet
	gearModel = model.Model_PallyProtMitigation_WithSet()
	priority = initialPriorityDeath
	trinkets = [2]items.ItemId{trinketVialCorruptHeroic, trinketThokTailHeroic}
	statWeightsGridFromInitialT5_inner(gearModel, priority, gearFile, trinkets, fight, gearModel.SimRatioWeighting, weightFileOut, printer, simSpeed)

	weightFileOut = files.WeightDpsFile
	gearFile = files.GearFileProtDps
	gearModel = model.Model_PallyProtDps()
	priority = initialPriorityDps
	trinkets = [2]items.ItemId{trinketCurseHubrisHeroic, trinketSkeerBloodHeroic}
	statWeightsGridFromInitialT5_inner(gearModel, priority, gearFile, trinkets, fight, gearModel.SimRatioWeighting, weightFileOut, printer, simSpeed)

	weightFileOut = files.WeightCompromiseFile
	gearFile = files.GearFileProtCompromise
	gearModel = model.Model_PallyProtCompromise()
	priority = initialPriorityCompromise
	trinkets = [2]items.ItemId{trinketCurseHubrisHeroic, trinketThokTailHeroic}
	statWeightsGridFromInitialT5_inner(gearModel, priority, gearFile, trinkets, fight, gearModel.SimRatioWeighting, weightFileOut, printer, simSpeed)
}

func statWeightsGridFromInitialT5_inner(model model.Model, priority []stats.StatType, gearFile string, trinkets [2]items.ItemId, fight stats.WowSim_Fight, ratios stats.SimData, weightFileOut string, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize) {
	// INITIAL MODEL BASED ON PRIORITIES PREVIOUSLY GUESSED AT
	model.StatRatings = ratings.StatRatingsWeights_FromPriorities(priority)

	// COME UP WITH A GEAR SET BASED ON THAT INITIAL MODEL
	_, itemOptions := allT5stuff(&model, gearFile, printer)
	itemOptions.ForceSlotOnlySpecifiedItemId(items.Equip_Trinket1, trinkets[0])
	itemOptions.ForceSlotOnlySpecifiedItemId(items.Equip_Trinket2, trinkets[1])

	baseItemSet := solver.Solver_Lite(&itemOptions, &model, printer)
	printer.Println("BASELINE SET")
	tools.ReportSetFewerParams(&model, &baseItemSet, printer)

	var weights stathighs.WeightResult
	if false {
		// SIMULATE STAT CHANGES
		// baseLine := simulate.WowSim_Execute_SpecifyAll(simSpeed, model.Spec, model.Goal, fight, model.Professions, baseItemSet.Items(), nil, nil)
		// inputData := generateRatingsInputFromArtificalStatOverrides_ForBasic(baseItemSet, printer, simSpeed, model.Spec, model.Goal, fight, model.Professions)

		// // SOLVE FOR STAT WEIGHTS
		// process := stathighs.BasicStatWeightProcess{}
		// process.Init(printer)
		// process.SetTargetRatios(ratios)
		// process.SetBaseline(baseLine)
		// for _, input :=  range inputData {
		// 	process.AddSimData(input.IncrementStat, input.IncrementValue, input.SimResult)
		// }
		// weights = process.Run()
		// writePawnString(weights, printer)

	} else {
		// SIMULATE STAT CHANGES
		inputData := weightfind.SimulateSteppedStatChangesForGrid(baseItemSet, printer, simSpeed, model.SimSpeedUp, model.StatsForWeighting, model.Spec, model.Goal, fight, model.Professions, util.TrackProgress_Start())

		// SOLVE FOR STAT WEIGHTS
		process := stathighs.GridStatWeightProcess{}
		process.Init(printer)
		process.SetRequiredStats(model.StatsForWeighting)
		process.SetTargetRatios(ratios)
		process.SupplyData(inputData)
		weights = process.Run(nil)
	}

	pawn := tools.WritePawnString(weights, printer)
	gearJson := tools.WowSimJson_Write(baseItemSet.Items(), &model, util.PrintRecorder_HoldAll())

	writeFile(weightFileOut, pawn)
	writeFile(gearFile, gearJson)
	simpleItemList(baseItemSet, printer)
}

func writeFile(filename, content string) {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0)
	if err != nil {
		panic(err)
	}
}
