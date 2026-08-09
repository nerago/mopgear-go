package main

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/gear_model/model_factory"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
	"strconv"
)

func basicReforge(printer *util.PrintRecorder) {
	model2 := model_factory.Model_PallyProtMitigation_WithSet()
	itemOptions, model := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationWithSet, &model2, setup.MissingEnchant_Panic, printer), model2

	output := solver.Solver(solver.SolveInput{
		ItemOptions: &itemOptions,
		Model:       &model,
		Printer:     nil})
	output.Report(printer)
}

func findBestSubjectToCommon(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtMitigation_WithSet()

	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationWithSet, &model, setup.MissingEnchant_Panic, printer)

	for _, itemId := range substituteItemsProt {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	common := commonComboCurrent()
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtCompromise, &model, printer)
	addGearFileToCommon(common, files.GearFileProtDps, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigationNoSet, &model, printer)
	restrictOptionsToCommon(common, &itemOptions)

	restrictSlotToId(&itemOptions, items.Equip_Ring1, 96481)

	output := solver.Solver(solver.SolveInput{
		ItemOptions: &itemOptions,
		Model:       &model,
		Printer:     printer})

	output.Report(printer)
}

func testSim(printer *util.PrintRecorder) {
	// testSimA(printer)
	// testSimB(printer)
	testSimEach(printer)
}

func testSimA(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtDps()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &model, setup.MissingEnchant_Panic, printer)
	// itemOptionsMit := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &model, setup.MissingEnchant_Panic, printer)
	// itemOptions[items.Equip_Trinket2] = itemOptionsMit[items.Equip_Trinket2]
	output := solver.Solver(solver.SolveInput{
		ItemOptions: &itemOptions,
		Model:       &model,
		Printer:     printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_QuickDirty, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}
func testSimB(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtMitigation_WithSet()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationWithSet, &model, setup.MissingEnchant_Panic, printer)
	output := solver.Solver(solver.SolveInput{
		ItemOptions: &itemOptions,
		Model:       &model,
		Printer:     printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_Common, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}
func testSimEach(printer *util.PrintRecorder) {
	modelList := []gear_model.SpecModel{model_factory.Model_PallyProtDps(), model_factory.Model_PallyProtCompromise(), model_factory.Model_PallyProtMitigation_NoSet(), model_factory.Model_PallyProtMitigation_WithSet(), model_factory.Model_PallyProtHeal()}
	for model := range util_collection.ForPointer(modelList) {
		equipped := loaders.GearFileReader_Read(model.ReferenceGearFile)
		equipSet := setup.OptionsSetup_ExactEquippedOnly(equipped, model, setup.MissingEnchant_Fix, util.PrintRecorder_Nop())
		// itemOptions := setup.OptionsSetup_FromGearFile(model.ReferenceGearFile, model, setup.MissingEnchant_Panic, printer)
		// output := solver.Solver(solver.SolveInput{
		// 	ItemOptions:         &itemOptions,
		// 	Model:               model,
		// 	EnableTrackProgress: true,
		// 	Printer:             printer})
		// printer.Printf("Running sim\n")
		// equipSet := output.FullSet.Items()
		resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_QuickDirty, model, &equipSet, nil, util.TrackProgress_Start())
		resultStats.Print(printer)
	}
}

func findSimpleUpgrade(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtMitigation_NoSet()
	gearFile := files.GearFileProtMitigationNoSet
	//model := model.Model_PallyProtCompromise()
	//gearFile := files.GearFileProtCompromise

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), &model, setup.MissingEnchant_Panic, printer)

	itemOptions := setup.OptionsSetup_FromGearFile(gearFile, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItemsProt {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	common := commonComboCurrent()
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtDps, &model, printer)
	addGearFileToCommon(common, files.GearFileProtCompromise, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigationWithSet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigationNoSet, &model, printer)
	//restrictOptionsToCommon(common, &itemOptions)

	addItems := []items.ItemId{95038, 95003, 95022, 95002, 95023}
	for _, itemId := range addItems {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	output := solver.Solver(solver.SolveInput{
		ItemOptions: &itemOptions,
		Model:       &model,
		Printer:     printer})

	output.Report(printer)

	currentStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_VerySlow, &model, &currentEquip, nil, util.TrackProgress_Start())
	resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_VerySlow, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())

	printer.Println("CURRENT STATS")
	currentStats.Print(printer)

	printer.Println("NEW SET STATS")
	resultStats.Print(printer)

	resultStats.QueryIncreaseOfEach(&currentStats).Print(printer)
}

