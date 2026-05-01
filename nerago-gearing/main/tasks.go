package main

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/upgrades"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

func basicReforge(printer *util.PrintRecorder) {
	itemOptions, model := setupPallyMitigationSet()

	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Long,
		Printer:             nil})
	output.Report(printer)
}

func checkHighs(printer *util.PrintRecorder) {
	itemOptions, model := setupPallyMitigationSet()

	solveOptions := items.SolvableOptionsMap_of(&itemOptions)
	solvedSet := withhighs.RunAllActiveSets(&solveOptions, &model)

	if solvedSet.IsEmpty() {
		printer.Println("FAILED SOLVE")
	} else {
		fullItemSet := items.FullItemSet_FromSolved(solvedSet.GetOrPanic(), &itemOptions)
		solver.ReportSet(printer, fullItemSet, model.CalcRatingFull(&fullItemSet), &model)
	}
}

func testSim(printer *util.PrintRecorder) {
	testSimA(printer)
	testSimB(printer)
}
func testSimA(printer *util.PrintRecorder) {
	model := model.Model_PallyProtDps()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &model, setup.MissingEnchant_Panic, printer)
	// itemOptionsMit := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &model, setup.MissingEnchant_Panic, printer)
	// itemOptions[items.Equip_Trinket2] = itemOptionsMit[items.Equip_Trinket2]
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Medium,
		Printer:             printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute(simulate.RunSize_QuickDirty, model.Spec, output.FullSet.Items(), model.Professions, nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}
func testSimB(printer *util.PrintRecorder) {
	model := model.Model_PallyProtMitigation_WithSet()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationSet, &model, setup.MissingEnchant_Panic, printer)
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Medium,
		Printer:             printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute(simulate.RunSize_QuickDirty, model.Spec, output.FullSet.Items(), model.Professions, nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}

func slotRating(itemArray []items.FullItem, model *model.Model, printer *util.PrintRecorder) {
	printer.Println("RATINGS")
	// printer.Println(model.StatRatings.(ratings.StatRatingsWeights).Weights())
	printer.Println(model.StatRatings.Weights().CreateString())
	printer.Println0()

	best := util_rank.BestCollector1[items.FullItem]{}
	for _, item := range itemArray {
		rate := model.CalcRatingFullItem(&item)
		printer.Println(item.CreateString())
		printer.Printf("%d\n\n", rate)
		best.Offer(&item, rate)
	}

	printer.Println0()
	printer.Println("BEST")
	printer.Println(best.BestObject.CreateString())
}

func findSimpleUpgrade(printer *util.PrintRecorder) {
	model := model.Model_PallyProtMitigation_WithSet()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(files.GearFileProtMitigationNoSet), &model, printer)

	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationNoSet, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItemsMiti {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)
		itemOptions.AddSeveralOptions(example.Slot, opts)
	}

	common := commonComboCurrent()
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtDps, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigationSet, &model, printer)
	restrictOptionsToCommon(common, &itemOptions)

	restrictSlotToId(&itemOptions, items.Equip_Ring1, 96481)

	// foreach substitute: force

	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Long,
		Printer:             printer})

	output.Report(printer)

	currentStats := simulate.WowSim_Execute(simulate.RunSize_SlowAccurate, model.Spec, &currentEquip, model.Professions, nil, util.TrackProgress_Start())
	resultStats := simulate.WowSim_Execute(simulate.RunSize_SlowAccurate, model.Spec, output.FullSet.Items(), model.Professions, nil, util.TrackProgress_Start())

	printer.Println("CURRENT STATS")
	currentStats.Print(printer)

	printer.Println("NEW SET STATS")
	resultStats.Print(printer)

	resultStats.IncreaseSimBreakdown(&currentStats).Print(printer)
}

