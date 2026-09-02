package main

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"

	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/solver"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

const c_miscDefaultTimeout = 1000

func basicReforge(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtSurvival()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtSurvival, &model, setup.MissingEnchant_Panic, printer)

	output := solver.Solver(
		&itemOptions,
		&model,
		printer,
		2,
		c_miscDefaultTimeout,
		nil)
	output.Report(printer)
}

func debugBadWeight(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtSurvival()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtSurvival, &model, setup.MissingEnchant_Panic, printer)

	//Removed row   35 (simValueFromStat DTPS) []= 9 --> Infeasible
	//Removed row   36 (simValueFromStat TMI) []= 9 --> Infeasible
	//Removed row   37 (simValueFromStat DEATH) []= 9 --> Infeasible
	//Removed row   38 (simValueFromStat DPS) []= 9 --> UnboundedOrInfeasible

	//simType := stats.Sim_DPS
	//w2 := model.StatWeights.Weight2
	//
	//w2.SimList = slices.DeleteFunc(w2.SimList, func(simType stats.SimType) bool {
	//	return simType == simType
	//})
	//w2.SimPriority.Delete(simType)
	//w2.DetailedWeights.DeleteAllForKey1(simType)
	//
	//simType = stats.Sim_DTPS
	//w2.SimList = slices.DeleteFunc(w2.SimList, func(simType stats.SimType) bool {
	//	return simType == simType
	//})
	//w2.SimPriority.Delete(simType)
	//w2.DetailedWeights.DeleteAllForKey1(simType)

	//simType = stats.Sim_TMI
	//w2.SimList = slices.DeleteFunc(w2.SimList, func(simType stats.SimType) bool {
	//	return simType == simType
	//})
	//w2.SimPriority.Delete(simType)
	//w2.DetailedWeights.DeleteAllForKey1(simType)

	output := solver.Solver(
		&itemOptions,
		&model,
		printer,
		2,
		c_miscDefaultTimeout,
		nil)
	output.Report(printer)
}

func findBestSubjectToCommon(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtSurvival()

	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtSurvival, &model, setup.MissingEnchant_Panic, printer)

	for _, itemId := range mygear.SubstituteItemsProt {
		opts, example := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	common := commonComboCurrent()
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtBalanced, &model, printer)
	addGearFileToCommon(common, files.GearFileProtDamage, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigation, &model, printer)
	restrictOptionsToCommon(common, &itemOptions)

	restrictSlotToId(&itemOptions, items.Equip_Ring1, 96481)

	output := solver.Solver(
		&itemOptions,
		&model,
		printer,
		1,
		c_miscDefaultTimeout,
		nil)

	output.Report(printer)
}

func testSim(printer *util.PrintRecorder) {
	// testSimA(printer)
	// testSimB(printer)
	testSimEach(printer)
}

func testSimA(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtDamage()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtDamage, &model, setup.MissingEnchant_Panic, printer)
	// itemOptionsMit := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &model, setup.MissingEnchant_Panic, printer)
	// itemOptions[items.Equip_Trinket2] = itemOptionsMit[items.Equip_Trinket2]
	output := solver.Solver(
		&itemOptions,
		&model,
		printer,
		1,
		c_miscDefaultTimeout,
		nil)
	printer.Println("Running sim")
	resultStats, err := simulate.ExecuteUseModel(simulate.RunSize_QuickDirty, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
	if err != nil {
		panic(err)
	}
	resultStats.Print(printer)
}
func testSimEach(printer *util.PrintRecorder) {
	modelList := []gear_model.SpecModel{model_factory.Model_PallyProtDamage(), model_factory.Model_PallyProtBalanced(), model_factory.Model_PallyProtMitigation(), model_factory.Model_PallyProtSurvival(), model_factory.Model_PallyProtHeal()}
	for model := range util_collection.ForPointer(modelList) {
		equipped := loaders.GearFileReader_Read(model.ReferenceGearFile)
		equipSet := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, model, setup.MissingEnchant_Fix, util.PrintRecorder_Nop())
		// itemOptions := setup.OptionsSetup_FromGearFile(model.ReferenceGearFile, model, setup.MissingEnchant_Panic, printer)
		// output := solver.Solver(solver.SolveInput{
		// 	ItemOptions:         &itemOptions,
		// 	Model:               model,
		// 	EnableTrackProgress: true,
		// 	Printer:             printer})
		// printer.Printf("Running sim\n")
		// equipSet := output.FullSet.Items()
		resultStats, err := simulate.ExecuteUseModel(simulate.RunSize_QuickDirty, model, &equipSet, nil, util.TrackProgress_Start())
		if err != nil {
			panic(err)
		}
		resultStats.Print(printer)
	}
}