func findSimpleUpgrade_ForceEach(printer *util.PrintRecorder) {
	// simSize := simulate.RunSize_Medium
	simSize := simulate.RunSize_QuickDirty

	// model := model.Model_PallyProtMitigation_NoSet()
	model := model_factory.Model_PallyProtMitigation_WithSet()
	startGear := files.GearFileProtMitigationWithSet

	printer.Println("READ existing")
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &model, setup.MissingEnchant_Panic, printer)
	currentStats := simulate.WowSim_Execute_UseModel(simSize, &model, &currentEquip, nil, util.TrackProgress_Start())

	printer.Println("SETUP options")
	itemOptionsShared := setup.OptionsSetup_FromGearFile(startGear, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItemsProt {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptionsShared.AddSeveralOptions(example.SlotItem(), opts)
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

		_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)

		for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
			if slotEquip == items.Equip_Ring1 {
				continue
			}

			printer.Printf("<<<<< FORCING %s %s >>>>>\n", slotEquip.Name(), example.BaseName())

			itemOptionsSpecific := itemOptionsShared.Clone()
			restrictSlotToId(&itemOptionsSpecific, slotEquip, itemId)

			output := solver.Solver(solver.SolveInput{
				ItemOptions: &itemOptionsSpecific,
				Model:       &model,
				Printer:     printer})
			output.Report(printer)

			resultStats := simulate.WowSim_Execute_UseModel(simSize, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())

			printer.Println("CURRENT STATS")
			currentStats.Print(printer)

			printer.Println("NEW SET STATS")
			resultStats.Print(printer)

			printer.Printf("INCREASE STATS (%s %s)\n", slotEquip.Name(), example.BaseName())
			resultStats.QueryIncreaseOfEach(&currentStats).Print(printer)

			mitiInc := resultStats.QueryIncreaseMitigation(&currentStats)
			dpsInc := resultStats.QueryIncreaseOf(&currentStats, stats.Sim_DPS)
			printer.Printf("INCREASE miti=%.3f dps=%.3f\n", mitiInc, dpsInc)

			resultPairs = append(resultPairs, pair{mitiInc, dpsInc, example.BaseName()})
		}
	}

	slices.SortFunc(resultPairs, func(a, b pair) int { return cmp.Compare(a.miti, b.miti) })

	for _, res := range resultPairs {
		printer.Printf("%9.3f %9.3f %s\n", res.miti, res.dps, res.name)
	}
}

