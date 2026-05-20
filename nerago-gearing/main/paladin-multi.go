package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
)

const (
	trinketZandSpark    = 96398
	trinketFortZand     = 96793
	trinketPrimRage     = 94519
	trinketTwinsGaze    = 94529
	trinketJiKun        = 94527
	trinketSoulBarrier  = 96555
	preLegendMeleeCloak = 98147
	preLegendTankCloak  = 98146
)

func PaladinMultiRun(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer, simulate.RunSize_Medium)
	// job := multi.MultiSetJob_Create(printer, simulate.RunSize_QuickDirty)

	ret := multi_types.MultiSetParam{
		Label:    "Ret",
		GearFile: files.GearFileRet,
		Model:    model.Model_PallyRet(),
		// IncludeInFirstPass:        false,
		RequestRatingPercent:      0.01,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	ret.AddExtraItems([]items.ItemId{
		preLegendMeleeCloak, // pre-legend strength dps

		95140, // shado assault band
		95141, // shado assault loop
		86957, // heroic bladed tempest ring
		96500, // scaled tyrant heroic
		96481, // durumu tentacle h
		96377, // jinrohk soulcrystaleroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		87024, // null greathelm
		94726, // cloudbreaker belt normal
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

		trinketZandSpark,
		trinketJiKun,     // ji-kun trinket
		trinketTwinsGaze, // gaze of the twins
		trinketPrimRage,

		87145, // defiled earth OFF
		96394, // frozen warlord bracer heroic
	})
	ret.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	// ret.ForceTryAllSlot(items.Equip_Trinket2, trinketTwinsGaze, trinketPrimRage)
	ret.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	ret.ForceSingleSlot(items.Equip_Ring2, 96500)              // scaled tyrant heroic
	ret.ForceSingleSlot(items.Equip_Back, preLegendMeleeCloak) // pre-legend strength dps
	blockHelmetsWithoutCapacitance(&ret)

	protDps := multi_types.MultiSetParam{
		Label:    "Prot-Damage",
		GearFile: files.GearFileProtDps,
		Model:    model.Model_PallyProtDps(),
		// IncludeInFirstPass:        true,
		RequestRatingPercent:      0.09,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
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
	})
	protDps.ForceSingleSlot(items.Equip_Ring2, 96500)               // scaled tyrant heroic
	protDps.ForceSingleSlot(items.Equip_Back, preLegendMeleeCloak)  // pre-legend strength dps
	protDps.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark) // zandalar trinket
	protDps.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	// protDps.ForceTryAllSlot(items.Equip_Offhand, 94945, 96182, 96436)
	blockHelmetsWithoutCapacitance(&protDps)

	protCompromise := multi_types.MultiSetParam{
		Label:    "Prot-Compromise",
		GearFile: files.GearFileProtCompromise,
		Model:    model.Model_PallyProtCompromise(),
		// IncludeInFirstPass:        true,
		RequestRatingPercent:      0.35,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
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

		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic

		preLegendMeleeCloak, // pre-legend strength dps

		96428, // shell-coated wrists
		96447, // rot-proof greatplate

		95535, // normal lightning legs
		94773, // centripetal shoulders normal REMOVE12
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate REMOVE1234

		96550, // doomed crown heroic
		87024, // null greathelm
		95778, // crown golden golem celestial [would need gem, acceptable]
		95282, // ret tier15 normal head [would need gem]

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

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,
		trinketSoulBarrier,
		trinketFortZand,

		preLegendMeleeCloak,
		preLegendTankCloak,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic
	})
	protCompromise.ForceSingleSlot(items.Equip_Ring2, 96500)             // scaled tyrant heroic
	protCompromise.ForceSingleSlot(items.Equip_Back, preLegendTankCloak) //                         see spreadsheet multisolve4d
	// protCompromise.ForceTryAllSlot(items.Equip_Back, preLegendMeleeCloak, preLegendTankCloak) // see spreadsheet multisolve4d
	blockHelmetsWithoutCapacitance(&protCompromise)

	protCompromise.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	// protCompromise.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	// protCompromise.BlockItem(trinketFortZand)
	// protCompromise.AddReportVariant(items.Equip_Trinket2, trinketFortZand)

	protCompromise.ForceSingleSlot(items.Equip_Trinket2, trinketSoulBarrier)
	protCompromise.BlockItem(trinketPrimRage)
	protCompromise.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	protMitigationNoSet := multi_types.MultiSetParam{
		Label:    "Prot-Mitigation-NoSet",
		GearFile: files.GearFileProtMitigationNoSet,
		Model:    model.Model_PallyProtMitigation_NoSet(),
		// IncludeInFirstPass:        true,
		RequestRatingPercent:      0.25,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
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

		preLegendMeleeCloak, // pre-legend strength dps
		preLegendTankCloak,  // pre-legend strength tank

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
	})

	protMitigationNoSet.ForceSingleSlot(items.Equip_Ring1, 96481) // durumu
	protMitigationNoSet.ForceSingleSlot(items.Equip_Back, preLegendTankCloak)
	blockHelmetsWithoutIndomitable(&protMitigationNoSet)

	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark) // zandalar trinket

	// protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)
	// protMitigationNoSet.BlockItem(trinketFortZand)
	// protMitigationNoSet.AddReportVariant(items.Equip_Trinket2, trinketFortZand)

	protMitigationNoSet.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)
	protMitigationNoSet.BlockItem(trinketPrimRage)
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	protMitigationWithSet := multi_types.MultiSetParam{
		Label:    "Prot-Mitigation-WithSet",
		GearFile: files.GearFileProtMitigationSet,
		Model:    model.Model_PallyProtMitigation_WithSet(),
		// IncludeInFirstPass:        true,
		RequestRatingPercent:      0.30,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
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

		preLegendTankCloak, // pre-legend strength tank

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
		95290, // prot tier15 chest normal
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

		trinketZandSpark,
		trinketPrimRage,
		trinketJiKun,
		trinketTwinsGaze,
		trinketSoulBarrier,
		trinketFortZand,

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic
	})

	blockHelmetsWithoutIndomitable(&protMitigationWithSet)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Ring1, 96481)               // durumu
	protMitigationWithSet.ForceSingleSlot(items.Equip_Back, preLegendTankCloak)   // pre-legend strength tank
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark) // zandalar trinket

	// protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket2, trinketFortZand)  // Fortitude of the Zandalari
	// protMitigationWithSet.BlockItem(trinketPrimRage)
	// protMitigationWithSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	// protMitigationWithSet.ForceTryAllSlot(items.Equip_Trinket2, trinketFortZand, trinketSoulBarrier)
	protMitigationWithSet.ForceSingleSlot(items.Equip_Trinket2, trinketSoulBarrier) // fort good for taken, soul for death, but soul for raden for now
	protMitigationWithSet.BlockItem(trinketPrimRage)
	protMitigationWithSet.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	ret.AddBagsExtra()
	protDps.AddBagsExtra()
	protCompromise.AddBagsExtra()
	protMitigationNoSet.AddBagsExtra()
	protMitigationWithSet.AddBagsExtra()

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protCompromise)
	job.AddSetParam(protMitigationNoSet)
	job.AddSetParam(protMitigationWithSet)

	// job.AddItemDistinctUsageGroups(96550, []multi_types.MultiSetParam{ret, protDps, protCompromise}, []multi_types.MultiSetParam{protMitigationNoSet, protMitigationWithSet})

	job.FindHighsResult_Sample(1)
	// job.FindSeveralHighsAndSim()
	// job.FindHighsResultPerPermute(2)
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
	param.BlockItem(87024) // nullification greathelm = capacitance
	param.BlockItem(95282) // lightning emp helmet = capacitance
	// param.BlockItem(87101) // white tiger helmet = prot gem
	// param.BlockItem(95292) // lightning emp faceguard = prot gem
	param.BlockItem(95778) // golden golem celestial = ignore in all sets
	blockGeneral(param)
}

func blockGeneral(param *multi_types.MultiSetParam) {
	param.BlockItem(95513) // normal ring
}
