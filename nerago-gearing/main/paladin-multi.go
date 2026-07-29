package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
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

	//trinketSkeer // TODO
	trinketFusionCoreCelestial  = 104961
	trinketThokTailCelestial    = 105111
	trinketVialCorruptNormal    = 102306
	trinketRookUnluckyNormal    = 102296
	trinketEyeGalakrasCelestial = 104993

	legendMeleeCloak = 102249
	legendTankCloak  = 102250

	ringScaledTyrant = 96500
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
	96664, // prot tier15 chest heroic
	96666, // prot tier15 head heroic
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
}
var orgOneHandAndShield = []items.ItemId{
	103826, // xifeng weapon
	103872, // bulwurk of fallen general
	103871, // tower shield
	104485, // shield of mockery
	103972, // kilruk sword
}

var legendCloaks = []items.ItemId{legendTankCloak, legendMeleeCloak}

func PaladinMultiRun() {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "multi-set")

	//job := multi.MultiSetJob_Create(printer, simulate.RunSize_Largish)
	job := multi.MultiSetJob_Create(printer, simulate.RunSize_Common)
	//job := multi.MultiSetJob_Create(printer, simulate.RunSize_QuickDirty)
	job.SetWriteBestToGearFiles()

	var generalUpgrade items.UpgradeLevel = 0
	var forceUpgrade items.UpgradeLevel = 0

	ret := multi_types.MultiSetParam{
		Label:                     "Ret",
		GearFile:                  files.GearFileRet,
		Model:                     gear_model.Model_PallyRet(),
		RequestRatingPercent:      0.01,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: 0,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protDps := multi_types.MultiSetParam{
		Label:                     "Prot-Damage",
		GearFile:                  files.GearFileProtDps,
		Model:                     gear_model.Model_PallyProtDps(),
		RequestRatingPercent:      0.04,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protCompromise := multi_types.MultiSetParam{
		Label:                     "Prot-Compromise",
		GearFile:                  files.GearFileProtCompromise,
		Model:                     gear_model.Model_PallyProtCompromise(),
		RequestRatingPercent:      0.25,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protMitigationNoSet := multi_types.MultiSetParam{
		Label:                     "Prot-Mitigation-NoSet",
		GearFile:                  files.GearFileProtMitigationNoSet,
		Model:                     gear_model.Model_PallyProtMitigation_NoSet(),
		RequestRatingPercent:      0.30,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protMitigationWithSet := multi_types.MultiSetParam{
		Label:                     "Prot-Mitigation-WithSet",
		GearFile:                  files.GearFileProtMitigationWithSet,
		Model:                     gear_model.Model_PallyProtMitigation_WithSet(),
		RequestRatingPercent:      0.35,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protHeal := multi_types.MultiSetParam{
		Label:                     "Prot-Heal",
		GearFile:                  files.GearFileProtMitigationWithSet,
		Model:                     gear_model.Model_PallyProtHeal(),
		RequestRatingPercent:      0.05,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}

	addExtrasToEach(legendCloaks, &ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	addExtrasToEach(retT15, &ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)
	addExtrasToEach(retT16, &ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	addExtrasToEach(protT15, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)
	addExtrasToEach(protT16, &ret, &protDps) // just ilevel temporary
	addExtrasToEach(protT16, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	addExtrasToEach(trinketsDpsP3, &ret, &protDps)
	addExtrasToEach(trinketsDpsP3, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)
	addExtrasToEach(trinketsTankP3, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	addExtrasToEach(phase3OneHandAndShield, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)
	addExtrasToEach(miscOtherP3, &ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

	newStuffP5 := slices.Concat(timeless, celestial, celestialRaden, orgRaidDrops, newTrinketsDamage)
	addExtrasToEach(newStuffP5, &ret, &protDps, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)
	addExtrasToEach(newTrinketsTank, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)
	addExtrasToEach(orgOneHandAndShield, &protCompromise, &protMitigationNoSet, &protMitigationWithSet, &protHeal)

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
		104938, // Sorrowpath Signet
		105033, // Wolf-Rider Spurs
		95153,  // Tyrant King Battleplate
	})

	protCompromise.AddExtraItems([]items.ItemId{
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

	protMitigationNoSet.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		95535,  // normal lightning legs
		96533,  // rein-binders fists heroic
		96550,  // doomed crown heroic
		96394,  // frozen warlord bracer heroic
		103791, // gauntlet of discarded
		96667,  // prot tier15 leg heroic
		96447,  // Rot-Proof Greatplate
		105090, // Ominous Mogu Greatboots
		86955,  // Waistplate of Overwhelming Assault
		105033, // Wolf-Rider Spurs
	})

	protMitigationWithSet.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		96667,  // prot tier15 leg heroic
		95291,  // prot tier15 Lightning Emperor's Handguards
		96447,  // Rot-Proof Greatplate
		105090, // Ominous Mogu Greatboots
		103791, // gauntlet of discarded
	})

	protHeal.AddExtraItems([]items.ItemId{
		96478, // treads of the blind heroic
		96394, // frozen warlord bracer heroic
		95142, // Striker's Battletags
	})

	// predetermined choices
	ret.ForceSingleSlot(items.Equip_Weapon, 103968) // britomark
	ret.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	ret.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	ret.ForceSingleSlot(items.Equip_Trinket2, trinketTwinsGaze)
	protDps.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	protDps.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protDps.ForceSingleSlot(items.Equip_Trinket2, trinketTwinsGaze)
	protCompromise.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	protCompromise.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protCompromise.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	protCompromise.AddReportVariant(items.Equip_Trinket2, trinketFortZand)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket1, trinketThokTailCelestial) // changed
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket2, trinketVialCorruptNormal)
	protMitigationWithSet.AddReportVariant(items.Equip_Trinket2, trinketThokTailCelestial)
	protHeal.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protHeal.ForceSingleSlot(items.Equip_Trinket1, trinketTwinsGaze)
	protHeal.ForceSingleSlot(items.Equip_Trinket2, trinketZandSpark)

	// protCompromise.ForceTryAllSlot(items.Equip_Weapon, []items.ItemId{96376, 96534, 103826})
	// protMitigationNoSet.ForceTryAllSlot(items.Equip_Weapon, []items.ItemId{96376, 96534, 103826})
	// protMitigationWithSet.ForceTryAllSlot(items.Equip_Weapon, []items.ItemId{96376, 96534, 103826})
	// protHeal.ForceTryAllSlot(items.Equip_Weapon, []items.ItemId{96376, 96534, 103826})

	// HELMET
	// TODO bake this into process
	blockHelmetsWithoutCapacitance(&ret)
	blockHelmetsWithoutCapacitance(&protDps)
	blockHelmetsWithoutCapacitance(&protCompromise)
	blockHelmetsWithoutIndomitable(&protMitigationNoSet)
	blockHelmetsWithoutIndomitable(&protMitigationWithSet)
	blockHelmetsWithoutIndomitable(&protHeal)

	//job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Haste, 320))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Stamina, 120))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Strength, 80, stats.Stat_Haste, 160))
	//job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Strength, 160))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Expertise, 160, stats.Stat_Hit, 160))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Hit, 160))
	//job.ActivateAlternateGemmingAsPermute()

	//job.MakeRandomVariants(101887, 0, -365, -352)

	job.SetMinimumExtraItemLevel(463)
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
	job.AddSetParam(protCompromise)
	job.AddSetParam(protMitigationNoSet)
	job.AddSetParam(protMitigationWithSet)
	job.AddSetParam(protHeal)

	// as part of BIS calcs try each meta as needed
	//capacitanceSets := []multi_types.MultiSetParam{ret, protDps, protCompromise}
	//indomitableSets := []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet, protHeal}
	//job.AddItemDistinctUsageGroups(104590, false, capacitanceSets, indomitableSets)
	//job.AddItemDistinctUsageGroups(104647, false, capacitanceSets, indomitableSets)
	//job.AddItemDistinctUsageGroups(105807, false, capacitanceSets, indomitableSets)
	//job.AddItemDistinctUsageGroups(99379, false, capacitanceSets, indomitableSets)
	//job.AddItemDistinctUsageGroups(104492, false, capacitanceSets, indomitableSets)

	//job.AddItemDistinctUsageGroups(103892, true, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet, protHeal})
	//ret.ForceTryAllSlot(items.Equip_Weapon, []items.ItemId{104981, 86386})
	//job.AddAlternateUpgradeChoices(
	//	101947, // Elder Tortoiseshell Seal
	//	99028,  // Handguards of Winged Triumph celestial
	//	103972, // kilruk sword
	//	105122, // Asgorathian Blood Seal
	//	103915, // Icy Blood Chestplate
	//	105033, // Wolf-Rider Spurs
	//)
	//job.AddAlternateUpgradeChoices(104417) //gloves corrupt

	job.VerifyNoExtraDuplicates()
	//job.RemoveAnyExtraDuplicates()

	job.RunNoPermutations_AllCommonAlternates(true)
	//job.RunNoPermutations_BestOnly(true, false)
	//job.RunForSolutionsPerPermute(3)

	//job.CullingReport()
	//job.RunCullingSets(500, time.Minute*30)
}

func addExtrasToEach(itemIdList []items.ItemId, params ...*multi_types.MultiSetParam) {
	for _, param := range params {
		param.AddExtraItems(itemIdList)
	}
}

func blockHelmetsWithoutCapacitance(param *multi_types.MultiSetParam) {
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

func blockHelmetsWithoutIndomitable(param *multi_types.MultiSetParam) {
	param.BlockItem(87024)  // nullification greathelm = capacitance
	param.BlockItem(95282)  // lightning emp helmet = capacitance
	param.BlockItem(101882) // cliffbreaker helm = capacitance
	param.BlockItem(98985)  // ret helm = capacitance
	param.BlockItem(103892) // thranok = capacitance
	blockGeneral(param)
}

func blockGeneral(param *multi_types.MultiSetParam) {
	param.BlockItem(95513)  // normal ring
	param.BlockItem(95778)  // golden golem celestial
	param.BlockItem(101942) // Elder Tortoiseshell Helm (just blocking to gem bis runs done, don't have one)
}

func addExtrasFromFinder(foundList []loaders.ItemFoundRef, params ...*multi_types.MultiSetParam) {
	for _, itemRef := range foundList {
		for _, param := range params {
			param.AddExtraItem(itemRef.ItemId)
		}
	}
}