func trinketSims(printer *util.PrintRecorder) {
	itemIds := []items.ItemId{
		94519,                    // crit prim rage
		945190,                   // crit->master prim rage
		96793,                    // none fort zand
		96555,                    // none soul barrier
		94529,                    // none gaze twins
		94527,                    // ji-kun
		945270,                   // exp->crit ji-kun
		94507,                    // valor
		94508,                    // valor
		103989,                   // timeless alacrity of xuen
		103990,                   // timeless resolve of niuzao
		103678,                   // time lost artifict
		trinketThokTailCelestial, // thok trinket
		trinketFusionCoreCelestial,
		trinketVialCorruptNormal,
	}

	// fight := stats.Fight_Animus
	fight := stats.Fight_Horridon_LowHeal

	type group struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	groups := []group{
		{
			"with_set",
			model_factory.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		}, {
			"compromise",
			model_factory.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"dps",
			model_factory.Model_PallyProtDps(),
			files.GearFileProtDps,
		}, {
			"ret",
			model_factory.Model_PallyProtDps(),
			files.GearFileProtDps,
		},
	}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(8)
	csv.AddStringMany("set", "item")
	for _, statType := range stats.SimTypeList {
		csv.AddString(statType.Name())
	}
	csv.FinishColumn()

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
		printer.Println(group.label + " CURRENT")
		printer.Println(equipMap[items.Equip_Trinket1].CreateString())
		printer.Println(equipMap[items.Equip_Trinket2].CreateString())
	}

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())

		for _, itemId := range itemIds {
			var item *items.FullItem
			switch itemId {
			case 945190:
				item = db.WowSimDB_LoadItemById_AllowFallback(94519, 2, printer)
				item = tools.Reforger_SinglePreset(item, stats.ReforgeRecipe_of_pointer(stats.Stat_Crit, stats.Stat_Mastery))
			case 945270:
				item = db.WowSimDB_LoadItemById_AllowFallback(94527, 2, printer)
				item = tools.Reforger_SinglePreset(item, stats.ReforgeRecipe_of_pointer(stats.Stat_Expertise, stats.Stat_Crit))
			default:
				item = db.WowSimDB_LoadItemById_AllowFallback(itemId, 2, printer)
			}

			printer.Println(group.label + " " + item.CreateFullName())
			csv.AddStringMany(group.label, item.CreateFullName())

			var newEquip items.FullEquipMap = equipMap
			newEquip[items.Equip_Trinket2] = item
			// fullSet := items.FullItemSet_FromMap(newEquip)

			resultStats := simulate.WowSim_Execute_SpecifyAll(simulate.RunSize_Largish, model.SimSpeedUp, model.Spec, model.Goal, fight, model.Professions, &newEquip, nil, util.TrackProgress_Nop())
			resultStats.Print(printer)
			for _, statType := range stats.SimTypeList {
				csv.AddFloat64(resultStats.GetFriendly(statType), 2)
			}

			csv.FinishColumn()
		}
	}

	csv.Write(printer)
}