func findSimpleUpgrade(printer *util.PrintRecorder) {
	model := model_factory.Model_PallyProtMitigation()
	gearFile := files.GearFileProtMitigation
	//model := model.Model_PallyProtCompromise()
	//gearFile := files.GearFileProtCompromise

	currentEquip := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(gearFile), &model, setup.MissingEnchant_Panic, printer)

	itemOptions := setup.OptionsSetup_FromGearFile(gearFile, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range mygear.SubstituteItemsProt {
		opts, example := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	common := commonComboCurrent()
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	addGearFileToCommon(common, files.GearFileProtDamage, &model, printer)
	addGearFileToCommon(common, files.GearFileProtBalanced, &model, printer)
	addGearFileToCommon(common, files.GearFileProtSurvival, &model, printer)
	addGearFileToCommon(common, files.GearFileProtMitigation, &model, printer)
	//restrictOptionsToCommon(common, &itemOptions)

	addItems := []items.ItemId{95038, 95003, 95022, 95002, 95023}
	for _, itemId := range addItems {
		opts, example := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}

	output := solver.Solver(
		&itemOptions,
		&model,
		printer,
		1,
		c_miscDefaultTimeout,
		nil)

	output.Report(printer)

	currentStats, err := simulate.ExecuteUseModel(simulate.RunSize_VerySlow, &model, &currentEquip, nil, util.TrackProgress_Start())
	if err != nil {
		panic(err)
	}
	resultStats, err := simulate.ExecuteUseModel(simulate.RunSize_VerySlow, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
	if err != nil {
		panic(err)
	}

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
	model := model_factory.Model_PallyProtSurvival()
	startGear := files.GearFileProtSurvival

	printer.Println("READ existing")
	currentEquip := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(startGear), &model, setup.MissingEnchant_Panic, printer)
	currentStats, err := simulate.ExecuteUseModel(simSize, &model, &currentEquip, nil, util.TrackProgress_Start())
	if err != nil {
		panic(err)
	}

	printer.Println("SETUP options")
	itemOptionsShared := setup.OptionsSetup_FromGearFile(startGear, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range mygear.SubstituteItemsProt {
		opts, example := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptionsShared.AddSeveralOptions(example.SlotItem(), opts)
	}

	common := commonComboCurrent()
	printer.Println("RESTRICT ret")
	addGearFileToCommon(common, files.GearFileRet, &model, printer)
	printer.Println("RESTRICT dps")
	addGearFileToCommon(common, files.GearFileProtDamage, &model, printer)
	printer.Println("RESTRICT mitset")
	addGearFileToCommon(common, files.GearFileProtMitigation, &model, printer)
	restrictOptionsToCommon(common, &itemOptionsShared)

	restrictSlotToId(&itemOptionsShared, items.Equip_Ring1, 96481)

	type pair struct {
		miti float64
		dps  float64
		name string
	}

	resultPairs := make([]pair, 0)

	checkItems := []items.ItemId{95141}
	for _, itemId := range checkItems {
		if currentEquip.IncludesItemId(itemId) {
			continue
		}

		_, example := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)

		for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
			if slotEquip == items.Equip_Ring1 {
				continue
			}

			printer.Printf("<<<<< FORCING %s %s >>>>>\n", slotEquip.Name(), example.BaseName())

			itemOptionsSpecific := itemOptionsShared.Clone()
			restrictSlotToId(&itemOptionsSpecific, slotEquip, itemId)

			output := solver.Solver(
				&itemOptionsSpecific,
				&model,
				printer,
				1,
				c_miscDefaultTimeout,
				nil)
			output.Report(printer)

			resultStats, err := simulate.ExecuteUseModel(simSize, &model, output.FullSet.Items(), nil, util.TrackProgress_Start())
			if err != nil {
				panic(err)
			}

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
		94519,                           // crit prim rage
		945190,                          // crit->master prim rage
		96793,                           // none fort zand
		96555,                           // none soul barrier
		94529,                           // none gaze twins
		94527,                           // ji-kun
		945270,                          // exp->crit ji-kun
		94507,                           // valor
		94508,                           // valor
		103989,                          // timeless alacrity of xuen
		103990,                          // timeless resolve of niuzao
		103678,                          // time lost artifict
		mygear.TrinketThokTailCelestial, // thok trinket
		mygear.TrinketFusionCoreHeroic,
		mygear.TrinketVialCorruptNormal,
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
			model_factory.Model_PallyProtSurvival(),
			files.GearFileProtSurvival,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation(),
			files.GearFileProtMitigation,
		}, {
			"compromise",
			model_factory.Model_PallyProtBalanced(),
			files.GearFileProtBalanced,
		}, {
			"dps",
			model_factory.Model_PallyProtDamage(),
			files.GearFileProtDamage,
		}, {
			"ret",
			model_factory.Model_PallyProtDamage(),
			files.GearFileProtDamage,
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
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
		printer.Println(group.label + " CURRENT")
		printer.Println(equipMap[items.Equip_Trinket1].CreateString())
		printer.Println(equipMap[items.Equip_Trinket2].CreateString())
	}

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())

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

			newEquip := equipMap
			newEquip[items.Equip_Trinket2] = item
			// fullSet := items.FullItemSet_FromMap(newEquip)

			resultStats, err := simulate.ExecuteSpecifyAll(simulate.RunSize_Largish, model.SimSpeedUp, model.Spec, model.Goal, fight, model.Professions, &newEquip, nil, util.TrackProgress_Nop())
			if err != nil {
				panic(err)
			}
			resultStats.Print(printer)
			for _, statType := range stats.SimTypeList {
				csv.AddFloat64(resultStats.GetFriendly(statType), 2)
			}

			csv.FinishColumn()
		}
	}

	csv.WriteTo(printer)
}

