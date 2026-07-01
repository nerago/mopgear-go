package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
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

	trinketCurseHubrisHeroic  = 105645 // heroic warforged
	trinketThokTailHeroic     = 105609 // heroic warforged
	trinketFusionCoreHeroic   = 105459 // heroic warforged
	trinketSkeerBloodHeroic   = 105632 // heroic warforged
	trinketVialCorruptHeroic  = 105568 // heroic warforged
	trinketRookUnluckyHeroic  = 105438 // heroic warforged
	trinketJuggFocusingHeroic = 105514 // heroic warforged
	trinketEyeGalakrasHeroic  = 102298

	trinketFusionCoreCelestial  = 104961
	trinketThokTailCelestial    = 105111
	trinketVialCorruptCelestial = 105070
	trinketRookUnluckyNormal    = 102296

	legendMeleeCloak = 102249
	legendTankCloak  = 102250

	ringScaledTyrant = 96500
)

func PaladinMultiRun(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer, simulate.RunSize_Common)
	// job := multi.MultiSetJob_Create(printer, simulate.RunSize_QuickDirty)

	var generalUpgrade items.UpgradeLevel = 0
	var forceUpgrade items.UpgradeLevel = 0

	ret := multi_types.MultiSetParam{
		Label:                     "Ret",
		GearFile:                  files.GearFileRet,
		Model:                     model.Model_PallyRet(),
		RequestRatingPercent:      0.01,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: 0,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protDps := multi_types.MultiSetParam{
		Label:                     "Prot-Damage",
		GearFile:                  files.GearFileProtDps,
		Model:                     model.Model_PallyProtDps(),
		RequestRatingPercent:      0.04,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protCompromise := multi_types.MultiSetParam{
		Label:                     "Prot-Compromise",
		GearFile:                  files.GearFileProtCompromise,
		Model:                     model.Model_PallyProtCompromise(),
		RequestRatingPercent:      0.20,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protMitigationNoSet := multi_types.MultiSetParam{
		Label:                     "Prot-Mitigation-NoSet",
		GearFile:                  files.GearFileProtMitigationNoSet,
		Model:                     model.Model_PallyProtMitigation_NoSet(),
		RequestRatingPercent:      0.40,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protMitigationWithSet := multi_types.MultiSetParam{
		Label:                     "Prot-Mitigation-WithSet",
		GearFile:                  files.GearFileProtMitigationWithSet,
		Model:                     model.Model_PallyProtMitigation_WithSet(),
		RequestRatingPercent:      0.30,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}
	protHeal := multi_types.MultiSetParam{
		Label:                     "Prot-Heal",
		GearFile:                  files.GearFileProtMitigationWithSet,
		Model:                     model.Model_PallyProtHeal(),
		RequestRatingPercent:      0.05,
		ExtraUpgradeLevel:         generalUpgrade,
		ForceUpgradeExistingItems: forceUpgrade,
		MissingEnchant:            setup.MissingEnchant_Panic,
	}

	// CLOAK: add the cloaks to all even though we often override
	legendCloaks := []items.ItemId{legendTankCloak, legendMeleeCloak}
	ret.AddExtraItems(legendCloaks)
	protDps.AddExtraItems(legendCloaks)
	protCompromise.AddExtraItems(legendCloaks)
	protMitigationNoSet.AddExtraItems(legendCloaks)
	protMitigationWithSet.AddExtraItems(legendCloaks)
	protHeal.AddExtraItems(legendCloaks)

	// TIER
	retT15 := []items.ItemId{
		95282, // ret tier15 normal head
		// 95910, // ret tier15 chest celestial (cull)
		96658, // ret tier15 shoulder heroic
	}
	ret.AddExtraItems(retT15)
	protDps.AddExtraItems(retT15)
	protCompromise.AddExtraItems(retT15)
	protMitigationNoSet.AddExtraItems(retT15) // was heroic subset only

	retT16 := []items.ItemId{
		// yes numbers don't line up, seems like the way it is
		99052, // ret t16 chest celstial
		99002, // ret t16 hand celstial
		98985, // ret t16 head celstial
	}
	ret.AddExtraItems(retT16)
	protDps.AddExtraItems(retT16)
	protCompromise.AddExtraItems(retT16)
	protMitigationNoSet.AddExtraItems(retT16)

	protT15 := []items.ItemId{
		96664, // prot tier15 chest heroic
		96666, // prot tier15 head heroic
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic
	}
	protCompromise.AddExtraItems(protT15)
	protMitigationNoSet.AddExtraItems(protT15)
	protMitigationWithSet.AddExtraItems(protT15)
	protHeal.AddExtraItems(protT15)

	protT16 := []items.ItemId{
		99126, // prot t16 chest normal
		99128, // prot t16 head normal
		99129, // prot t16 legs normal
		99027, // prot t16 shoulder celestial
		99028, // prot t16 hand celestial
	}
	ret.AddExtraItems(protT16) // just ilevel temporory
	protCompromise.AddExtraItems(protT16)
	protMitigationNoSet.AddExtraItems(protT16)
	protMitigationWithSet.AddExtraItems(protT16)
	protHeal.AddExtraItems(protT16)

	// RING
	throneRings := []items.ItemId{
		// 86957, // heroic bladed tempest ring // cull
		96500, // scaled tyrant heroic

		96481, // durumu tentacle heroic (not a tank ring, but ret likes?)
	}
	throneTankRings := []items.ItemId{
		96377, // jinrohk soulcrystal
	}
	ret.AddExtraItems(throneRings)
	protDps.AddExtraItems(throneRings)
	protCompromise.AddExtraItems(throneRings)
	protCompromise.AddExtraItems(throneTankRings)
	protMitigationNoSet.AddExtraItems(throneRings)
	protMitigationNoSet.AddExtraItems(throneTankRings)
	protMitigationWithSet.AddExtraItems(throneRings)
	protMitigationWithSet.AddExtraItems(throneTankRings)
	protHeal.AddExtraItems(throneRings)
	protHeal.AddExtraItems(throneTankRings)

	// TRINKET
	trinketsDpsP3 := []items.ItemId{
		trinketZandSpark,
		// trinketJiKun, // culled
		trinketTwinsGaze,
		trinketPrimRage,
	}
	trinketsTankP3 := []items.ItemId{
		// trinketSoulBarrier, // culled
		trinketFortZand,
	}
	trinketsBothP3 := slices.Concat(trinketsDpsP3, trinketsTankP3)
	ret.AddExtraItems(trinketsDpsP3)
	protDps.AddExtraItems(trinketsDpsP3)
	protCompromise.AddExtraItems(trinketsBothP3)
	protMitigationNoSet.AddExtraItems(trinketsBothP3)
	protMitigationWithSet.AddExtraItems(trinketsBothP3)
	protHeal.AddExtraItems(trinketsBothP3) // TODO cut down on on-use trinkets etc

	// WEAPONS
	oneHandAndShieldP3 := []items.ItemId{
		96376, // worldbreaker weapon
		96534, // qon's scimitar
		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
	}
	protDps.AddExtraItems(oneHandAndShieldP3)
	protCompromise.AddExtraItems(oneHandAndShieldP3)
	protMitigationNoSet.AddExtraItems(oneHandAndShieldP3)
	protMitigationWithSet.AddExtraItems(oneHandAndShieldP3)
	protHeal.AddExtraItems(oneHandAndShieldP3)

	// NECK
	miscNecksP3 := []items.ItemId{
		// 95142, // striker's battletags (cull)
		95205, // terra-cotta neck
		94776, // primal turtle amulet
	}
	goodNeckP3 := []items.ItemId{
		96420, // talisman of angry spirits
	}
	allNecksP3 := slices.Concat(miscNecksP3, goodNeckP3)
	ret.AddExtraItems(allNecksP3)
	protDps.AddExtraItems(allNecksP3)
	protCompromise.AddExtraItems(allNecksP3)
	protMitigationNoSet.AddExtraItems(allNecksP3)
	protMitigationWithSet.AddExtraItems(allNecksP3)
	protHeal.AddExtraItems(allNecksP3)

	// ORGRIMMAR
	timeless := []items.ItemId{
		101882, // cliffbreaker helm exp/mastery
		101887, // timeless ring haste/mastery
	}
	celestial := []items.ItemId{
		105092, // tower shield
		105122, // blood seal
		// 104938, // sorrowpath signet // would need upgrade to be useful
	}
	raidDrops := []items.ItemId{
		103787, // poisonbinder girth
		103742, // blood rage bracers
		103738, // bubble bracers
		105785, // burly bracer
		103734, // zoid gauntlets
		103826, // xifeng weapon
		103735, // tar-coated gauntlets
		103872, // bulwurk of fallen general
	}
	newTrinkets := []items.ItemId{
		trinketThokTailCelestial,
		trinketFusionCoreCelestial,
		trinketVialCorruptCelestial,
		trinketRookUnluckyNormal,
	}
	newStuffP5 := slices.Concat(timeless, celestial, raidDrops, newTrinkets)
	ret.AddExtraItems(newStuffP5)
	protDps.AddExtraItems(newStuffP5)
	protCompromise.AddExtraItems(newStuffP5)
	protMitigationNoSet.AddExtraItems(newStuffP5)
	protMitigationWithSet.AddExtraItems(newStuffP5)
	protHeal.AddExtraItems(newStuffP5)

	ret.AddExtraItems([]items.ItemId{
		96373, // cloudbreaker belt heroic - good 2nd belt
		96395, // bloodsoaked legplates - decent 2nd legs, offset
		96542, // tidal force treads
		86386, // Shin'ka, Execution of Dominion
		95281, // ret tier15 gloves normal (moved from shared lists was 7.7%)
		96657, // ret tier15 legs heroic
	})

	protDps.AddExtraItems([]items.ItemId{
		86979, // heroic impaling treads
		96478, // treads of the blind heroic
		96373, // cloudbreaker belt heroic
		94773, // centripetal shoulders normal - actually good despite the level
		95535, // normal lightning legs - actually good despite the level
		96533, // rein-binders fists heroic
		96542, // tidal force treads
		95140, // shado assault band (decent 3rd)
		96468, // talonrender chest heroic
		96657, // ret tier15 legs heroic
	})

	protCompromise.AddExtraItems([]items.ItemId{
		86979,  // heroic impaling treads
		96478,  // treads of the blind heroic
		96373,  // cloudbreaker belt heroic
		94773,  // centripetal shoulders normal - actually good despite the level
		95535,  // normal lightning legs - actually good despite the level
		96533,  // rein-binders fists heroic - actually good despite the level
		96395,  // bloodsoaked legplates
		96542,  // tidal force treads
		96394,  // frozen warlord bracer heroic (moved from shared lists was still 3rd)
		95140,  // shado assault band (decent 4th)
		96468,  // talonrender chest heroic
		103791, // gauntlet of discarded
	})

	protMitigationNoSet.AddExtraItems([]items.ItemId{
		96478,  // treads of the blind heroic
		96373,  // cloudbreaker belt heroic
		95535,  // normal lightning legs
		96533,  // rein-binders fists heroic
		96550,  // doomed crown heroic
		96395,  // bloodsoaked legplates
		96542,  // tidal force treads
		96394,  // frozen warlord bracer heroic (moved from shared lists was still 3rd)
		103791, // gauntlet of discarded
	})

	protMitigationWithSet.AddExtraItems([]items.ItemId{
		96478, // treads of the blind heroic
		96373, // cloudbreaker belt heroic
		96533, // rein-binders fists heroic
		96550, // doomed crown heroic
		96395, // bloodsoaked legplates
		96542, // tidal force treads
		95291, // prot tier15 hand normal
	})

	protHeal.AddExtraItems([]items.ItemId{
		86979, // heroic impaling treads
		96478, // treads of the blind heroic
		96373, // cloudbreaker belt heroic
		96533, // rein-binders fists heroic
		96550, // doomed crown heroic
		96395, // bloodsoaked legplates
		96542, // tidal force treads
	})

	// predetermined choices
	ret.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	ret.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	ret.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	ret.ForceSingleSlot(items.Equip_Trinket2, trinketTwinsGaze)
	// protDps.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protDps.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	protDps.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protDps.ForceSingleSlot(items.Equip_Trinket2, trinketTwinsGaze)
	// protCompromise.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protCompromise.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protCompromise.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	protCompromise.AddReportVariant(items.Equip_Trinket2, trinketFortZand)
	// protMitigationNoSet.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket1, trinketThokTailCelestial)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket1, trinketZandSpark)
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)
	//protMitigationWithSet.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket2, trinketVialCorruptCelestial)
	protMitigationWithSet.AddReportVariant(items.Equip_Trinket2, trinketThokTailCelestial)
	// protHeal.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
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
	// job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Expertise, 320))
	// job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Hit, 320))
	// job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Strength, 160))
	// job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Expertise, 160, stats.Stat_Hit, 160))
	// job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Haste, 160, stats.Stat_Hit, 160))
	//job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Strength, 80, stats.Stat_Haste, 160))

	job.MakeRandomVariants(101887, 0, -365, -352)

	ret.AddBagsExtra()
	protDps.AddBagsExtra()
	protCompromise.AddBagsExtra()
	protMitigationNoSet.AddBagsExtra()
	protMitigationWithSet.AddBagsExtra()
	protHeal.AddBagsExtra()

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protCompromise)
	job.AddSetParam(protMitigationNoSet)
	job.AddSetParam(protMitigationWithSet)
	job.AddSetParam(protHeal)

	// job.AddItemDistinctUsageGroups(96550, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet})
	// job.AddItemDistinctUsageGroups(101882, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet})

	// job.RunNoPermutations_AllCommonAlternates()
	job.RunForSolutionsPerPerumte(6)
	// job.RunCullingSets(500, time.Minute*30)
}

func blockHelmetsWithoutCapacitance(param *multi_types.MultiSetParam) {
	param.BlockItem(87101) // white tiger helmet = prot gem
	param.BlockItem(95292) // lightning emp faceguard = prot gem
	param.BlockItem(95778) // golden golem celestial = ignore in all sets
	param.BlockItem(96666) // lightning emp faceguard heroic = prot gem
	param.BlockItem(96550) // doomed crown heroic = prot gem
	param.BlockItem(99128) // winged faceguard = prot gem
	blockGeneral(param)
}

func blockHelmetsWithoutIndomitable(param *multi_types.MultiSetParam) {
	param.BlockItem(87024)  // nullification greathelm = capacitance
	param.BlockItem(95282)  // lightning emp helmet = capacitance
	param.BlockItem(95778)  // golden golem celestial = ignore in all sets
	param.BlockItem(101882) // cliffbreaker helm = capacitance
	param.BlockItem(98985)  // ret helm = capacitance
	blockGeneral(param)
}

func blockGeneral(param *multi_types.MultiSetParam) {
	param.BlockItem(95513) // normal ring
}