func trinketSimsBoth(printer *util.PrintRecorder) {
	// 103989, // timeless alacrity of xuen
	// 103990, // timeless resolve of niuzao
	// 94527,  // ji-kun

	itemIds := []items.ItemId{
		94519,  // crit prim rage
		96793,  // none fort zand
		94529,  // none gaze twins
		103678, // time lost artifact
		trinketZandSpark,
		trinketThokTailCelestial,   // up 2
		trinketFusionCoreCelestial, // 528
		trinketVialCorruptNormal,   // 567 (up 2)
		trinketRookUnluckyNormal,
		trinketSkeerBloodCelestial,
	}

	upLevel := func(id items.ItemId) int32 {
		var upgrade int32 = 0
		if id < 100000 || id == trinketVialCorruptNormal || id == trinketThokTailCelestial || id == trinketRookUnluckyNormal {
			upgrade = 2
		}
		return upgrade
	}

	fight := stats.Fight_Juggernaut_NoExternalHeal
	simRun := simulate.RunSize_Common

	type group struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	groups := []group{
		{
			"heal",
			model_factory.Model_PallyProtHeal(),
			files.GearFileProtHeal,
		},
		{
			"with_set",
			model_factory.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		},
		{
			"compromise",
			model_factory.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"dps",
			model_factory.Model_PallyProtDps(),
			files.GearFileProtDps,
		},
		{
			"ret",
			model_factory.Model_PallyRet(),
			files.GearFileRet,
		},
	}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(9)
	csv.AddStringMany("set", "item1", "item2")
	for _, statType := range stats.SimTypeList {
		csv.AddString(statType.Name())
	}
	csv.FinishColumn()

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
		printer.Println(group.label + " CURRENT")
		printer.Println(equipMap[items.Equip_Trinket1].CreateString())
		printer.Println(equipMap[items.Equip_Trinket2].CreateString())
	}

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())

		processItemIds := itemIds
		if group.label == "ret" {
			processItemIds = util_collection.MapSliceAsNew_NoPointer(processItemIds, func(x items.ItemId) items.ItemId {
				if x == trinketVialCorruptNormal {
					x = trinketEyeGalakrasCelestial
				}
				return x
			})
		}

		for _, itemIdOne := range processItemIds {
			upgrade := upLevel(itemIdOne)
			itemOne := db.WowSimDB_LoadItemById_AllowFallback(itemIdOne, upgrade, printer)
			for _, itemIdTwo := range processItemIds {
				if itemIdTwo >= itemIdOne {
					continue
				}

				upgrade2 := upLevel(itemIdOne)
				itemTwo := db.WowSimDB_LoadItemById_AllowFallback(itemIdTwo, upgrade2, printer)

				nameOne := itemOne.CreateFullName() + " " + strconv.FormatUint(uint64(itemOne.ItemLevel()), 10)
				nameTwo := itemTwo.CreateFullName() + " " + strconv.FormatUint(uint64(itemTwo.ItemLevel()), 10)
				printer.Println(group.label + " " + nameOne + " " + nameTwo)
				csv.AddStringMany(group.label, nameOne, nameTwo)

				var newEquip items.FullEquipMap = equipMap
				newEquip[items.Equip_Trinket1] = itemOne
				newEquip[items.Equip_Trinket2] = itemTwo

				resultStats := simulate.WowSim_Execute_SpecifyAll(simRun, model.SimSpeedUp, model.Spec, model.Goal, fight, model.Professions, &newEquip, nil, util.TrackProgress_Nop())
				resultStats.Print(printer)
				for _, statType := range stats.SimTypeList {
					csv.AddFloat64(resultStats.GetFriendly(statType), 2)
				}

				csv.FinishColumn()
			}
		}
	}

	csv.Write(printer)
}

func currentSimGear(printer *util.PrintRecorder) {
	fight := stats.Fight_Juggernaut_NoExternalHeal
	simRun := simulate.RunSize_Largish

	type group struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	groups := []group{
		{
			"heal",
			model_factory.Model_PallyProtHeal(),
			files.GearFileProtHeal,
		},
		{
			"with_set",
			model_factory.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		},
		{
			"compromise",
			model_factory.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"dps",
			model_factory.Model_PallyProtDps(),
			files.GearFileProtDps,
		},
		//{
		//	"ret",
		//	model.Model_PallyProtDps(),
		//	files.GearFileProtDps,
		//},
	}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(7)
	csv.AddStringMany("set")
	for _, statType := range stats.SimTypeList {
		csv.AddString(statType.Name())
	}
	csv.FinishColumn()

	for _, group := range groups {
		model := group.model
		file := group.file
		csv.AddString(group.label)

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())

		resultStats := simulate.WowSim_Execute_SpecifyAll(simRun, model.SimSpeedUp, model.Spec, model.Goal, fight, model.Professions, &equipMap, nil, util.TrackProgress_Nop())
		resultStats.Print(printer)
		for _, statType := range stats.SimTypeList {
			csv.AddFloat64(resultStats.GetFriendly(statType), 2)
		}

		csv.FinishColumn()
	}

	csv.Write(printer)
}

func addGearFileToCommon(common map[items.ItemId]stats.ReforgeRecipe, gearFile string, model *gear_model.SpecModel, printer *util.PrintRecorder) {
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), model, setup.MissingEnchant_Panic, printer)
	for item := range currentEquip.AllItemSeq() {
		common[item.ItemId()] = item.Reforge()
	}
}