func findMitigationWithCapicitance(printer *util.PrintRecorder) {
	// simSize := simulate.RunSize_SlowAccurate
	simSize := simulate.RunSize_Medium
	// simSize := simulate.RunSize_QuickDirty

	solveSize := solver.SolveSize_Long
	// solveSize := solver.SolveSize_Medium
	// solveSize := solver.SolveSize_PerItem

	// model := model.Model_PallyProtMitigation()
	// startGear := files.GearFileProtMitigationSet
	model := model.Model_PallyProtMitigation_NoSet()
	startGear := files.GearFileProtMitigationNoSet

	printer.Println("READ existing")
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &model, printer)
	currentStats := simulate.WowSim_Execute(simSize, model.Spec, &currentEquip, model.Professions, nil, util.TrackProgress_Start())

	printer.Println("SETUP options")
	allGear := loaders.BagsFile_PlusPaladinGear_Read()
	// itemOptions := items.FullOptionsMap{}
	itemOptions := setup.OptionsSetup_FromGearFile(startGear, &model, setup.MissingEnchant_Fix, printer)
	for _, equip := range allGear {
		if slices.Contains(ignoredItems, equip.ItemId) {
			continue
		}
		if equip.ItemId == 96534 { // missing gem in export (temporary issue)
			if len(equip.GemChoice) == 1 {
				continue
			}
			// equip.GemChoice = []uint32{76667, 76699}
		}
		equip.UpgradeStep = 2
		opts, example := setup.OptionsSetup_Single_FromEquipped(equip, &model, setup.MissingEnchant_Fix, printer)

		if example.Slot == items.Item_Head {
			if len(example.GemChoice) == 0 {
				panic("dunno")
			}
			// switch example.GemChoice[0].Id {
			// case 95344, 95346:
			// 	for i := range opts {
			// 		opts[i].GemChoice[0] = db.GemData_ById(95346)
			// 	}

			// default:
			// 	panic("unexpected meta")
			// }

			if len(equip.GemChoice) == 0 {
				printer.Printf("(head) %s none\n", example.BaseName)
			} else {
				printer.Printf("(head) %s %d\n", example.BaseName, equip.GemChoice[0])
			}
		}

		can := upgrades.CouldAddUpgradeToSet_ItemSlot(&itemOptions, example.Slot, printer, example)
		if can == upgrades.CanUpgrade_Yes || can == upgrades.CanUpgrade_Equipped_Similar {
			itemOptions.AddSeveralOptions(example.Slot, opts)
		}
	}

	common := commonComboCurrent()
	printer.Println("RESTRICT ret")
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	printer.Println("RESTRICT dps")
	addGearFileToCommon(common, files.GearFileProtDps, &model, printer)
	printer.Println("RESTRICT mitset")
	addGearFileToCommon(common, files.GearFileProtMitigationSet, &model, printer)
	printer.Println("RESTRICT mitnoset")
	addGearFileToCommon(common, files.GearFileProtMitigationNoSet, &model, printer)
	restrictOptionsToCommon(common, &itemOptions)

	printer.Println0()
	printer.Println("HEADS")
	for _, item := range itemOptions.Get(items.Equip_Head) {
		printer.Println(item.CreateString())
		for _, gem := range item.GemChoice {
			name := gem.Name()
			if name != "" {
				printer.Println("  " + name)
			}
		}
	}
	printer.Println0()
	printer.Println0()

	// restrictSlotToId(&itemOptionsShared, items.Equip_Ring1, 96481)
	// restrictSlotToId(&itemOptionsShared, items.Equip_Head, 96481)

	// remove prot piece
	itemOptions.MapSlot(items.Equip_Head, func(lst []items.FullItem) []items.FullItem {
		return util.FilterSliceAsNew(lst, func(x *items.FullItem) bool { return x.ItemId() != 95292 })
	})
	// remove White Tiger Helmet
	itemOptions.MapSlot(items.Equip_Head, func(lst []items.FullItem) []items.FullItem {
		return util.FilterSliceAsNew(lst, func(x *items.FullItem) bool { return x.ItemId() != 87101 })
	})

	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solveSize,
		Printer:             printer})
	output.Report(printer)

	resultStats := simulate.WowSim_Execute(simSize, model.Spec, output.FullSet.Items(), model.Professions, nil, util.TrackProgress_Start())

	printer.Println("CURRENT STATS")
	currentStats.Print(printer)

	printer.Println("NEW SET STATS")
	resultStats.Print(printer)

	printer.Printf("INCREASE STATS\n")
	resultStats.IncreaseSimBreakdown(&currentStats).Print(printer)

	mitiInc := resultStats.IncreaseMitigation(&currentStats)
	dpsInc := resultStats.IncreaseOf(&currentStats, simulate.Result_DPS)
	printer.Printf("INCREASE miti=%.3f dps=%.3f\n", mitiInc, dpsInc)

}

