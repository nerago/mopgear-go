package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/util"
)

func PaladinMultiRun(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer)

	ret := multi.MultiSetParam{
		Label:                "Ret",
		GearFile:             files.GearFileRet,
		Model:                model.Model_PallyRet(),
		IncludeInFirstPass:   false,
		RequestRatingPercent: 0.02,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2,
		// ForceUpgradeExistingItems: 2,
	}
	ret.AddExtraItems([]items.ItemId{
		98147, // pre-legend strength dps

		95140, // shado assault band
		86957, // heroic bladed tempest ring
		95513, // scaled tyrant normal
		// 94462, // pvp ring REMOVE1
		96481, // durumu tentacle heroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		87024, // null greathelm
		// 86955, // heroic overwhelm assault belt REMOVE1
		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic
		96478, // treads of the blind heroic

		// 95535, // normal lightning legs REMOVE1
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		// 95098, // Sightblinder Shoulderguards REMOVE1
		// 95778, // golden golem celestial [would need gem, but otherwise good]

		85343, // white tiger battleplate t14
		96447, // rot-proof greatplate
		87100, // White Tiger Gauntlets
		87101, // White Tiger Helmet [wrong gem]
		// 95292, // prot tier15 head normal [wrong gem] // TODO meta gem model

		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic
		//    95912, // ret tier15 celestial head (don't have yet)

		95142, // striker's battletags
		95205, // terra-cotta neck
		94776, // primal turtle amulet
		95178, // lootraptor amulet

		87145, // defiled earth
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic
	})
	// ret.AddBagsExtra()
	ret.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	ret.AddFixedSlot(items.Equip_Ring2, 95513)    // scaled tyrant normal
	ret.AddFixedSlot(items.Equip_Back, 98147)     // pre-legend strength dps
	job.AddSetParam(ret)

	protDps := multi.MultiSetParam{
		Label:                     "Prot-Damage",
		GearFile:                  files.GearFileProtDps,
		Model:                     model.Model_PallyProtDps(),
		IncludeInFirstPass:        true,
		RequestRatingPercent:      0.45,
		PhasedAcceptable:          false,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protDps.AddExtraItems([]items.ItemId{
		86957, // heroic bladed tempest ring REMOVE1
		95140, // shado assault band
		95513, // scaled tyrant normal
		96481, // durumu tentacle heroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic
		87024, // null greathelm

		98147, // pre-legend strength dps

		96428, // shell-coated wrists  REMOVE1
		// 96447, // rot-proof greatplate REMOVE1

		// 86955, // heroic overwhelm assault belt REMOVE1
		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		96478, // treads of the blind heroic
		95153, // Tyrant King Battleplate
		// 95778, // crown golden golem celestial [would need gem, acceptable]

		95912, // ret tier15 celestial head (don't have yet)
		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic

		// 95290, // prot tier15 chest normal
		// 95291, // prot tier15 hand normal
		// 95292, // prot tier15 head normal [defensive gem, avoid]
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		// 87036, // soulgrasp heroic REMOVE1
		94776, // primal turtle amulet

		87145, // defiled earth
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
		96534, // qon's scimitar

		94527, // ji-kun trinket
		94529, // gaze of the twins
	})
	// protDps.AddBagsExtra()                            // TODO allow ignore trinkets
	protDps.AddFixedSlot(items.Equip_Ring2, 95513)    // scaled tyrant normal
	protDps.AddFixedSlot(items.Equip_Offhand, 96182)  // ultimate prot of the emperor thunder normal
	protDps.AddFixedSlot(items.Equip_Back, 98147)     // pre-legend strength dps
	protDps.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	job.AddSetParam(protDps)

	protMitigation := multi.MultiSetParam{
		Label:                     "Prot-Mitigation",
		GearFile:                  files.GearFileProtMitigation,
		Model:                     model.Model_PallyProtMitigation(),
		IncludeInFirstPass:        true,
		RequestRatingPercent:      0.53,
		PhasedAcceptable:          false,
		ExtraUpgradeLevel:         2,
		ForceUpgradeExistingItems: 2,
	}
	protMitigation.AddExtraItems([]items.ItemId{
		86979, // heroic impaling treads
		// 87015, // heroic clawfeet
		96478, // treads of the blind heroic

		86957, // heroic bladed tempest ring
		95140, // shado assault band
		95513, // scaled tyrant normal
		96481, // durumu tentacle heroic

		96373, // cloudbreaker belt heroic
		// 86955, // heroic overwhelm assault belt REMOVE1

		98146, // pre-legend strength tank

		// 95535, // normal lightning legs
		// 94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		96428, // shell-coated wrists
		96447, // rot-proof greatplate

		// 95910, // ret tier15 chest celestial   REMOVE12
		95281, // ret tier15 gloves normal    REMOVE12
		96657, // ret tier15 legs heroic   REMOVE1
		96658, // ret tier15 shoulder heroic  REMOVE1

		95291, // prot tier15 hand normal
		95290, // prot tier15 chest normal
		95292, // prot tier15 head normal
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		95178, // lootraptor amulet
		// 94776, // primal turtle amulet

		87145, // defiled earth
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
		96534, // qon's scimitar
	})
	// protMitigation.AddBagsExtra()
	protMitigation.AddFixedSlot(items.Equip_Trinket1, 96398) // zandalar trinket
	protMitigation.AddFixedSlot(items.Equip_Ring1, 96481)    // durumu
	protMitigation.AddFixedSlot(items.Equip_Offhand, 94945)  // greatshield of the gloaming normal
	protMitigation.AddFixedSlot(items.Equip_Back, 98146)     // pre-legend strength tank
	job.AddSetParam(protMitigation)

	// job.AddSuppressSlotCheck(86946) // Vizier's Ruby Signet
	// job.AddSuppressSlotCheck(86957) // Ring of the Bladed Tempest
	// job.AddSuppressSlotCheck(95140) // Band of the Shado-Pan Assault
	// job.AddSuppressSlotCheck(96481) // Durumu's Severed Tentacle

	job.AddSpecificAllowRate(94527, 0.30) // ji-kun trinket

	// job.AddSuppressSlotCheck(94529) // Gaze of the Twins
	// job.AddSuppressSlotCheck(94527) // Ji-Kun's Rising Winds

	// job.SuggestCulls(100, 2)
	// job.SuggestCulls(1000, 40)
	//job.SuggestCulls(5000, 10)
	// job.SuggestCulls(10000, 100)
	// job.SuggestCulls(25000, 150)
	job.SuggestCulls(150000, 200)

	// job.FindTopAndPassToSim(50, 2, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(150000, 50, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(75000, 20, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(150000, 25, simulate.RunSize_Medium)
}