func restrictOptionsToCommon(common map[items.ItemId]stats.ReforgeRecipe, optionsMap *items.FullOptionsMap) {
	optionsMap.FilterAllItems(func(check *items.FullItem) bool {
		reforge, isCommon := common[check.ItemId()]
		if isCommon {
			return check.Reforge().Equals(&reforge)
		} else {
			return true
		}
	})
}

func restrictSlotToId(itemOptions *items.FullOptionsMap, slotEquip items.SlotEquip, id items.ItemId) {
	itemOptions.ForceSlotOnlySpecifiedItemId(slotEquip, id)
}

func basicListRatingEach(printer *util.PrintRecorder) {
	type group struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	groups := []group{
		{
			"ret",
			model_factory.Model_PallyRet(),
			files.GearFileRet,
		}, {
			"dps",
			model_factory.Model_PallyProtDps(),
			files.GearFileProtDps,
		}, {
			"compromise",
			model_factory.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		},
		{
			"with_set",
			model_factory.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		},
	}

	weightType := weight_types.WeightType(1)

	for _, group := range groups {
		equipItems := loaders.GearFileReader_Read(group.file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipItems, &group.model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
		itemSet := items.FullItemSet_FromMap(equipMap)
		rating := group.model.CalcRatingFull(&itemSet, weightType)
		tools.ReportSet(&group.model, &itemSet, printer)

		printer.Printf("%20s %10.0f %s\n", group.label, rating, group.model.StatWeights.CreateString())
	}

	// for _, group := range groups {
	// 	rating := group.model.CalcRatingFull(itemSet)

	// }
}

func solveForRatings(printer *util.PrintRecorder) {
	type group struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	var prescaleTarget float64 = 100000000.0

	groups := []group{
		{
			"ret",
			model_factory.Model_PallyRet(),
			files.GearFileRet,
		}, {
			"dps",
			model_factory.Model_PallyProtDps(),
			files.GearFileProtDps,
		}, {
			"compromise",
			model_factory.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		},
		{
			"with_set",
			model_factory.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		},
	}

	weightType := weight_types.WeightType(1)

	for _, group := range groups {
		equipItems := loaders.GearFileReader_Read(group.file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipItems, &group.model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
		itemSet := items.FullItemSet_FromMap(equipMap)
		rating := group.model.CalcRatingFull(&itemSet, weightType)
		// solver.ReportSet(printer, itemSet, rating, &group.model)

		prescaleMult := prescaleTarget / rating

		printer.Printf("%20s %10.0f %.4f %s\n", group.label, rating, rating*prescaleMult, group.model.StatWeights.CreateString())
	}
}

type standardisedItemSet struct {
	itemSet    items.FullItemSet
	bonusStats stats.StatTypeMap[int32]
}
type standardisedItemSetGroup struct {
	zero, two, four standardisedItemSet
}

func determineSetBonusValueBySim() {
	//runSize := simulate.RunSize_QuickDirty
	optionCount := 10
	//runSize := simulate.RunSize_Largish
	//optionCount := 128

	//goal := stats.OptimiseGoal_Mitigation
	goal := stats.OptimiseGoal_HalfMitiHeal
	fight := stats.Fight_Juggernaut_NoExternalHeal
	spec := stats.Spec_PaladinProt
	profession := gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true}

	//startGear := files.GearFileProtMitigationWithSet
	//modelEquipOnly := model_factory.Model_PallyProtMitigation_WithSet()
	//targetRatio := modelEquipOnly.SimPriority

	gearFile := files.GearFileProtMitigationNoSet
	model := model_factory.Model_PallyProtMitigation_NoSet()
	model.BonusEnabled = bonus_set.SpecSetsEnableNamed("Plate of Winged Triumph")
	model.BonusRequiredSolve = bonus_set.ItemCountsRequiredOptions{model_factory.BonusItems_ZeroAll}
	model.BonusRequiredWeight = nil

	initialSets, itemOptions := weightfind.GenerateRandomSets(gearFile, substituteItemsProt, &model, optionCount, printer, "")

	setItems := []*items.FullItem{
		itemOptions.FindItemIdFirst(99126),
		itemOptions.FindItemIdFirst(99128),
		itemOptions.FindItemIdFirst(99129),
		itemOptions.FindItemIdFirst(99130),
	}

	preparedSetGroups := util_collection.MapSliceAsNew(initialSets, func(itemSet *items.FullItemSet) standardisedItemSetGroup {
		return standardisedItemSetGroup{
			standardisedItemSet{*itemSet, stats.StatTypeMap[int32]{}},
			replaceWithEquivalentSetItems(itemSet, setItems, 2),
			replaceWithEquivalentSetItems(itemSet, setItems, 4),
		}
	})

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(preparedSetGroups) * 3)

	bonusData := util_async.Map_SliceToSlice(4, preparedSetGroups, func(group *standardisedItemSetGroup) bonuses {
		dataZero := simulate.WowSim_Execute_SpecifyAll(runSize, 0, spec, goal, fight, profession,
			group.zero.itemSet.Items(), nil,
			tracker.NewChild())
		dataTwo := simulate.WowSim_Execute_SpecifyAll(runSize, 0, spec, goal, fight, profession,
			group.two.itemSet.Items(), &group.two.bonusStats,
			tracker.NewChild())
		dataFour := simulate.WowSim_Execute_SpecifyAll(runSize, 0, spec, goal, fight, profession,
			group.four.itemSet.Items(), &group.four.bonusStats,
			tracker.NewChild())

		return bonuses{
			twoPiece:  makeBonusDiff(dataZero, dataTwo),
			fourPiece: makeBonusDiff(dataTwo, dataFour),
		}
	})

	for simType := range stats.SimTypeEnum.ValueSeq() {
		average2 := util_collection.FindAverageFunc(bonusData, func(x bonuses) float64 {
			return x.twoPiece.GetOrPanic(simType)
		})
		printer.Printf("BONUS 2 %s %f\n", simType.Name(), average2)
	}
	for simType := range stats.SimTypeEnum.ValueSeq() {
		average4 := util_collection.FindAverageFunc(bonusData, func(x bonuses) float64 {
			return x.fourPiece.GetOrPanic(simType)
		})
		printer.Printf("BONUS 4 %s %f\n", simType.Name(), average4)
	}
}

