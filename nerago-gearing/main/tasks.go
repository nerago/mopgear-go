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
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

func basicReforge(printer *util.PrintRecorder) {
	itemOptions, model := setupPallyMitigationSet()

	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		EnableTrackProgress: true,
		Printer:             nil})
	output.Report(printer)
}

func findBestSubjectToCommon(printer *util.PrintRecorder) {
	model := model.Model_PallyProtMitigation_WithSet()

	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationWithSet, &model, setup.MissingEnchant_Panic, printer)

	for _, itemId := range substituteItemsMiti {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)
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
		ItemOptions:         &itemOptions,
		Model:               &model,
		EnableTrackProgress: true,
		Printer:             printer})

	output.Report(printer)
}

var allSetItems = []items.ItemId{
	95291, 95290, 96666, 96667, 96668, // prot tier15
	95282, 95910, 95281, 96657, 96658, // ret tier15
	87101, 87103, 87099, 87100, 87102, // prot tier14 (heroic versions)
	87109, 87110, 87111, 87112, 87113, // ret tier14 (heroic versions)
}

var miscRings = []items.ItemId{
	96481, // durumu tentacle heroic
	86957, // heroic bladed tempest ring
	95140, // shado assault band
	95141, // shado assault loop
	96500, // scaled tyrant heroic
	96377, // jinrohk soulcrystal
}

func checkHighs(printer *util.PrintRecorder) {
	// bonus := model.SetBonus_Named("White Tiger Battlegear Prot Mitigation", "White Tiger Plate", "Plate of the Lightning Emperor Prot Mitigation", "Battlegear of the Lightning Emperor")
	// bonus := model.SetBonus_Named("White Tiger Battlegear Prot Mitigation", "White Tiger Plate", "Plate of the Lightning Emperor Prot Mitigation")
	// bonus := model.SetBonus_Named("White Tiger Battlegear Prot Mitigation", "White Tiger Plate", "Plate of the Lightning Emperor Prot Damage")
	// bonus := model.SetBonus_Named( "Plate of the Lightning Emperor Prot Damage", "White Tiger Plate")
	// bonus := model.SetBonus_Named( "Plate of the Lightning Emperor Prot Damage")
	// bonus := model.SetBonus_Empty()

	// itemOptions, model := setupPallyMitigationSet()
	// model.SetBonus = bonus

	// model := model.Model_PallyRet()
	// itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileRet, &model, setup.MissingEnchant_Panic, printer)

	model := model.Model_PallyProtDps()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &model, setup.MissingEnchant_Panic, printer)

	// extraItemsCombined := slices.Concat(substituteItemsMiti, allSetItems, miscRings)

	extraItemsCombined := []items.ItemId{
		86957, // heroic bladed tempest ring
		// 95140, // shado assault band
		// 95141, // shado assault loop OFF123
		// 96500, // scaled tyrant heroic
		// 96481, // durumu tentacle heroic
		// 96377, // jinrohk soulcrystal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		96478, // treads of the blind heroic

		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic

		preLegendMeleeCloak, // pre-legend strength dps

		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate

		96550, // doomed crown heroic
		87024, // null greathelm
		95778, // crown golden golem celestial [would need gem]
		95282, // ret tier15 normal head
		95292, // prot tier15 head normal

		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic

		95290, // prot tier15 chest normal
		95291, // prot tier15 hand normal
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags OFF12
		95205, // terra-cotta neck
		94776, // primal turtle amulet
		96420, // talisman of angry spirits

		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		96375, // bracers implosion
		96395, // bloodsoaked legplates

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic
	}

	for _, itemId := range extraItemsCombined {
		if !itemOptions.IncludesItemId(itemId) {
			opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)
			for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
				if itemOptions.Has(slotEquip) {
					itemOptions.AddSeveralOptionsSpecific(slotEquip, opts)
				}
			}
		}
	}

	// itemOptions.RemoveItemIdFromAll(96500)
	// itemOptions.RemoveItemIdFromAll(96481)
	// itemOptions.RemoveItemIdFromAll(96377)
	// itemOptions.RemoveItemIdFromAll(95140)
	// itemOptions.RemoveItemIdFromAll(95141)

	solveOptions := items.SolvableOptionsMap_of(&itemOptions)
	// solvedSet := withhighs.RunSingleAcrossSets_ReturnBest(&solveOptions, &model, printer)
	solvedSet := withhighs.SolverHighsMain(&solveOptions, &model, printer)
	// solvedSet := withhighs.RunBasic(&solveOptions, &model, nil, util.Optional_Empty[int]())

	var fullItemSet items.FullItemSet
	if solvedSet.IsEmpty() {
		printer.Println("FAILED SOLVE")
	} else {
		fullItemSet = items.FullItemSet_FromSolved(solvedSet.GetOrPanic(), &itemOptions)
		fullItemSet.DebugValidate()
		fullItemSet.ValidateItemRules()
		tools.ReportSetFewerParams(&model, &fullItemSet, printer)
	}
}