func findSimpleUpgrade_ForceEach(printer *util.PrintRecorder) {
	// simSize := simulate.RunSize_Medium
	simSize := simulate.RunSize_QuickDirty

	// solveSize := solver.SolveSize_Long
	// solveSize := solver.SolveSize_Medium
	solveSize := solver.SolveSize_PerItem

	// model := model.Model_PallyProtMitigation_NoSet()
	model := model.Model_PallyProtMitigation_WithSet()
	startGear := files.GearFileProtMitigationSet

	printer.Println("READ existing")
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &model, printer)
	currentStats := simulate.WowSim_Execute(simSize, model.Spec, &currentEquip, model.Professions, nil, util.TrackProgress_Start())

	printer.Println("SETUP options")
	itemOptionsShared := setup.OptionsSetup_FromGearFile(startGear, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItemsMiti {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)
		itemOptionsShared.AddSeveralOptions(example.Slot, opts)
	}

	common := commonComboCurrent()
	printer.Println("RESTRICT ret")
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	printer.Println("RESTRICT dps")
	addGearFileToCommon(common, files.GearFileProtDps, &model, printer)
	printer.Println("RESTRICT mitset")
	addGearFileToCommon(common, files.GearFileProtMitigationNoSet, &model, printer)
	restrictOptionsToCommon(common, &itemOptionsShared)

	restrictSlotToId(&itemOptionsShared, items.Equip_Ring1, 96481)

	type pair struct {
		miti float64
		dps  float64
		name string
	}

	resultPairs := []pair{}

	checkItems := []items.ItemId{95141}
	for _, itemId := range checkItems {
		if currentEquip.IncludesItemId(itemId) {
			continue
		}

		_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)

		for _, slotEquip := range example.Slot.ToSlotEquipOptions() {
			if slotEquip == items.Equip_Ring1 {
				continue
			}

			printer.Printf("<<<<< FORCING %s %s >>>>>\n", slotEquip.Name(), example.BaseName)

			itemOptionsSpecific := itemOptionsShared.Clone()
			restrictSlotToId(&itemOptionsSpecific, slotEquip, itemId)

			output := solver.Solver(solver.SolveInput{
				ItemOptions:         &itemOptionsSpecific,
				Model:               &model,
				PhasedAcceptable:    false,
				EnableTrackProgress: true,
				SolveSize:           solveSize,
				Printer:             printer})
			output.Report(printer)

			resultStats := simulate.WowSim_Execute(simSize, model.Spec, output.FullSet.Items(), model.Professions, nil, util.TrackProgress_Start())

			printer.Println("CURRENT STATS")
			currentStats.Print(printer)

			printer.Println("NEW SET STATS")
			resultStats.Print(printer)

			printer.Printf("INCREASE STATS (%s %s)\n", slotEquip.Name(), example.BaseName)
			resultStats.IncreaseSimBreakdown(&currentStats).Print(printer)

			mitiInc := resultStats.IncreaseMitigation(&currentStats)
			dpsInc := resultStats.IncreaseOf(&currentStats, simulate.Result_DPS)
			printer.Printf("INCREASE miti=%.3f dps=%.3f\n", mitiInc, dpsInc)

			resultPairs = append(resultPairs, pair{mitiInc, dpsInc, example.BaseName})
		}
	}

	slices.SortFunc(resultPairs, func(a, b pair) int { return cmp.Compare(a.miti, b.miti) })

	for _, res := range resultPairs {
		printer.Printf("%9.3f %9.3f %s\n", res.miti, res.dps, res.name)
	}
}

