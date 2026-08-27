package main

import (
	"slices"

	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/multi"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func paladinMultiRun(printer *util.PrintRecorder) {
	job := multi_types.JobInputs{}
	job.SetMinimumExtraItemLevel(463)
	job.SetTimeLimitEachSolver(200)
	job.SetSimSize(simulate.RunSize_Largish)
	//simSize := simulate.RunSize_Common
	//simSize := simulate.RunSize_QuickDirty
	job.SetWriteBestToGearFiles(true)

	var extraUpgrade items.UpgradeLevel = 2
	var forceUpgrade items.UpgradeLevel = 0

	ret := multi_types.SpecParam{
		Label: "Ret",
		Model: model_factory.Model_PallyRet(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileRet,
			RequestRatingPercent:         0.04,
			ExtraUpgradeLevel:            extraUpgrade,
			ForceUpgradeExistingItems:    0,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: true,
		},
	}
	protDps := multi_types.SpecParam{
		Label: "Prot-Damage",
		Model: model_factory.Model_PallyProtDamage(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileProtDamage,
			RequestRatingPercent:         0.01,
			ExtraUpgradeLevel:            extraUpgrade,
			ForceUpgradeExistingItems:    forceUpgrade,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: false,
		},
	}
	protBalanced := multi_types.SpecParam{
		Label: "Prot-Balanced",
		Model: model_factory.Model_PallyProtBalanced(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileProtBalanced,
			RequestRatingPercent:         0.25,
			ExtraUpgradeLevel:            extraUpgrade,
			ForceUpgradeExistingItems:    forceUpgrade,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: false,
		},
	}
	protMitigation := multi_types.SpecParam{
		Label: "Prot-Mitigation",
		Model: model_factory.Model_PallyProtMitigation(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileProtMitigation,
			RequestRatingPercent:         0.40,
			ExtraUpgradeLevel:            extraUpgrade,
			ForceUpgradeExistingItems:    forceUpgrade,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: true,
		},
	}
	protSurvival := multi_types.SpecParam{
		Label: "Prot-Survival",
		Model: model_factory.Model_PallyProtSurvival(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileProtSurvival,
			RequestRatingPercent:         0.25,
			ExtraUpgradeLevel:            extraUpgrade,
			ForceUpgradeExistingItems:    forceUpgrade,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: true,
		},
	}
	protHeal := multi_types.SpecParam{
		Label: "Prot-Heal",
		Model: model_factory.Model_PallyProtHeal(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileProtHeal,
			RequestRatingPercent:         0.05,
			ExtraUpgradeLevel:            extraUpgrade,
			ForceUpgradeExistingItems:    forceUpgrade,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: true,
		},
	}

	ret.AddExtraItem(mygear.LegendMeleeCloak)
	addExtrasToEach(mygear.LegendCloaks, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(mygear.RetT15, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(mygear.RetT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(mygear.ProtT15, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(mygear.ProtT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(mygear.TrinketsDpsP3, &ret, &protDps)
	addExtrasToEach(mygear.TrinketsDpsP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(mygear.TrinketsTankP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(mygear.MiscOtherP3, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	newStuffP5 := slices.Concat(mygear.Timeless, mygear.Celestial, mygear.CelestialRaden, mygear.OrgRaidDrops, mygear.NewTrinketsDamage)
	addExtrasToEach(newStuffP5, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(mygear.NewTrinketsTank, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(mygear.OrgOneHandAndShield, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	ret.AddExtraItems([]items.ItemId{
		95281,  // ret tier15 gloves normal
		96657,  // ret tier15 legs heroic
		95535,  // normal lightning legs
		104993, // Evil Eye of Galakras trinket celestial
		105033, // Wolf-Rider Spurs

		103968, // britomart pike
		105122, // Asgorathian Blood Seal
	})

	protDps.AddExtraItems([]items.ItemId{
		96657,  // ret tier15 legs heroic
		105033, // Wolf-Rider Spurs
		94776,  // primal turtle amulet
		96395,  // bloodsoaked legplates
		96668,  // prot tier15 shoulder heroic
		96376,  // worldbreaker weapon
	})

	protBalanced.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		105033, // Wolf-Rider Spurs
		105122, // Asgorathian Blood Seal
		94776,  // primal turtle amulet
		96395,  // bloodsoaked legplates
		96534,  // qon's scimitar
	})

	protMitigation.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		105033, // Wolf-Rider Spurs
		96481,  // durumu tentacle heroic
		95020,  // ra-den contemplative loop
		96534,  // qon's scimitar
		95205,  // terra-cotta neck
	})

	protSurvival.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		105033, // Wolf-Rider Spurs
		95020,  // ra-den contemplative loop
		96534,  // qon's scimitar
		95205,  // terra-cotta neck
	})

	protHeal.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		105033, // Wolf-Rider Spurs
		105122, // Asgorathian Blood Seal
		96534,  // qon's scimitar
		95205,  // terra-cotta neck
	})

	// predetermined choices
	ret.ForceSingleSlot(items.Equip_Weapon, 103968) // britomark
	ret.ForceSingleSlot(items.Equip_Back, mygear.LegendMeleeCloak)
	ret.ForceSingleSlot(items.Equip_Trinket1, mygear.TrinketThokTailCelestial)
	ret.ForceSingleSlot(items.Equip_Trinket2, mygear.TrinketEyeGalakrasCelestial)
	protDps.ForceSingleSlot(items.Equip_Back, mygear.LegendMeleeCloak)
	protDps.ForceSingleSlot(items.Equip_Trinket1, mygear.TrinketThokTailCelestial)
	protDps.ForceSingleSlot(items.Equip_Trinket2, mygear.TrinketSkeerBloodCelestial)
	protBalanced.ForceSingleSlot(items.Equip_Back, mygear.LegendMeleeCloak)
	protBalanced.ForceSingleSlot(items.Equip_Trinket1, mygear.TrinketThokTailCelestial)
	protBalanced.ForceSingleSlot(items.Equip_Trinket2, mygear.TrinketSkeerBloodCelestial)
	protBalanced.AddReportVariant(items.Equip_Trinket2, mygear.TrinketVialCorruptNormal)
	protMitigation.ForceSingleSlot(items.Equip_Back, mygear.LegendTankCloak)
	protMitigation.ForceSingleSlot(items.Equip_Trinket1, mygear.TrinketThokTailCelestial)
	protMitigation.ForceSingleSlot(items.Equip_Trinket2, mygear.TrinketZandSpark)
	protMitigation.AddReportVariant(items.Equip_Trinket1, mygear.TrinketThokTailCelestial)
	protMitigation.AddReportVariant(items.Equip_Trinket2, mygear.TrinketSkeerBloodCelestial)
	protSurvival.ForceSingleSlot(items.Equip_Back, mygear.LegendTankCloak)
	protSurvival.ForceSingleSlot(items.Equip_Trinket1, mygear.TrinketThokTailCelestial)
	protSurvival.ForceSingleSlot(items.Equip_Trinket2, mygear.TrinketFortZand)
	protSurvival.AddReportVariant(items.Equip_Trinket1, mygear.TrinketSkeerBloodCelestial)
	protSurvival.AddReportVariant(items.Equip_Trinket2, mygear.TrinketThokTailCelestial)
	protHeal.ForceSingleSlot(items.Equip_Back, mygear.LegendTankCloak)
	protHeal.ForceSingleSlot(items.Equip_Trinket1, mygear.TrinketSkeerBloodCelestial)
	protHeal.ForceSingleSlot(items.Equip_Trinket2, mygear.TrinketZandSpark)

	// HELMET
	blockHelmetsWithoutCapacitance(&ret)
	blockHelmetsWithoutCapacitance(&protDps)
	blockHelmetsWithoutCapacitance(&protBalanced)
	blockHelmetsWithoutIndomitable(&protMitigation)
	blockHelmetsWithoutIndomitable(&protSurvival)
	blockHelmetsWithoutIndomitable(&protHeal)

	//job.MakeRandomVariants(101887, 0, -365, -352)

	//ret.AddBagsExtra()
	//protDps.AddBagsExtra()
	//protBalanced.AddBagsExtra()
	//protMitigation.AddBagsExtra()
	//protSurvival.AddBagsExtra()
	//protHeal.AddBagsExtra()

	//addExtrasFromFinder(loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic),
	//	&ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protBalanced)
	job.AddSetParam(protMitigation)
	job.AddSetParam(protSurvival)
	job.AddSetParam(protHeal)

	job.VerifyNoExtraDuplicates()
	//job.RemoveAnyExtraDuplicates()

	if false {
		taskQuick := multi_types.JobInputTask{
			AlsoExistingEquipped:  true,
			AlsoSpecOptimums:      true,
			Alternates:            multi_types.AlternateModeNone,
			AlternatesLimit:       util_collection.Optional_Empty[int](),
			IncludeInterimResults: false,
			WeightTypeList:        []weight_types.WeightType{1},
			//WeightTypeList:          []weight_types.WeightType{3},
			RunDecimate:             false,
			ReforgingAllowNonCommon: false,
		}
		multi.JobCreate(printer, job, taskQuick).Run()
	} else if true {
		weightTypes := []weight_types.WeightType{2}

		// all standard and alternates. plus optional regem
		task1 := multi_types.JobInputTask{
			AlsoExistingEquipped:    false,
			AlsoSpecOptimums:        true,
			Alternates:              multi_types.AlternateModeReforgeBlocks,
			AlternatesLimit:         util_collection.Optional_OfValue(5),
			IncludeInterimResults:   false,
			WeightTypeList:          weightTypes,
			RunDecimate:             false,
			ReforgingAllowNonCommon: false,
		}
		//task1.AddAlternateGem(stats.StatBlock_of(stats.Stat_Haste, 320))
		//task1.Permute.AlternateGemsEnableAsPermute = true
		//task1.Permute.PermuteOnItemCountOptions = true

		// interim results for main solve only
		task2 := multi_types.JobInputTask{
			AlsoExistingEquipped:    false,
			AlsoSpecOptimums:        false,
			Alternates:              multi_types.AlternateModeNone,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   true,
			WeightTypeList:          weightTypes,
			RunDecimate:             false,
			ReforgingAllowNonCommon: false,
		}

		_ = task1
		_ = task2

		run := multi.JobCreate(printer, job, task1, task2)
		run.Run()
	} else if true {
		weightTypes := []weight_types.WeightType{1}

		// all standard and alternates. plus optional regem
		task1 := multi_types.JobInputTask{
			AlsoExistingEquipped:    true,
			AlsoSpecOptimums:        true,
			Alternates:              multi_types.AlternateModeItemAndReforgeBlocks,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   false,
			WeightTypeList:          weightTypes,
			RunDecimate:             false,
			ReforgingAllowNonCommon: true,
		}
		task1.AddAlternateGem(stats.StatBlock_of(stats.Stat_Haste, 320))
		task1.Permute.AlternateGemsEnableAsPermute = true
		task1.Permute.PermuteOnItemCountOptions = true

		// interim results for main solve only
		task2 := multi_types.JobInputTask{
			AlsoExistingEquipped:    false,
			AlsoSpecOptimums:        false,
			Alternates:              multi_types.AlternateModeNone,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   true,
			WeightTypeList:          weightTypes,
			RunDecimate:             false,
			ReforgingAllowNonCommon: true,
		}

		// weight 3 with optimums
		task3 := multi_types.JobInputTask{
			AlsoExistingEquipped:    false,
			AlsoSpecOptimums:        true,
			Alternates:              multi_types.AlternateModeNone,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   false,
			WeightTypeList:          []weight_types.WeightType{3},
			RunDecimate:             false,
			ReforgingAllowNonCommon: false,
		}

		_ = task1
		_ = task2
		_ = task3

		//run := multi.JobCreate(printer, job, task1, task2)
		run := multi.JobCreate(printer, job, task3)
		//run := multi.JobCreate(printer, job, task1, task2, task3)
		run.Run()
	} else {
		task1 := multi_types.JobInputTask{
			AlsoExistingEquipped:    true,
			AlsoSpecOptimums:        true,
			Alternates:              multi_types.AlternateModeItemAndReforgeBlocks,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   false,
			WeightTypeList:          []weight_types.WeightType{1, 2},
			RunDecimate:             false,
			ReforgingAllowNonCommon: false,
		}
		task1.AddAlternateGem(stats.StatBlock_of(stats.Stat_Haste, 320))

		task2 := multi_types.JobInputTask{
			AlsoExistingEquipped:    false,
			AlsoSpecOptimums:        false,
			Alternates:              multi_types.AlternateModeNone,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   true,
			WeightTypeList:          []weight_types.WeightType{1, 2},
			RunDecimate:             true,
			ReforgingAllowNonCommon: true,
		}

		task3 := multi_types.JobInputTask{
			AlsoExistingEquipped:    false,
			AlsoSpecOptimums:        true,
			Alternates:              multi_types.AlternateModeNone,
			AlternatesLimit:         util_collection.Optional_Empty[int](),
			IncludeInterimResults:   false,
			WeightTypeList:          []weight_types.WeightType{3},
			RunDecimate:             true,
			ReforgingAllowNonCommon: false,
		}

		//task1.AddAlternateGem(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Stamina, 120))
		//task1.AddAlternateGem(stats.StatBlock_of2(stats.Stat_Strength, 80, stats.Stat_Haste, 160))
		//task1.AddAlternateGem(stats.StatBlock_of(stats.Stat_Strength, 160))
		//task1.AddAlternateGem(stats.StatBlock_of2(stats.Stat_Expertise, 160, stats.Stat_Hit, 160))
		//task1.AddAlternateGem(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Hit, 160))
		//task1.Permute.AlternateGemsEnableAsPermute = true

		//task1.AddAlternateUpgradeChoices(
		//	99028,  // Handguards of Winged Triumph celestial
		//	103972, // kilruk sword - still passable option #4 for most sets, upgrade rank #3 of remaining
		//	105122, // Asgorathian Blood Seal - option #5, upgrade rank #4
		//)
		//task1.AddAlternateUpgradeChoices(104494) // krugruk shoulderplates

		//task1.EnablePermuteOnItemCountOptions()

		run := multi.JobCreate(printer, job, task1, task2, task3)
		run.Run()
	}

	//job.CullingReport()
	//run.RunCullingSets(400, time.Minute*45)
}

func addExtrasToEach(itemIdList []items.ItemId, params ...*multi_types.SpecParam) {
	for _, param := range params {
		param.AddExtraItems(itemIdList)
	}
}

func blockHelmetsWithoutCapacitance(param *multi_types.SpecParam) {
	param.BlockItem(87101)  // white tiger helmet = prot gem
	param.BlockItem(95292)  // lightning emp faceguard = prot gem
	param.BlockItem(96666)  // lightning emp faceguard heroic = prot gem
	param.BlockItem(96550)  // doomed crown heroic = prot gem
	param.BlockItem(99128)  // winged faceguard = prot gem
	param.BlockItem(104461) // rage-blind = prot gem

	param.BlockItem(99370)  // Faceguard of Winged Triumph
	param.BlockItem(105805) // Dominik's Casque of Raging Flame

	blockGeneral(param)
}

func blockHelmetsWithoutIndomitable(param *multi_types.SpecParam) {
	param.BlockItem(87024)  // nullification greathelm = capacitance
	param.BlockItem(95282)  // lightning emp helmet = capacitance
	param.BlockItem(101882) // cliffbreaker helm = capacitance
	param.BlockItem(98985)  // ret helm = capacitance
	param.BlockItem(103892) // thranok = capacitance
	blockGeneral(param)
}

func blockGeneral(param *multi_types.SpecParam) {
	param.BlockItem(95513)  // normal ring
	param.BlockItem(95778)  // golden golem celestial
	param.BlockItem(101942) // Elder Tortoiseshell Helm (just blocking to gem bis runs done, don't have one)
}

func addExtrasFromFinder(foundList []loaders.ItemFoundRef, params ...*multi_types.SpecParam) {
	for _, itemRef := range foundList {
		for _, param := range params {
			param.AddExtraItem(itemRef.ItemId)
		}
	}
}

// as part of BIS calcs try each meta as needed
//capacitanceSets := []multi_types.MultiSetParam{ret, protDps, protCompromise}
//indomitableSets := []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet, protHeal}
//job.AddItemDistinctUsageGroups(104590, false, capacitanceSets, indomitableSets)
//job.AddItemDistinctUsageGroups(104647, false, capacitanceSets, indomitableSets)
//job.AddItemDistinctUsageGroups(105807, false, capacitanceSets, indomitableSets)
//job.AddItemDistinctUsageGroups(99379, false, capacitanceSets, indomitableSets)
//job.AddItemDistinctUsageGroups(104492, false, capacitanceSets, indomitableSets)

//job.AddItemDistinctUsageGroups(103892, true, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet, protHeal})