func trinketSimsBoth(printer *util.PrintRecorder) {
	// 103989, // timeless alacrity of xuen
	// 103990, // timeless resolve of niuzao
	// 94527,  // ji-kun

	itemIds := []items.ItemId{
		94519,                             // prim rage normal 2/2 536
		96793,                             // none fort zand
		94529,                             // none gaze twins normal 2/2
		103678,                            // time lost artifact 0/0
		mygear.TrinketZandSpark,           // heroic 2/2
		mygear.TrinketThokTailCelestial,   // 2/2
		mygear.TrinketFusionCoreHeroic,    // 2/2
		mygear.TrinketVialCorruptNormal,   // 2/2
		mygear.TrinketRookUnluckyHeroic,   // 2/2
		mygear.TrinketSkeerBloodCelestial, // 2/2
		mygear.TrinketJuggFocusCelestial,  //jugg focusing crystal (self heal) celestial, 0/2
		// evil eye 0/2
	}

	upLevel := func(id items.ItemId) int32 {
		return 2
	}
	//upLevel := func(id items.ItemId) int32 {
	//	var upgrade int32 = 0
	//	if id < 100000 || id == mygear.TrinketVialCorruptNormal ||
	//		id == mygear.TrinketThokTailCelestial ||
	//		id == mygear.TrinketRookUnluckyHeroic ||
	//		id == mygear.TrinketSkeerBloodCelestial ||
	//		id == mygear.TrinketFusionCoreHeroic {
	//		upgrade = 2
	//	}
	//	return upgrade
	//}

	fight := stats.Fight_Juggernaut_NoExternalHeal
	simRun := simulate.RunSize_Largish

	type groupStruct struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	groups := []groupStruct{
		{
			"heal",
			model_factory.Model_PallyProtHeal(),
			files.GearFileProtHeal,
		},
		{
			"survival",
			model_factory.Model_PallyProtSurvival(),
			files.GearFileProtSurvival,
		},
		{
			"mitigation",
			model_factory.Model_PallyProtMitigation(),
			files.GearFileProtMitigation,
		},
		{
			"balanced",
			model_factory.Model_PallyProtBalanced(),
			files.GearFileProtBalanced,
		},
		{
			"dps",
			model_factory.Model_PallyProtDamage(),
			files.GearFileProtDamage,
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
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
		printer.Println(group.label + " CURRENT")
		printer.Println(equipMap[items.Equip_Trinket1].CreateString())
		printer.Println(equipMap[items.Equip_Trinket2].CreateString())
	}

	for _, group := range groups {
		model := group.model
		file := group.file

		equipped := loaders.GearFileReader_Read(file)
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())

		processItemIds := itemIds
		if group.label == "ret" {
			processItemIds = util_collection.MapSliceAsNew_NoPointer(processItemIds, func(x items.ItemId) items.ItemId {
				if x == mygear.TrinketVialCorruptNormal {
					x = mygear.TrinketEyeGalakrasCelestial
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

				newEquip := equipMap
				newEquip[items.Equip_Trinket1] = itemOne
				newEquip[items.Equip_Trinket2] = itemTwo

				checkNotLow(newEquip)

				resultStats, err := simulate.ExecuteSpecifyAll(simRun, model.SimSpeedUp, model.Spec, model.Goal, fight, model.Professions, &newEquip, nil, util.TrackProgress_Nop())
				if err != nil {
					panic(err)
				}
				resultStats.Print(printer)
				for _, statType := range stats.SimTypeList {
					csv.AddFloat64(resultStats.GetFriendly(statType), 2)
				}

				csv.FinishColumn()
			}
		}
	}

	csv.WriteTo(printer)
}

func checkNotLow(equip items.FullEquipMap) {
	for x := range equip.AllItemSeq() {
		if x.ItemId() == 94519 {
			if x.ItemLevel() != 536 {
				panic("low")
			}
		}
	}

}

func currentSimGear(printer *util.PrintRecorder) {
	fight := stats.Fight_Juggernaut_NoExternalHeal
	goal := stats.OptimiseGoal_Mitigation
	simRun := simulate.RunSize_VerySlow

	type group struct {
		label string
		model gear_model.SpecModel
		file  string
	}

	groups := []group{

		{
			"survival",
			model_factory.Model_PallyProtSurvival(),
			files.GearFileProtSurvival,
		}, {
			"mitigation",
			model_factory.Model_PallyProtMitigation(),
			files.GearFileProtMitigation,
		},
		{
			"heal",
			model_factory.Model_PallyProtHeal(),
			files.GearFileProtHeal,
		},
		{
			"balance",
			model_factory.Model_PallyProtBalanced(),
			files.GearFileProtBalanced,
		}, {
			"dps",
			model_factory.Model_PallyProtDamage(),
			files.GearFileProtDamage,
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
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, &model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())

		resultStats, err := simulate.ExecuteSpecifyAll(simRun, model.SimSpeedUp, model.Spec, goal, fight, model.Professions, &equipMap, nil, util.TrackProgress_Nop())
		if err != nil {
			panic(err)
		}
		resultStats.Print(printer)
		for _, statType := range stats.SimTypeList {
			csv.AddFloat64(resultStats.GetFriendly(statType), 2)
		}

		csv.FinishColumn()
	}

	csv.WriteTo(printer)
}

func addGearFileToCommon(common map[items.ItemId]stats.ReforgeRecipe, gearFile string, model *gear_model.SpecModel, printer *util.PrintRecorder) {
	currentEquip := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(gearFile), model, setup.MissingEnchant_Panic, printer)
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
			model_factory.Model_PallyProtDamage(),
			files.GearFileProtDamage,
		}, {
			"compromise",
			model_factory.Model_PallyProtBalanced(),
			files.GearFileProtBalanced,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation(),
			files.GearFileProtMitigation,
		},
		{
			"with_set",
			model_factory.Model_PallyProtSurvival(),
			files.GearFileProtSurvival,
		},
	}

	weightType := weight_types.WeightType(1)

	for _, group := range groups {
		equipItems := loaders.GearFileReader_Read(group.file)
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipItems, &group.model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
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

	var prescaleTarget = 100000000.0

	groups := []group{
		{
			"ret",
			model_factory.Model_PallyRet(),
			files.GearFileRet,
		}, {
			"dps",
			model_factory.Model_PallyProtDamage(),
			files.GearFileProtDamage,
		}, {
			"compromise",
			model_factory.Model_PallyProtBalanced(),
			files.GearFileProtBalanced,
		}, {
			"no_set",
			model_factory.Model_PallyProtMitigation(),
			files.GearFileProtMitigation,
		},
		{
			"with_set",
			model_factory.Model_PallyProtSurvival(),
			files.GearFileProtSurvival,
		},
	}

	weightType := weight_types.WeightType(1)

	for _, group := range groups {
		equipItems := loaders.GearFileReader_Read(group.file)
		equipMap := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipItems, &group.model, setup.MissingEnchant_Panic, util.PrintRecorder_Nop())
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

func determineSetBonusValueBySim(printer *util.PrintRecorder) {
	//runSize := simulate.RunSize_QuickDirty
	//optionCount := 10
	runSize := simulate.RunSize_Largish
	//optionCount := 128
	optionCount := 64

	//goal02 := stats.OptimiseGoal_Mitigation
	goal02 := stats.OptimiseGoal_Dps
	goal4 := stats.OptimiseGoal_Dps
	//goal4 := stats.OptimiseGoal_HalfMitiHeal
	//fight := stats.Fight_Horridon_LowHeal
	fight := stats.Fight_Horridon_HighHeal
	//spec := stats.Spec_PaladinProt
	spec := stats.Spec_PaladinRet
	profession := gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true}

	//startGear := files.GearFileProtMitigationWithSet
	//modelEquipOnly := model_factory.Model_PallyProtMitigation_WithSet()
	//targetRatio := modelEquipOnly.SimPriority

	//gearFile := files.GearFileProtMitigationNoSet
	//model := model_factory.Model_PallyProtMitigation_NoSet()
	////model.BonusEnabled = bonus_set.SpecSetsEnableNamed("Plate of Winged Triumph")
	//model.BonusEnabled = bonus_set.SpecSetsEnableNamed("Plate of the Lightning Emperor")

	gearFile := files.GearFileRet
	model := model_factory.Model_PallyRet()
	model.BonusEnabled = bonus_set.SpecSetsEnableNamed(&model.SimPriority, "Battlegear of Winged Triumph", "Battlegear of the Lightning Emperor")
	model.BonusRequiredSolve = bonus_set.ItemCountsRequiredOptions{
		Mode:    bonus_set.CountMode_Exact,
		Options: []bonus_set.ItemCountsRequired{bonus_set.ItemCountsRequiredMake("Battlegear of Winged Triumph", 0, "Battlegear of the Lightning Emperor", 0)},
	}
	model.BonusRequiredWeight = nil

	initialSets, itemOptions := weightfind.GenerateRandomSets(gearFile, mygear.SubstituteItemsRet, &model, optionCount, printer, "", false)
	//initialSets, itemOptions := weightfind.GenerateRandomSets(gearFile, substituteItemsProt, &model, optionCount, printer, "")

	//T16 prot
	//setItems := []*items.FullItem{
	//	itemOptions.FindItemIdFirst(99126),
	//	itemOptions.FindItemIdFirst(99128),
	//	itemOptions.FindItemIdFirst(99129),
	//	itemOptions.FindItemIdFirst(99130),
	//}
	//T15 prot
	//setItems := []*items.FullItem{
	//	itemOptions.FindItemIdFirst(96668), // prot tier15 shoulder heroic
	//	itemOptions.FindItemIdFirst(96664), // prot tier15 chest heroic
	//	itemOptions.FindItemIdFirst(96666), // prot tier15 head heroic
	//	itemOptions.FindItemIdFirst(96667), // prot tier15 leg heroic, had removed in multi
	//}
	//T15 ret
	//setItems := []*items.FullItem{
	//	// not sure which heroics owned apart from this first
	//	setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults_NoForges(96658, 2, items.NO_RANDOM_SUFFIX, &model, printer),
	//	setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults_NoForges(96654, 2, items.NO_RANDOM_SUFFIX, &model, printer),
	//	setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults_NoForges(96655, 2, items.NO_RANDOM_SUFFIX, &model, printer),
	//	setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults_NoForges(96656, 2, items.NO_RANDOM_SUFFIX, &model, printer),
	//}
	//T16 ret
	setItems := []*items.FullItem{
		itemOptions.FindItemIdFirst(99052),
		itemOptions.FindItemIdFirst(99002),
		itemOptions.FindItemIdFirst(98986),
		itemOptions.FindItemIdFirst(98987),
	}
	preparedSetGroups := util_collection.MapSliceAsNew(initialSets, func(itemSet *items.FullItemSet) standardisedItemSetGroup {
		if model.BonusEnabled.CountInAnySet(itemSet.Items()) != 0 {
			panic("not zero")
		}
		return standardisedItemSetGroup{
			standardisedItemSet{*itemSet, stats.StatTypeMap[int32]{}},
			replaceWithEquivalentSetItems(itemSet, setItems, 2),
			replaceWithEquivalentSetItems(itemSet, setItems, 4),
		}
	})

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(preparedSetGroups) * 3)

	bonusData := util_async.Map_SliceToSlice(4, preparedSetGroups, func(group *standardisedItemSetGroup) bonuses {
		if model.BonusEnabled.CountInAnySet(group.zero.itemSet.Items()) != 0 {
			panic("not zero")
		}
		if model.BonusEnabled.CountInAnySet(group.two.itemSet.Items()) != 2 {
			panic("not two")
		}

		dataZero, err := simulate.ExecuteSpecifyAll(runSize, 0, spec, goal02, fight, profession,
			group.zero.itemSet.Items(), nil,
			tracker.NewChild())
		if err != nil {
			panic(err)
		}
		dataTwo, err := simulate.ExecuteSpecifyAll(runSize, 0, spec, goal02, fight, profession,
			group.two.itemSet.Items(), &group.two.bonusStats,
			tracker.NewChild())
		if err != nil {
			panic(err)
		}
		dataFour, err := simulate.ExecuteSpecifyAll(runSize, 0, spec, goal4, fight, profession,
			group.four.itemSet.Items(), &group.four.bonusStats,
			tracker.NewChild())
		if err != nil {
			panic(err)
		}

		printer.Printf("%s %s\n", dataZero.CompactStringGeneral(), dataTwo.CompactStringGeneral())

		return bonuses{
			simZero:  dataZero,
			simTwo:   dataTwo,
			twoDiff:  makeBonusDiff(dataZero, dataTwo),
			fourDiff: makeBonusDiff(dataTwo, dataFour),
		}
	})

	for simType := range stats.SimTypeEnum.ValueSeq() {
		average2 := util_collection.FindAverageFunc(bonusData, func(x bonuses) float64 {
			return x.twoDiff.GetOrPanic(simType)
		})
		printer.Printf("BONUS 2 %s %f\n", simType.Name(), average2)
	}
	for simType := range stats.SimTypeEnum.ValueSeq() {
		average4 := util_collection.FindAverageFunc(bonusData, func(x bonuses) float64 {
			return x.fourDiff.GetOrPanic(simType)
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
	simZero, simTwo, simFour stats.SimData
	twoDiff                  util_collection.EnumMap[stats.SimType, float64]
	fourDiff                 util_collection.EnumMap[stats.SimType, float64]
}

func replaceWithEquivalentSetItems(baseSet *items.FullItemSet, bonusItems []*items.FullItem, bonusCount int) standardisedItemSet {
	substitutedEquip := *baseSet.Items()
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

func determineBestUseOfGearSets(printer *util.PrintRecorder) {
	//runSize := simulate.RunSize_QuickDirty
	//optionCount := 10
	//runSize := simulate.RunSize_QuickDirty
	runSize := simulate.RunSize_VerySlow
	//optionCount := 128
	//optionCount := 50

	goal := stats.OptimiseGoal_HalfMitiHeal
	//goal := stats.OptimiseGoal_Mitigation
	fight := stats.Fight_Juggernaut_NoExternalHeal
	spec := stats.Spec_PaladinProt
	profession := gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true}
	substitutes := mygear.SubstituteItemsProt

	gearFile := files.GearFileProtMitigation
	model := model_factory.Model_PallyProtMitigation()

	setNameT15Prot := "Plate of the Lightning Emperor"
	setNameT16Prot := "Plate of Winged Triumph"
	setNameT15Ret := "Battlegear of the Lightning Emperor"
	setNameT16Ret := "Battlegear of Winged Triumph"
	model.BonusEnabled = bonus_set.SpecSetsEnableNamed(&model.SimPriority, setNameT15Prot, setNameT16Prot, setNameT15Ret, setNameT16Ret)
	model.StatWeights = tools.StatRatingsWeights_ReadFile(files.WeightProtHeal)

	bonusCombos := []bonus_set.ItemCountsRequired{
		bonus_set.ItemCountsRequiredMake(setNameT16Prot, 0, setNameT15Prot, 0),
		bonus_set.ItemCountsRequiredMake(setNameT16Prot, 0, setNameT15Prot, 2),
		bonus_set.ItemCountsRequiredMake(setNameT16Prot, 2, setNameT15Prot, 0),
		bonus_set.ItemCountsRequiredMake(setNameT16Prot, 4),
		bonus_set.ItemCountsRequiredMake(setNameT15Prot, 2, setNameT16Prot, 2),
	}
	comboNames := []string{
		"Zero",
		"Prot15_2pcOnly",
		"Prot16_2pcOnly",
		"Prot16_4pc",
		"Prot_2pcEach",
	}

	itemOptions := setup.OptionsSetup_FromGearFile(gearFile, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substitutes {
		opts, example := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &model, printer)
		itemOptions.AddSeveralOptions(example.SlotItem(), opts)
	}
	itemOptions.RemoveDuplicates()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(bonusCombos))

	noteList := make([]string, 0)
	for n, comboSet := range bonusCombos {
		comboModel := model.CloneShallow()
		comboModel.BonusRequiredSolve = bonus_set.ItemCountsRequiredOptions{
			Mode:    bonus_set.CountMode_AllowPlusOne,
			Options: []bonus_set.ItemCountsRequired{comboSet},
		}
		solveOutput := solver.Solver(
			&itemOptions,
			comboModel,
			printer,
			1,
			c_miscDefaultTimeout,
			nil,
		)
		if !solveOutput.Success {
			panic(solveOutput.Error)
		}
		gear := solveOutput.FullSet.Items()
		bonusText := bonus_set.AllBonusesText(gear)
		simData, err := simulate.ExecuteSpecifyAll(runSize, 0, spec, goal, fight, profession,
			gear, nil,
			tracker.NewChild())
		if err != nil {
			panic(err)
		}

		note := fmt.Sprintf("COMBO=%s %s %s", comboNames[n], simData.CompactStringGeneral(), bonusText)
		noteList = append(noteList, note)
		printer.Println0()
	}
	printer.Println0()
	printer.Println0()
	for _, note := range noteList {
		printer.Println(note)
	}
}
