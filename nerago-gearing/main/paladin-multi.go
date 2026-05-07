package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
)

var dontRemoveMyImportPlease = simulate.RunSize_TestOnly

func PaladinMultiRun(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer, solver.SolveSize_PerItem, solver.SolveSize_Medium)
	// job := multi.MultiSetJob_Create(printer, solver.SolveSize_PerItem, solver.SolveSize_PerItem)

	ret := multi.MultiSetParam{
		Label:                "Ret",
		GearFile:             files.GearFileRet,
		Model:                model.Model_PallyRet(),
		IncludeInFirstPass:   false,
		RequestRatingPercent: 0.01,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2,
		// ForceUpgradeExistingItems: 2,
	}
	ret.AddExtraItems([]items.ItemId{
		98147, // pre-legend strength dps

		95140, // shado assault band
		95141, // shado assault loop
		86957, // heroic bladed tempest ring
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic

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

		94527, // ji-kun trinket
		94529, // gaze of the twins

		87145, // defiled earth OFF
		96394, // frozen warlord bracer heroic
	})
	ret.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	ret.AddFixedSlot(items.Equip_Ring2, 96500)    // scaled tyrant heroic
	ret.AddFixedSlot(items.Equip_Back, 98147)     // pre-legend strength dps
	blockHelmetsWithoutCapacitance(&ret)

	protDps := multi.MultiSetParam{
		Label:                     "Prot-Damage",
		GearFile:                  files.GearFileProtDps,
		Model:                     model.Model_PallyProtDps(),
		IncludeInFirstPass:        true,
		RequestRatingPercent:      0.04,
		PhasedAcceptable:          false,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protDps.AddExtraItems([]items.ItemId{
		86957, // heroic bladed tempest ring
		95140, // shado assault band
		// 95141, // shado assault loop OFF123
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		96478, // treads of the blind heroic

		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic

		98147, // pre-legend strength dps

		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate

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

		94527, // ji-kun trinket
		94529, // gaze of the twins

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic
	})
	protDps.AddFixedSlot(items.Equip_Ring2, 96500) // scaled tyrant heroic
	// protDps.AddFixedSlot(items.Equip_Offhand, 96182)  // ultimate prot of the emperor thunder normal
	protDps.AddFixedSlot(items.Equip_Back, 98147)     // pre-legend strength dps
	protDps.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	blockHelmetsWithoutCapacitance(&protDps)

	protCompromise := multi.MultiSetParam{
		Label:                     "Prot-Compromise",
		GearFile:                  files.GearFileProtCompromise,
		Model:                     model.Model_PallyProtCompromise(),
		IncludeInFirstPass:        true,
		RequestRatingPercent:      0.10,
		PhasedAcceptable:          false,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protCompromise.AddExtraItems([]items.ItemId{
		86957, // heroic bladed tempest ring
		95140, // shado assault band
		95141, // shado assault loop
		96500, // scaled tyrant heroic
		96481, // durumu tentacle heroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		96478, // treads of the blind heroic

		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic

		98147, // pre-legend strength dps

		96428, // shell-coated wrists
		96447, // rot-proof greatplate

		95535, // normal lightning legs
		94773, // centripetal shoulders normal REMOVE12
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		95153, // Tyrant King Battleplate REMOVE1234

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

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		94527, // ji-kun trinket
		94529, // gaze of the twins

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
		96436, // tortos shell heroic
	})
	protCompromise.AddFixedSlot(items.Equip_Ring2, 96500) // scaled tyrant heroic
	// protCompromise.AddFixedSlot(items.Equip_Offhand, 96182)  // ultimate prot of the emperor thunder normal
	protCompromise.AddFixedSlot(items.Equip_Back, 98147)     // pre-legend strength dps
	protCompromise.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	// protCompromise.AddFixedSlot(items.Equip_Trinket2, 94529) // gaze of the twins
	blockHelmetsWithoutCapacitance(&protCompromise)
	protCompromise.BlockItem(96793) // Fortitude of the Zandalari
	protCompromise.AddReportVariant(items.Equip_Trinket2, 96793) // Fortitude of the Zandalari

	protMitigationNoSet := multi.MultiSetParam{
		Label:                     "Prot-Mitigation-NoSet",
		GearFile:                  files.GearFileProtMitigationNoSet,
		Model:                     model.Model_PallyProtMitigation_NoSet(),
		IncludeInFirstPass:        true,
		RequestRatingPercent:      0.70,
		PhasedAcceptable:          false,
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

		96373, // cloudbreaker belt heroic

		98147, // pre-legend strength dps
		98146, // pre-legend strength tank

		95535, // normal lightning legs
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		96428, // shell-coated wrists UPGRADE LIKED
		96447, // rot-proof greatplate

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

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
	})
	protMitigationNoSet.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	protMitigationNoSet.AddFixedSlot(items.Equip_Ring1, 96481)    // durumu
	// protMitigationNoSet.AddFixedSlot(items.Equip_Offhand, 94945)  // greatshield of the gloaming normal
	protMitigationNoSet.AddFixedSlot(items.Equip_Trinket2, 96793) // Fortitude of the Zandalari
	protMitigationNoSet.AddReportVariant(items.Equip_Trinket2, 94529) // gaze of the twins
	blockHelmetsWithoutIndomitable(&protMitigationNoSet)

	protMitigationWithSet := multi.MultiSetParam{
		Label:                     "Prot-Mitigation-WithSet",
		GearFile:                  files.GearFileProtMitigationSet,
		Model:                     model.Model_PallyProtMitigation_WithSet(),
		IncludeInFirstPass:        true,
		RequestRatingPercent:      0.15,
		PhasedAcceptable:          false,
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

		96373, // cloudbreaker belt heroic

		98146, // pre-legend strength tank

		95535, // normal lightning legs REMOVE12
		94773, // centripetal shoulders normal REMOVE1234
		96468, // talonrender chest heroic REMOVE12
		96533, // rein-binders fists heroic
		96428, // shell-coated wrists REMOVE1234
		96447, // rot-proof greatplate

		96657, // ret tier15 legs heroic   REMOVE12
		96658, // ret tier15 shoulder heroic  REMOVE12

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

		94945, // greatshield of the gloaming normal
		96182, // ultimate prot of the emperor thunder normal
	})
	protMitigationWithSet.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	protMitigationWithSet.AddFixedSlot(items.Equip_Ring1, 96481)    // durumu
	// protMitigationSet.AddFixedSlot(items.Equip_Offhand, 94945)  // greatshield of the gloaming normal
	protMitigationWithSet.AddFixedSlot(items.Equip_Back, 98146) // pre-legend strength tank
	protMitigationWithSet.AddFixedSlot(items.Equip_Trinket2, 96793) // Fortitude of the Zandalari
	protMitigationWithSet.AddReportVariant(items.Equip_Trinket2, 94529) // gaze of the twins
	blockHelmetsWithoutIndomitable(&protMitigationWithSet)

	// job.AddSuppressSlotCheck(86946) // Vizier's Ruby Signet
	// job.AddSuppressSlotCheck(86957) // Ring of the Bladed Tempest
	// job.AddSuppressSlotCheck(95140) // Band of the Shado-Pan Assault
	// job.AddSuppressSlotCheck(96481) // Durumu's Severed Tentacle

	// job.AddSuppressSlotCheck(94529) // Gaze of the Twins
	// job.AddSuppressSlotCheck(94527) // Ji-Kun's Rising Winds
	job.AddSpecificAllowRate(94527, multi.Force_Forbidden, multi.Force_Optional, 0.20) // ji-kun trinket
	// job.AddSpecificAllowRate(96436, multi.Force_Optional, multi.Force_FixedWhereAvailable, 0.40) // tortos shell heroic

	// ret.AddBagsExtra()
	// protDps.AddBagsExtra()
	// protCompromise.AddBagsExtra()
	// protMitigationNoSet.AddBagsExtra()
	// protMitigationWithSet.AddBagsExtra()

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protCompromise)
	job.AddSetParam(protMitigationNoSet)
	job.AddSetParam(protMitigationWithSet)

	// job.SuggestCulls(100, 10)
	// job.SuggestCulls(1000, 40)
	// job.SuggestCulls(5000, 10)
	// job.SuggestCulls(10000, 100)
	// job.SuggestCulls(25000, 150)
	// job.SuggestCulls(150000, 200)

	// job.FindTopAndPassToSim(10, 1, false, simulate.RunSize_QuickDirty)
	// job.FindTopAndPassToSim(100, 10, true, simulate.RunSize_QuickDirty)
	// job.FindTopAndPassToSim(100, 10, true, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(5000, 30, false, simulate.RunSize_QuickDirty)
	// job.FindTopAndPassToSim(20000, 24, true, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(75000, 30, true, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(150000, 50, true, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(500000, 75, true, simulate.RunSize_Medium)

	// job.FindHighsResult()
	job.FindSeveralHighsAndSim(simulate.RunSize_Medium)

	
	// job.DetermineWhatRatingsLeadToResult(commonComboCurrent())
}

func blockHelmetsWithoutCapacitance(param *multi.MultiSetParam) {
	// param.BlockItem(87024) // nullification greathelm = capacitance
	// param.BlockItem(95282) // lightning emp helmet = capacitance
	param.BlockItem(87101) // white tiger helmet = prot gem
	param.BlockItem(95292) // lightning emp faceguard = prot gem
	param.BlockItem(95778) // golden golem celestial = ignore in all sets
	param.BlockItem(96666) // lightning emp faceguard heroic = prot gem
	blockGeneral(param)
}

func blockHelmetsWithoutIndomitable(param *multi.MultiSetParam) {
	param.BlockItem(87024) // nullification greathelm = capacitance
	param.BlockItem(95282) // lightning emp helmet = capacitance
	// param.BlockItem(87101) // white tiger helmet = prot gem
	// param.BlockItem(95292) // lightning emp faceguard = prot gem
	param.BlockItem(95778) // golden golem celestial = ignore in all sets
	blockGeneral(param)
}

func blockGeneral(param *multi.MultiSetParam) {
	param.BlockItem(95513) // normal ring
}
