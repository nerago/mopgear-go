package main

import (
	"slices"

	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func paladinMultiRunLite(printer *util.PrintRecorder) {
	// TIER
	var retT16 = []items.ItemId{
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
		99028, // prot t16 hand celestial
	}

	// TRINKET
	var trinketsDpsP3 = []items.ItemId{
		mygear.TrinketZandSpark,
	}
	var trinketsTankP3 = []items.ItemId{
		mygear.TrinketFortZand,
	}
	var newTrinketsDamage = []items.ItemId{
		mygear.TrinketThokTailCelestial,
		mygear.TrinketFusionCoreHeroic,
		mygear.TrinketSkeerBloodCelestial,
	}
	var newTrinketsTank = []items.ItemId{
		mygear.TrinketVialCorruptNormal,
		mygear.TrinketRookUnluckyNormal,
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
		101887, // timeless ring haste/mastery. Cliffbreaker Seal of the Faultline. 549 (is upgraded)
		// Cliffbreaker Seal of the Landslide. hit/expertise. 535 (not upgraded)
	}
	var celestialRaden = []items.ItemId{
		95011, // lighting clawfeet
		95022, // Ra-den's Ruinous Ring
	}
	var orgRaidDrops = []items.ItemId{
		103787, // poisonbinder girth
		105785, // vanguard burly bracer
		103916, // jugg ignition keys
		104461, // rage-blind greathelm
		104415, // bubble bracer heroic
		103892, // tharnok helm
		104417, // corruption-rotted gauntlets
		104416, // chest congealed corruption heroic
		103796, // seal kings norm
		103798, // bloodclaw band
		105761, // Partik's Purified Legplates
	}
	var orgOneHandAndShield = []items.ItemId{
		104485, // shield of mockery
		104464, // xifeng heroic
		104560, // bulwurk of fallen general heroic
	}

	var legendCloaks = []items.ItemId{mygear.LegendTankCloak, mygear.LegendMeleeCloak}

	simSize := simulate.RunSize_Largish
	//simSize := simulate.RunSize_Common
	//simSize := simulate.RunSize_QuickDirty

	job := multi_types.JobInputs{}
	job.SetMinimumExtraItemLevel(463)
	job.SetTimeLimitEachSolver(2500)
	job.SetSimSize(simSize)

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
			RequestRatingPercent:         0.35,
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
			RequestRatingPercent:         0.30,
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

	ret.AddExtraItem(mygear.LegendMeleeCloak)
	addExtrasToEach(legendCloaks, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(retT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(protT15, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(protT16, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(trinketsDpsP3, &ret, &protDps)
	addExtrasToEach(trinketsDpsP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(trinketsTankP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(miscOtherP3, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	newStuffP5 := slices.Concat(timeless, celestialRaden, orgRaidDrops, newTrinketsDamage)
	addExtrasToEach(newStuffP5, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(newTrinketsTank, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(orgOneHandAndShield, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	ret.AddExtraItems([]items.ItemId{
		95281,  // ret tier15 gloves normal
		96657,  // ret tier15 legs heroic
		95282,  // ret tier15 normal head
		96658,  // ret tier15 shoulder heroic
		95910,  // ret tier15 chest celestial
		95535,  // normal lightning legs
		104993, // Evil Eye of Galakras trinket celestial
		105033, // Wolf-Rider Spurs
		103968, // britomart pike
		105122, // Asgorathian Blood Seal
		105767, // hoodrych chest ordos
		99002,  // ret t16 hand celestial
		99052,  // ret t16 chest celestial
		98985,  // ret t16 head celestial
		98987,  // ret t16 shoulder celestial
		103735, // tar-coated gauntlets
	})

	protDps.AddExtraItems([]items.ItemId{
		96657,  // ret tier15 legs heroic
		105033, // Wolf-Rider Spurs
		96668,  // prot tier15 shoulder heroic
		101882, // cliffbreaker helm exp/mastery
		96658,  // ret tier15 shoulder heroic
		105767, // hoodrych chest ordos
		103735, // tar-coated gauntlets
	})

	protBalanced.AddExtraItems([]items.ItemId{
		96534,  // qon's scimitar
		101882, // cliffbreaker helm exp/mastery
		96658,  // ret tier15 shoulder heroic
		101947, // Elder Tortoiseshell Seal of the Mountainbed
		103735, // tar-coated gauntlets
	})

	protMitigation.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		96481,  // durumu tentacle heroic
		95020,  // ra-den contemplative loop
		96534,  // qon's scimitar
		95205,  // terra-cotta neck
		103734, // zoid gauntlets
	})

	protSurvival.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		95020,  // ra-den contemplative loop
		96534,  // qon's scimitar
		95205,  // terra-cotta neck
		101947, // Elder Tortoiseshell Seal of the Mountainbed
	})

	protHeal.AddExtraItems([]items.ItemId{
		105033, // Wolf-Rider Spurs
		96534,  // qon's scimitar
		95205,  // terra-cotta neck
		101947, // Elder Tortoiseshell Seal of the Mountainbed
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

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protBalanced)
	job.AddSetParam(protMitigation)
	job.AddSetParam(protSurvival)
	job.AddSetParam(protHeal)

	//job.AddAlternateUpgradeChoices(105033) // Wolf-Rider Spurs

	job.VerifyNoExtraDuplicates()
	//job.RemoveAnyExtraDuplicates()

	task := multi_types.JobInputTask{
		WeightTypeList:          []weight_types.WeightType{3},
		AlsoExistingEquipped:    false,
		AlsoSpecOptimums:        false,
		Alternates:              multi_types.AlternateModeNone,
		AlternatesLimit:         util_collection.Optional[int]{},
		IncludeInterimResults:   false,
		RunDecimate:             true,
		ReforgingAllowNonCommon: false,
	}

	run := multi.JobCreate(printer, job, task)

	run.Run()

	//job.CullingReport()
	//run.TestDecimate()
	//run.RunCullingSets(200, time.Minute*15)
}
