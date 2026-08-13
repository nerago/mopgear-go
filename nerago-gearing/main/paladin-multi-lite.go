package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model/model_factory"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
	"slices"
	"time"
)

func PaladinMultiRunLite() {
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
		trinketTwinsGaze,
		trinketPrimRage,
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
	var phase3OneHandAndShield = []items.ItemId{
		96376, // worldbreaker weapon
		96534, // qon's scimitar
		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
	}
	var miscOtherP3 = []items.ItemId{
		95205, // terra-cotta neck
		94776, // primal turtle amulet
		96420, // talisman of angry spirits
		96373, // cloudbreaker belt heroic
		96395, // bloodsoaked legplates
		96542, // tidal force treads
		96500, // scaled tyrant heroic
		96377, // jinrohk soulcrystal
		96481, // durumu tentacle heroic
	}

	// ORGRIMMAR
	var timeless = []items.ItemId{
		101882, // cliffbreaker helm exp/mastery
		101887, // timeless ring haste/mastery
		101947, // timeless ring exp/mastery
	}
	var celestial = []items.ItemId{
		105122, // Asgorathian Blood Seal
		105011, // Demolisher's Reinforced Belt
		105077, // Visage of the Monstrous (haste shield)
	}
	var celestialRaden = []items.ItemId{
		95011, // lighting clawfeet
		95020, // ra-den contemplative loop
		95022, // Ra-den's Ruinous Ring
		95038, // carapace core
	}
	var orgRaidDrops = []items.ItemId{
		103787, // poisonbinder girth
		103742, // blood rage bracers
		103738, // bubble bracers
		105785, // vanguard burly bracer
		103734, // zoid gauntlets
		103735, // tar-coated gauntlets
		103916, // jugg ignition keys
		104461, // rage-blind greathelm
		104415, // bubble bracer heroic
		103892, // tharnok helm
		103915, // icy blood chest
		105767, // hoodrych chest ordos
		104417, // corruption-rotted gauntlets
		104416, // chest congealed corruption heroic
		103796, // seal kings norm
		103798, // bloodclaw band
		103737, // breastplate shaman mirror
	}
	var orgOneHandAndShield = []items.ItemId{
		103826, // xifeng weapon
		103872, // bulwurk of fallen general
		103871, // tower shield
		104485, // shield of mockery
		103972, // kilruk sword
		104464, // xifeng heroic
	}

	var legendCloaks = []items.ItemId{legendTankCloak, legendMeleeCloak}

	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "multi-set")

	simSize := simulate.RunSize_Largish
	//simSize := simulate.RunSize_Common
	//simSize := simulate.RunSize_QuickDirty

	job := multi_types.JobInputs{}
	job.SetWeightTypes(1, 2, 3)
	//job.SetWeightTypes(1)
	//job.SetWeightTypes(2)
	job.SetSimSize(simSize)
	job.SetReforgingAllowNonCommon(true)
	//job.SetWriteBestToGearFiles()

	var generalUpgrade items.UpgradeLevel = 0
	var forceUpgrade items.UpgradeLevel = 0

	ret := multi_types.SpecParam{
		Label: "Ret",
		Model: model_factory.Model_PallyRet(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                  files.GearFileRet,
			RequestRatingPercent:      0.04,
			ExtraUpgradeLevel:         generalUpgrade,
			ForceUpgradeExistingItems: 0,
			MissingEnchant:            setup.MissingEnchant_Panic,
		},
	}
	protDps := multi_types.SpecParam{
		Label: "Prot-Damage",
		Model: model_factory.Model_PallyProtDamage(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                  files.GearFileProtDamage,
			RequestRatingPercent:      0.01,
			ExtraUpgradeLevel:         generalUpgrade,
			ForceUpgradeExistingItems: forceUpgrade,
			MissingEnchant:            setup.MissingEnchant_Panic,
		},
	}
	protBalanced := multi_types.SpecParam{
		Label: "Prot-Balanced",
		Model: model_factory.Model_PallyProtBalanced(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                  files.GearFileProtBalanced,
			RequestRatingPercent:      0.25,
			ExtraUpgradeLevel:         generalUpgrade,
			ForceUpgradeExistingItems: forceUpgrade,
			MissingEnchant:            setup.MissingEnchant_Panic,
		},
	}
	protMitigation := multi_types.SpecParam{
		Label: "Prot-Mitigation",
		Model: model_factory.Model_PallyProtMitigation(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                  files.GearFileProtMitigation,
			RequestRatingPercent:      0.35,
			ExtraUpgradeLevel:         generalUpgrade,
			ForceUpgradeExistingItems: forceUpgrade,
			MissingEnchant:            setup.MissingEnchant_Panic,
		},
	}
	protSurvival := multi_types.SpecParam{
		Label: "Prot-Survival",
		Model: model_factory.Model_PallyProtSurvival(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                  files.GearFileProtSurvival,
			RequestRatingPercent:      0.30,
			ExtraUpgradeLevel:         generalUpgrade,
			ForceUpgradeExistingItems: forceUpgrade,
			MissingEnchant:            setup.MissingEnchant_Panic,
		},
	}
	protHeal := multi_types.SpecParam{
		Label: "Prot-Heal",
		Model: model_factory.Model_PallyProtHeal(),
		ItemInputs: multi_types.ItemInputs{
			GearFile:                  files.GearFileProtHeal,
			RequestRatingPercent:      0.05,
			ExtraUpgradeLevel:         generalUpgrade,
			ForceUpgradeExistingItems: forceUpgrade,
			MissingEnchant:            setup.MissingEnchant_Panic,
		},
	}

	ret.AddExtraItem(legendMeleeCloak)
	addExtrasToEach(legendCloaks, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(retT15, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(retT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(protT15, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(protT16, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(trinketsDpsP3, &ret, &protDps)
	addExtrasToEach(trinketsDpsP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(trinketsTankP3, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	addExtrasToEach(phase3OneHandAndShield, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(miscOtherP3, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	newStuffP5 := slices.Concat(timeless, celestial, celestialRaden, orgRaidDrops, newTrinketsDamage)
	addExtrasToEach(newStuffP5, &ret, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(newTrinketsTank, &protBalanced, &protMitigation, &protSurvival, &protHeal)
	addExtrasToEach(orgOneHandAndShield, &protDps, &protBalanced, &protMitigation, &protSurvival, &protHeal)

	ret.AddExtraItems([]items.ItemId{
		95281,  // ret tier15 gloves normal
		96657,  // ret tier15 legs heroic
		95535,  // normal lightning legs
		96533,  // rein-binders fists heroic
		96468,  // talonrender chest heroic
		104993, // Evil Eye of Galakras trinket celestial
		94773,  // Shoulderguards of Centripetal Destruction
		105033, // Wolf-Rider Spurs
		104938, // Sorrowpath Signet
		95140,  // shado assault band
		104981, // greatsword pride fall
		103968, // britomart pike
		96394,  // Wrist; Frozen Warlord's Bracers
	})

	protDps.AddExtraItems([]items.ItemId{
		86979,  // heroic impaling treads
		94773,  // centripetal shoulders normal
		95535,  // normal lightning legs
		96533,  // rein-binders fists heroic
		95140,  // shado assault band
		96468,  // talonrender chest heroic
		96657,  // ret tier15 legs heroic
		96375,  // Bracers of Constant Implosion
		100644, // Grievous Gladiator's Warboots of Cruelty
		104938, // Sorrowpath Signet celestial
		105033, // Wolf-Rider Spurs
		95153,  // Tyrant King Battleplate
	})

	protBalanced.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		94773,  // centripetal shoulders normal
		95535,  // normal lightning legs
		96533,  // rein-binders fists heroic
		96394,  // frozen warlord bracer heroic
		95140,  // shado assault band
		96468,  // talonrender chest heroic
		103791, // gauntlet of discarded
		96447,  // Rot-Proof Greatplate
		86955,  // Waistplate of Overwhelming Assault
		105033, // Wolf-Rider Spurs
	})

	protMitigation.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		95535,  // normal lightning legs
		96533,  // rein-binders fists heroic
		96550,  // doomed crown heroic
		96394,  // frozen warlord bracer heroic
		103791, // gauntlet of discarded
		96447,  // Rot-Proof Greatplate
		105090, // Ominous Mogu Greatboots
		86955,  // Waistplate of Overwhelming Assault
		105033, // Wolf-Rider Spurs
	})

	protSurvival.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		96447,  // Rot-Proof Greatplate
		105090, // Ominous Mogu Greatboots
		103791, // gauntlet of discarded
		105033, // Wolf-Rider Spurs
	})

	protHeal.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		96394,  // frozen warlord bracer heroic
		95142,  // Striker's Battletags
		105033, // Wolf-Rider Spurs
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
	// TODO bake this into process
	blockHelmetsWithoutCapacitance(&ret)
	blockHelmetsWithoutCapacitance(&protDps)
	blockHelmetsWithoutCapacitance(&protBalanced)
	blockHelmetsWithoutIndomitable(&protMitigation)
	blockHelmetsWithoutIndomitable(&protSurvival)
	blockHelmetsWithoutIndomitable(&protHeal)

	job.SetMinimumExtraItemLevel(463)
	ret.AddBagsExtra()
	protDps.AddBagsExtra()
	protBalanced.AddBagsExtra()
	protMitigation.AddBagsExtra()
	protSurvival.AddBagsExtra()
	protHeal.AddBagsExtra()

	//addExtrasFromFinder(loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic),
	//	&ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protBalanced)
	job.AddSetParam(protMitigation)
	job.AddSetParam(protSurvival)
	job.AddSetParam(protHeal)

	//job.EnablePermuteOnItemCountOptions()

	job.VerifyNoExtraDuplicates()
	//job.RemoveAnyExtraDuplicates()

	run := multi.JobCreate(printer, job)

	//job.RunNoPermutations_AllCommonAlternates(true, true)
	//run.RunNoPermutations_BestOnly(false, false)
	//run.RunForSolutionsPerPermute(1, false)

	//job.CullingReport()
	run.RunCullingSets(200, time.Minute*30)
}
