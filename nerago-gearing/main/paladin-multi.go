package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

const (
	trinketZandSpark   = 96398
	trinketFortZand    = 96793
	trinketPrimRage    = 94519
	trinketTwinsGaze   = 94529
	trinketJiKun       = 94527
	trinketSoulBarrier = 96555

	trinketCurseHubris  = 105645 // heroic warforged
	trinketThokTail     = 105609 // heroic warforged
	trinketFusionCore   = 105459 // heroic warforged
	trinketSkeerBlood   = 105632 // heroic warforged
	trinketVialCorrupt  = 105568 // heroic warforged
	trinketRookUnlucky  = 105438 // heroic warforged
	trinketJuggFocusing = 105514 // heroic warforged
	trinketEyeGalakras  = 102298

	legendMeleeCloak = 102249
	legendTankCloak  = 102250

	ringScaledTyrant = 96500
)

func PaladinMultiRun(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer, simulate.RunSize_Medium)
	// job := multi.MultiSetJob_Create(printer, simulate.RunSize_QuickDirty)

	ret := multi_types.MultiSetParam{
		Label:                     "Ret",
		GearFile:                  files.GearFileRet,
		Model:                     model.Model_PallyRet(),
		RequestRatingPercent:      0.01,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 0,
	}
	protDps := multi_types.MultiSetParam{
		Label:                     "Prot-Damage",
		GearFile:                  files.GearFileProtDps,
		Model:                     model.Model_PallyProtDps(),
		RequestRatingPercent:      0.09,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protCompromise := multi_types.MultiSetParam{
		Label:                     "Prot-Compromise",
		GearFile:                  files.GearFileProtCompromise,
		Model:                     model.Model_PallyProtCompromise(),
		RequestRatingPercent:      0.30,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protMitigationNoSet := multi_types.MultiSetParam{
		Label:                     "Prot-Mitigation-NoSet",
		GearFile:                  files.GearFileProtMitigationNoSet,
		Model:                     model.Model_PallyProtMitigation_NoSet(),
		RequestRatingPercent:      0.45,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protMitigationWithSet := multi_types.MultiSetParam{
		Label:                     "Prot-Mitigation-WithSet",
		GearFile:                  files.GearFileProtMitigationWithSet,
		Model:                     model.Model_PallyProtMitigation_WithSet(),
		RequestRatingPercent:      0.15,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}

	ret.AddExtraItems([]items.ItemId{
		legendMeleeCloak,

		95140, // shado assault band
		95141, // shado assault loop
		86957, // heroic bladed tempest ring
		96500, // scaled tyrant heroic
		96481, // durumu tentacle h
		96377, // jinrohk soulcrystaleroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		87024, // null greathelm
		96373, // cloudbreaker belt heroic
		96478, // treads of the blind heroic

		95535, // normal lightning legs REMOVE12
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate

		96550, // doomed crown heroic
		87024, // null greathelm
		95282, // ret tier15 normal head
		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic

		85343, // white tiger battleplate t14
		96447, // rot-proof greatplate
		87100, // White Tiger Gauntlets

		95142, // striker's battletags
		95205, // terra-cotta neck
		94776, // primal turtle amulet
		96420, // talisman of angry spirits

		96375, // bracers implosion
		96395, // bloodsoaked legplates
		96542, // tidal force treads

		trinketZandSpark,
		trinketJiKun,     // ji-kun trinket
		trinketTwinsGaze, // gaze of the twins
		trinketPrimRage,

		87145, // defiled earth OFF
		96394, // frozen warlord bracer heroic

		101882, // cliffbreaker helm exp/mastery
		103787, // poisonbinder girth
		103742, // blood rage bracers
	})
	blockHelmetsWithoutCapacitance(&ret)
	ret.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	ret.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	ret.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	ret.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)

	protDps.AddExtraItems([]items.ItemId{
		86957, // heroic bladed tempest ring
		95140, // shado assault band
		95141, // shado assault loop OFF123
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic
		96377, // jinrohk soulcrystal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		96478, // treads of the blind heroic

		96373, // cloudbreaker belt heroic

		legendMeleeCloak,

		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate

		96550, // doomed crown heroic
		87024, // null greathelm
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
		96542, // tidal force treads

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic

		101882, // cliffbreaker helm exp/mastery
		103787, // poisonbinder girth
		103742, // blood rage bracers
	})
	blockHelmetsWithoutCapacitance(&protDps)
	protDps.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protDps.ForceSingleSlot(items.Equip_Back, legendMeleeCloak)
	protDps.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark) // zandalar trinket
	protDps.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)

	protCompromise.AddExtraItems([]items.ItemId{
		86957, // heroic bladed tempest ring
		95140, // shado assault band
		95141, // shado assault loop
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic
		96377, // jinrohk soulcrystal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		96478, // treads of the blind heroic

		96373, // cloudbreaker belt heroic

		96428, // shell-coated wrists
		96447, // rot-proof greatplate

		95535, // normal lightning legs
		94773, // centripetal shoulders normal REMOVE12
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate REMOVE1234

		96550, // doomed crown heroic
		87024, // null greathelm
		95282, // ret tier15 normal head

		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic
		96666, // prot tier15 head heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		94776, // primal turtle amulet
		96420, // talisman of angry spirits

		87145, // defiled earth OFF12
		94820, // caustic spike bracers OFF1234
		96394, // frozen warlord bracer heroic

		96375, // bracers implosion
		96395, // bloodsoaked legplates
		96542, // tidal force treads

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,
		trinketSoulBarrier,
		trinketFortZand,

		legendTankCloak,
		legendMeleeCloak,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic

		101882, // cliffbreaker helm exp/mastery
		103787, // poisonbinder girth
		103742, // blood rage bracers
	})
	blockHelmetsWithoutCapacitance(&protCompromise)
	protCompromise.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protCompromise.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	protCompromise.ForceSingleSlot(items.Equip_Trinket2, trinketSoulBarrier)
	protCompromise.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	protMitigationNoSet.AddExtraItems([]items.ItemId{
		86979, // heroic impaling treads REMOVE1
		87015, // heroic clawfeet REMOVE1
		96478, // treads of the blind heroic

		86957, // heroic bladed tempest ring REMOVE12
		95140, // shado assault band REMOVE123
		95141, // shado assault loop
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic
		96377, // jinrohk soulcrystal

		96373, // cloudbreaker belt heroic

		legendMeleeCloak,
		legendTankCloak,

		95535, // normal lightning legs
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		96428, // shell-coated wrists UPGRADE LIKED
		96447, // rot-proof greatplate

		96550, // doomed crown heroic
		95910, // ret tier15 chest celestial REMOVE12
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic

		95290, // prot tier15 chest normal
		95291, // prot tier15 hand normal
		96666, // prot tier15 head heroic
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		95178, // lootraptor amulet
		94776, // primal turtle amulet
		96420, // talisman of angry spirits

		87145, // defiled earth
		94820, // caustic spike bracers REMOVE1234
		96394, // frozen warlord bracer heroic

		96375, // bracers implosion
		96395, // bloodsoaked legplates
		96542, // tidal force treads

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,
		trinketSoulBarrier,
		trinketFortZand,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic

		101882, // cliffbreaker helm exp/mastery
		103787, // poisonbinder girth
		103742, // blood rage bracers
	})
	blockHelmetsWithoutIndomitable(&protMitigationNoSet)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	protMitigationWithSet.AddExtraItems([]items.ItemId{
		86979, // heroic impaling treads
		96478, // treads of the blind heroic

		86957, // heroic bladed tempest ring REMOVE12
		95140, // shado assault band REMOVE12
		95141, // shado assault loop
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic
		96377, // jinrohk soulcrystal

		96373, // cloudbreaker belt heroic

		legendTankCloak,

		95535, // normal lightning legs REMOVE12
		94773, // centripetal shoulders normal REMOVE1234
		96468, // talonrender chest heroic REMOVE12
		96533, // rein-binders fists heroic
		96428, // shell-coated wrists REMOVE1234
		96447, // rot-proof greatplate

		96657, // ret tier15 legs heroic   REMOVE12
		96658, // ret tier15 shoulder heroic  REMOVE12

		96550, // doomed crown heroic
		95291, // prot tier15 hand normal
		96664, // prot tier15 chest heroic
		96666, // prot tier15 head heroic
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		95178, // lootraptor amulet REMOVE123
		94776, // primal turtle amulet REMOVE12
		96420, // talisman of angry spirits

		87145, // defiled earth
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		96375, // bracers implosion
		96395, // bloodsoaked legplates
		96542, // tidal force treads

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,
		trinketSoulBarrier,
		trinketFortZand,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic

		101882, // cliffbreaker helm exp/mastery
		103787, // poisonbinder girth
		103742, // blood rage bracers
	})
	blockHelmetsWithoutIndomitable(&protMitigationWithSet)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Ring1, ringScaledTyrant)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Back, legendTankCloak)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)
	protMitigationWithSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	// ret.AddBagsExtra()
	// protDps.AddBagsExtra()
	// protCompromise.AddBagsExtra()
	// protMitigationNoSet.AddBagsExtra()
	// protMitigationWithSet.AddBagsExtra()

	job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Haste, 320))
	job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Expertise, 320))
	job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Hit, 320))
	job.AddAlternateGemming(stats.StatBlock_of(stats.Stat_Strength, 160))
	job.AddAlternateGemming(stats.StatBlock_of2(stats.Stat_Expertise, 160, stats.Stat_Hit, 160))

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protCompromise)
	job.AddSetParam(protMitigationNoSet)
	job.AddSetParam(protMitigationWithSet)

	// job.AddItemDistinctUsageGroups(96550, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet})
	// job.AddItemDistinctUsageGroups(101882, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet})

	job.FindHighsResult_Sample(1)
	// job.FindSeveralHighsAndSim()
	// job.FindHighsResultPerPermute(12)
	// job.RunWithMinimumHaste("Prot-Mitigation-WithSet", 11000, 18000, 250)
	// job.RunWithMinimumHaste("Prot-Mitigation-NoSet", 11000, 18000, 250)
}

func blockHelmetsWithoutCapacitance(param *multi_types.MultiSetParam) {
	// param.BlockItem(87024) // nullification greathelm = capacitance
	// param.BlockItem(95282) // lightning emp helmet = capacitance
	param.BlockItem(87101) // white tiger helmet = prot gem
	param.BlockItem(95292) // lightning emp faceguard = prot gem
	param.BlockItem(95778) // golden golem celestial = ignore in all sets
	param.BlockItem(96666) // lightning emp faceguard heroic = prot gem
	param.BlockItem(96550) // doomed crown heroic = prot gem
	blockGeneral(param)
}

func blockHelmetsWithoutIndomitable(param *multi_types.MultiSetParam) {
	param.BlockItem(87024)  // nullification greathelm = capacitance
	param.BlockItem(95282)  // lightning emp helmet = capacitance
	param.BlockItem(95778)  // golden golem celestial = ignore in all sets
	param.BlockItem(101882) // cliffbreaker helm = capacitance
	blockGeneral(param)
}

func blockGeneral(param *multi_types.MultiSetParam) {
	param.BlockItem(95513) // normal ring
}