func testSim(printer *util.PrintRecorder) {
	// testSimA(printer)
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
		EnableTrackProgress: true,
		Printer:             printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_QuickDirty, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}
func testSimB(printer *util.PrintRecorder) {
	model := model.Model_PallyProtMitigation_WithSet()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationWithSet, &model, setup.MissingEnchant_Panic, printer)
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		EnableTrackProgress: true,
		Printer:             printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_Medium, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}

func slotRating(printer *util.PrintRecorder) {
	itemOptions, model := setupPallyMitigationNoSet()

	itemArray := itemOptions[items.Equip_Chest]

	printer.Println("RATINGS")
	// printer.Println(model.StatRatings.(ratings.StatRatingsWeights).Weights())
	printer.Println(model.StatRatings.Weights().CreateString())
	printer.Println0()

	best := util_rank.BestCollector1[items.FullItem]{}
	for _, item := range itemArray {
		rate := model.CalcRatingFullItem(&item)
		printer.Println(item.CreateString())
		printer.Printf("%.0f\n\n", rate)
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
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	common := commonComboCurrent()
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtDps, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigationWithSet, &model, printer)
	restrictOptionsToCommon(common, &itemOptions)

	restrictSlotToId(&itemOptions, items.Equip_Ring1, 96481)

	// foreach substitute: force

	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		EnableTrackProgress: true,
		Printer:             printer})

	output.Report(printer)

	currentStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_SlowAccurate, &model, &currentEquip, nil, util.TrackProgress_Start())
	resultStats := simulate.WowSim_Execute_UseModel(simulate.RunSize_SlowAccurate, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())

	printer.Println("CURRENT STATS")
	currentStats.Print(printer)

	printer.Println("NEW SET STATS")
	resultStats.Print(printer)

	resultStats.IncreaseSimBreakdown(&currentStats).Print(printer)
}

