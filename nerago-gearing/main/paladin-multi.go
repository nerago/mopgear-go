package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model/model_factory"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
	"slices"
)

const (
	trinketZandSpark = 96398
	trinketFortZand  = 96793
	trinketPrimRage  = 94519
	trinketTwinsGaze = 94529

	trinketFusionCoreCelestial  = 104961
	trinketThokTailCelestial    = 105111
	trinketVialCorruptNormal    = 102306
	trinketRookUnluckyNormal    = 102296
	trinketEyeGalakrasCelestial = 104993
	trinketSkeerBloodCelestial  = 105134

	legendMeleeCloak = 102249
	legendTankCloak  = 102250
)

// TIER
var retT15 = []items.ItemId{
	95282, // ret tier15 normal head
	96658, // ret tier15 shoulder heroic
}
var retT16 = []items.ItemId{
	99052, // ret t16 chest celestial
	99002, // ret t16 hand celestial
	98985, // ret t16 head celestial
	98986, // ret t16 legs celestial
	98987, // ret t16 shoulder celestial
	99139, // ret t16 legs normal
}
var protT15 = []items.ItemId{
	95291, // prot tier15 hand normal
	96664, // prot tier15 chest heroic
	96666, // prot tier15 head heroic
	96667, // prot tier15 leg heroic
	96668, // prot tier15 shoulder heroic
}
var protT16 = []items.ItemId{
	99126, // prot t16 chest normal
	99128, // prot t16 head normal
	99129, // prot t16 legs normal
	99130, // prot t16 shoulder normal
	99026, // prot t16 legs celestial
	99027, // prot t16 shoulder celestial
	99028, // prot t16 hand celestial
}

// TRINKET
var trinketsDpsP3 = []items.ItemId{
	trinketZandSpark,
}
var trinketsTankP3 = []items.ItemId{
	trinketFortZand,
}
var newTrinketsDamage = []items.ItemId{
	trinketThokTailCelestial,
	trinketFusionCoreCelestial,
	trinketSkeerBloodCelestial,
}
var newTrinketsTank = []items.ItemId{
	trinketVialCorruptNormal,
	trinketRookUnluckyNormal,
}

// REMAINING P3
var miscOtherP3 = []items.ItemId{
	96420, // talisman of angry spirits
	96373, // cloudbreaker belt heroic
	96542, // tidal force treads
	96500, // scaled tyrant heroic
}

// ORGRIMMAR
var timeless = []items.ItemId{
	101882, // cliffbreaker helm exp/mastery
	101887, // timeless ring haste/mastery. Cliffbreaker Seal of the Faultline. 549 (is upgraded)
	// Cliffbreaker Seal of the Landslide. hit/expertise. 535 (not upgraded)
	101947, //  Elder Tortoiseshell Seal of the Mountainbed. 549 (is upgraded)
}
var celestial = []items.ItemId{
	105011, // Demolisher's Reinforced Belt
}
var celestialRaden = []items.ItemId{
	95011, // lighting clawfeet
	95022, // Ra-den's Ruinous Ring
}
var orgRaidDrops = []items.ItemId{
	103787, // poisonbinder girth
	103738, // bubble bracers
	105785, // vanguard burly bracer
	103734, // zoid gauntlets
	103735, // tar-coated gauntlets
	103916, // jugg ignition keys
	104461, // rage-blind greathelm
	104415, // bubble bracer heroic
	103892, // tharnok helm
	105767, // hoodrych chest ordos
	104417, // corruption-rotted gauntlets
	104416, // chest congealed corruption heroic
	103796, // seal kings norm
	103798, // bloodclaw band
	105761, // Partik's Purified Legplates
}
var orgOneHandAndShield = []items.ItemId{
	103826, // xifeng weapon
	103872, // bulwurk of fallen general
	103871, // tower shield
	104485, // shield of mockery
	103972, // kilruk sword
	104464, // xifeng heroic
	104560, // bulwurk of fallen general heroic
}

var legendCloaks = []items.ItemId{legendTankCloak, legendMeleeCloak}