func trinketSims(printer *util.PrintRecorder) {
	itemIds := []items.ItemId{
		94519, // crit->master prim rage
		87063, // none vial dragon
		95779, // none vial sang
		95811, // none soul celesial
		96793, // none fort zand
		87172, // none darkmist
		94529, // none gaze twins
		94527, // exp->haste ji-kun
		// 94507, // (could have)
		// 94508, // (could have)
	}

	type group struct {
		label string
		model model.Model
		file  string
	}

	groups := []group{
		{
			"with_set",
			model.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationSet,
		}, {
			"no_set",
			model.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		}, {
			"compromise",
			model.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"dps",
			model.Model_PallyProtDps(),
			files.GearFileProtDps,
		},
	}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(8)
	csv.AddStringMany("set", "item")
	for _, statType := range simulate.SimResultTypeList {
		csv.AddString(statType.String())
	}
	csv.FinishColumn()

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, printer)

		for _, itemId := range itemIds {
			item := db.WowSimDB_ByIdAndUpgrade_AllowFallback(itemId, 2, printer)
			switch itemId {
			case 94519:
				item = tools.Reforger_SinglePreset(item, stats.ReforgeRecipe_of_pointer(stats.Stat_Crit, stats.Stat_Mastery))
			case 94527:
				item = tools.Reforger_SinglePreset(item, stats.ReforgeRecipe_of_pointer(stats.Stat_Expertise, stats.Stat_Haste))
			}

			printer.Println(group.label + " " + item.BaseName)
			csv.AddStringMany(group.label, item.BaseName)

			var newEquip items.FullEquipMap = equipMap
			newEquip[items.Equip_Trinket2] = item
			// fullSet := items.FullItemSet_FromMap(newEquip)

			resultStats := simulate.WowSim_Execute_SelectFight(simulate.RunSize_Medium, model.Spec, stats.Fight_Animus, &newEquip, model.Professions, nil, util.TrackProgress_Nop())
			resultStats.Print(printer)
			for _, statType := range simulate.SimResultTypeList {
				csv.AddFloat64(resultStats.GetFriendly(statType), 2)
			}

			csv.FinishColumn()
		}
	}

	csv.Write(printer)
}

func addGearFileToCommon(common map[items.ItemId]stats.ReforgeRecipe, gearFile string, model *model.Model, printer *util.PrintRecorder) {
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), model, printer)
	for item := range currentEquip.AllItemSeq() {
		common[item.ItemId()] = item.Reforge
	}
}

func restrictOptionsToCommon(common map[items.ItemId]stats.ReforgeRecipe, optionsMap *items.FullOptionsMap) {
	optionsMap.MapSlotsAll(func(slotCurrent []items.FullItem) []items.FullItem {
		return util.FilterSliceAsNew(slotCurrent, func(check *items.FullItem) bool {
			reforge, isCommon := common[check.ItemId()]
			if isCommon {
				return check.Reforge == reforge
			} else {
				return true
			}
		})
	})
}

func restrictSlotToId(itemOptions *items.FullOptionsMap, slotEquip items.SlotEquip, id items.ItemId) {
	itemOptions.MapSlot(slotEquip, func(currSlot []items.FullItem) []items.FullItem {
		return util.FilterSliceAsNew(currSlot, func(check *items.FullItem) bool {
			return check.ItemId() == id
		})
	})
}