func findSimpleUpgrade_ForceEach(printer *util.PrintRecorder) {
	// simSize := simulate.RunSize_Medium
	simSize := simulate.RunSize_QuickDirty

	// model := model.Model_PallyProtMitigation_NoSet()
	model := model.Model_PallyProtMitigation_WithSet()
	startGear := files.GearFileProtMitigationWithSet

	printer.Println("READ existing")
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &model, printer)
	currentStats := simulate.WowSim_Execute_UseModel(simSize, &model, &currentEquip, nil, util.TrackProgress_Start())

	printer.Println("SETUP options")
	itemOptionsShared := setup.OptionsSetup_FromGearFile(startGear, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItemsMiti {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)
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

		_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)

		for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
			if slotEquip == items.Equip_Ring1 {
				continue
			}

			printer.Printf("<<<<< FORCING %s %s >>>>>\n", slotEquip.Name(), example.BaseName())

			itemOptionsSpecific := itemOptionsShared.Clone()
			restrictSlotToId(&itemOptionsSpecific, slotEquip, itemId)

			output := solver.Solver(solver.SolveInput{
				ItemOptions:         &itemOptionsSpecific,
				Model:               &model,
				EnableTrackProgress: true,
				Printer:             printer})
			output.Report(printer)

			resultStats := simulate.WowSim_Execute_UseModel(simSize, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())

			printer.Println("CURRENT STATS")
			currentStats.Print(printer)

			printer.Println("NEW SET STATS")
			resultStats.Print(printer)

			printer.Printf("INCREASE STATS (%s %s)\n", slotEquip.Name(), example.BaseName())
			resultStats.IncreaseSimBreakdown(&currentStats).Print(printer)

			mitiInc := resultStats.IncreaseMitigation(&currentStats)
			dpsInc := resultStats.IncreaseOf(&currentStats, simulate.Result_DPS)
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
		94519,  // crit prim rage
		945190, // crit->master prim rage
		87063,  // none vial dragon
		95779,  // none vial sang
		96793,  // none fort zand
		96555,  // none soul barrier
		87172,  // none darkmist
		94529,  // none gaze twins
		94527,  // ji-kun
		945270, // exp->crit ji-kun
		// 94507,  // valor
		// 94508,  // valor
	}

	// fight := stats.Fight_Animus
	fight := stats.Fight_Horridon_LowHeal

	type group struct {
		label string
		model model.Model
		file  string
	}

	groups := []group{
		{
			"with_set",
			model.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
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
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, util.PrintRecorder_HoldAll())
		printer.Println(group.label + " CURRENT")
		printer.Println(equipMap[items.Equip_Trinket1].CreateString())
		printer.Println(equipMap[items.Equip_Trinket2].CreateString())
	}

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipped, &model, util.PrintRecorder_HoldAll())

		for _, itemId := range itemIds {
			var item *items.FullItem
			switch itemId {
			case 945190:
				item = db.WowSimDB_ByIdAndUpgrade_AllowFallback(94519, 2, printer)
				item = tools.Reforger_SinglePreset(item, stats.ReforgeRecipe_of_pointer(stats.Stat_Crit, stats.Stat_Mastery))
			case 945270:
				item = db.WowSimDB_ByIdAndUpgrade_AllowFallback(94527, 2, printer)
				item = tools.Reforger_SinglePreset(item, stats.ReforgeRecipe_of_pointer(stats.Stat_Expertise, stats.Stat_Crit))
			default:
				item = db.WowSimDB_ByIdAndUpgrade_AllowFallback(itemId, 2, printer)
			}

			printer.Println(group.label + " " + item.CreateFullName())
			csv.AddStringMany(group.label, item.CreateFullName())

			var newEquip items.FullEquipMap = equipMap
			newEquip[items.Equip_Trinket2] = item
			// fullSet := items.FullItemSet_FromMap(newEquip)

			resultStats := simulate.WowSim_Execute_SpecifyAll(simulate.RunSize_Medium, model.Spec, model.Goal, fight, model.Professions, &newEquip, nil, util.TrackProgress_Nop())
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
		model model.Model
		file  string
	}

	groups := []group{
		{
			"ret",
			model.Model_PallyRet(),
			files.GearFileRet,
		}, {
			"dps",
			model.Model_PallyProtDps(),
			files.GearFileProtDps,
		}, {
			"compromise",
			model.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"no_set",
			model.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		},
		{
			"with_set",
			model.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		},
	}

	for _, group := range groups {
		equipItems := loaders.GearFileReader_Read(group.file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipItems, &group.model, util.PrintRecorder_HoldAll())
		itemSet := items.FullItemSet_FromMap(equipMap)
		rating := group.model.CalcRatingFull(&itemSet)
		tools.ReportSet(&group.model, &itemSet, rating, printer)

		printer.Printf("%20s %10.0f %s\n", group.label, rating, group.model.StatRatings.Weights().CreateString())
	}

	// for _, group := range groups {
	// 	rating := group.model.CalcRatingFull(itemSet)

	// }
}

func solveForRatings(printer *util.PrintRecorder) {
	type group struct {
		label string
		model model.Model
		file  string
	}

	var prescaleTarget float64 = 100000000.0

	groups := []group{
		{
			"ret",
			model.Model_PallyRet(),
			files.GearFileRet,
		}, {
			"dps",
			model.Model_PallyProtDps(),
			files.GearFileProtDps,
		}, {
			"compromise",
			model.Model_PallyProtCompromise(),
			files.GearFileProtCompromise,
		}, {
			"no_set",
			model.Model_PallyProtMitigation_NoSet(),
			files.GearFileProtMitigationNoSet,
		},
		{
			"with_set",
			model.Model_PallyProtMitigation_WithSet(),
			files.GearFileProtMitigationWithSet,
		},
	}

	for _, group := range groups {
		equipItems := loaders.GearFileReader_Read(group.file)
		equipMap := setup.OptionsSetup_ExactEquippedOnly(equipItems, &group.model, util.PrintRecorder_HoldAll())
		itemSet := items.FullItemSet_FromMap(equipMap)
		rating := group.model.CalcRatingFull(&itemSet)
		// solver.ReportSet(printer, itemSet, rating, &group.model)

		prescaleMult := prescaleTarget / rating

		printer.Printf("%20s %10.0f %.4f %s\n", group.label, rating, rating*prescaleMult, group.model.StatRatings.Weights().CreateString())
	}

	// for _, group := range groups {
	// 	rating := group.model.CalcRatingFull(itemSet)

	// }
}