func PaladinMultiRun() {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "multi-set")

	simSize := simulate.RunSize_Largish
	//simSize := simulate.RunSize_Common
	//simSize := simulate.RunSize_QuickDirty

	job := multi_types.JobInputs{}
	job.SetMinimumExtraItemLevel(463)
	job.SetTimeLimitEachSolver(8000)
	job.SetSimSize(simSize)
	job.SetReforgingAllowNonCommon(true)
	//job.SetWriteBestToGearFiles()
	//job.SetWeightTypes(1)
	//job.SetWeightTypes(1, 2, 3)
	job.SetWeightTypes(1, 2)

	var generalUpgrade items.UpgradeLevel = 0
	var forceUpgrade items.UpgradeLevel = 0

	ret := multi_types.SpecParam{
		Label: "Ret",
		Model: model_factory.Model_PallyRet(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                     files.GearFileRet,
			RequestRatingPercent:         0.04,
			ExtraUpgradeLevel:            generalUpgrade,
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
			ExtraUpgradeLevel:            generalUpgrade,
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
			ExtraUpgradeLevel:            generalUpgrade,
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
			ExtraUpgradeLevel:            generalUpgrade,
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
			ExtraUpgradeLevel:            generalUpgrade,
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
			ExtraUpgradeLevel:            generalUpgrade,
			ForceUpgradeExistingItems:    forceUpgrade,
			MissingEnchant:               setup.MissingEnchant_Panic,
			ExpectAllBonusItemsAvailable: true,
		},
	}

	ret.AddExtraItem(legendMeleeCloak)
	addExtrasToEach(legendCloaks, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(retT15, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(retT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(protT15, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(protT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(trinketsDpsP3, &ret, &protDps)
	addExtrasToEach(trinketsDpsP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(trinketsTankP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(miscOtherP3, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	newStuffP5 := slices.Concat(timeless, celestial, celestialRaden, orgRaidDrops, newTrinketsDamage)
	addExtrasToEach(newStuffP5, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(newTrinketsTank, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(orgOneHandAndShield, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

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
	ret.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	ret.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	ret.ForceSingleSlot(items.Equip_Trinket2, trinketEyeGalakrasCelestial)
	protDps.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	protDps.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protDps.ForceSingleSlot(items.Equip_Trinket2, trinketSkeerBloodCelestial)
	protBalanced.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	protBalanced.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protBalanced.ForceSingleSlot(items.Equip_Trinket2, trinketSkeerBloodCelestial)
	protBalanced.AddReportVariant(items.Equip_Trinket2, trinketVialCorruptNormal)
	protMitigation.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigation.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protMitigation.ForceSingleSlot(items.Equip_Trinket2, trinketZandSpark)
	protMitigation.AddReportVariant(items.Equip_Trinket1, trinketThokTailCelestial)
	protMitigation.AddReportVariant(items.Equip_Trinket2, trinketSkeerBloodCelestial)
	protSurvival.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protSurvival.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protSurvival.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)
	protSurvival.AddReportVariant(items.Equip_Trinket1, trinketSkeerBloodCelestial)
	protSurvival.AddReportVariant(items.Equip_Trinket2, trinketThokTailCelestial)
	protHeal.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protHeal.ForceSingleSlot(items.Equip_Trinket1, trinketSkeerBloodCelestial)
	protHeal.ForceSingleSlot(items.Equip_Trinket2, trinketZandSpark)

	// HELMET
	blockHelmetsWithoutCapacitance(&ret)
	blockHelmetsWithoutCapacitance(&protDps)
	blockHelmetsWithoutCapacitance(&protBalanced)
	blockHelmetsWithoutIndomitable(&protMitigation)
	blockHelmetsWithoutIndomitable(&protSurvival)
	blockHelmetsWithoutIndomitable(&protHeal)

	//job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Haste, 320))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Stamina, 120))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Strength, 80, stats.Stat_Haste, 160))
	//job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Strength, 160))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Expertise, 160, stats.Stat_Hit, 160))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Hit, 160))
	//job.ActivateAlternateGemmingAsPermute()

	//job.MakeRandomVariants(101887, 0, -365, -352)

	//ret.AddBagsExtra()
	//protDps.AddBagsExtra()
	//protCompromise.AddBagsExtra()
	//protMitigationNoSet.AddBagsExtra()
	//protMitigationWithSet.AddBagsExtra()
	//protHeal.AddBagsExtra()

	//addExtrasFromFinder(loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic),
	//	&ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protBalanced)
	job.AddSetParam(protMitigation)
	job.AddSetParam(protSurvival)
	job.AddSetParam(protHeal)

	//job.AddAlternateUpgradeChoices(
	//	99028,  // Handguards of Winged Triumph celestial
	//	103972, // kilruk sword - still passable option #4 for most sets, upgrade rank #3 of remaining
	//	105122, // Asgorathian Blood Seal - option #5, upgrade rank #4
	//)
	//job.AddAlternateUpgradeChoices(105033) // Wolf-Rider Spurs

	//job.EnablePermuteOnItemCountOptions()

	job.VerifyNoExtraDuplicates()
	//job.RemoveAnyExtraDuplicates()

	run := multi.JobCreate(printer, job)

	run.RunNoPermutations_AllCommonAlternates(true, true)
	//run.RunNoPermutations_BestOnly(true, true)
	//run.RunForSolutionsPerPermute(5, true)

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