func makeBonusDiff(lo stats.SimData, hi stats.SimData) util_collection.EnumMap[stats.SimType, float64] {
	result := util_collection.EnumMapMake[stats.SimType, float64](stats.SimTypeEnum)
	for simType := range stats.SimTypeEnum.ValueSeq() {
		ratio := hi.Get(simType) / lo.Get(simType)
		result.Put(simType, ratio)
	}
	return result
}

type bonuses struct {
	twoPiece  util_collection.EnumMap[stats.SimType, float64]
	fourPiece util_collection.EnumMap[stats.SimType, float64]
}

func replaceWithEquivalentSetItems(baseSet *items.FullItemSet, bonusItems []*items.FullItem, bonusCount int) standardisedItemSet {
	var substitutedEquip items.FullEquipMap = *baseSet.Items()
	for i := range bonusCount {
		item := bonusItems[i]
		slot := item.SlotItem().ToSlotEquipOptions()[0]
		substitutedEquip[slot] = item
	}
	substitutedSet := items.FullItemSet_FromMap(substitutedEquip)

	bonusStats := stats.StatTypeMap[int32]{}
	for _, statType := range stats.StatType_List {
		baseValue := baseSet.Total().GetUInt(statType)
		subValue := substitutedSet.Total().GetUInt(statType)
		diff := int32(baseValue) - int32(subValue)
		if diff != 0 {
			bonusStats.Put(statType, diff)
		}
	}

	return standardisedItemSet{substitutedSet, bonusStats}
}
