package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/simulate"
)

func PaladinMultiRun() {
	job := multi.MultiSetJob{}

	protDps := multi.MultiSetParam{
		Label:                "Prot-Damage",
		GearFile:             gearFileProtDps,
		Model:                model.Model_PallyProtDps(),
		IncludeInFirstPass:   true,
		RequestRatingPercent: 0.45,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	protDps.AddExtraItems([]uint32{
		86957, // heroic bladed tempest ring
		95140, // shado assault band
		//                                86946, // ruby signet heroic
		95513, // scaled tyrant normal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		96373, // cloudbreaker belt heroic
		87024, // null greathelm
		94942, // hydra bloodcloak

		87026, // heroic peacock cloak
		86955, // heroic overwhelm assault belt
		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		95910, // ret tier15 chest celestial
		//      95281, // ret tier15 gloves normal
		//      95914, // ret tier15 shoulder celestial

		//      96667, // prot tier15 leg heroic
		//      96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		//      87036, // soulgrasp heroic
		94776, // primal turtle amulet

		96182, // ultimate prot of the emperor thunder

		//      87145, // defiled earth
		//      89934, // soul bracer
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
	})
	protDps.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal

	protMitigation := multi.MultiSetParam{
		Label:                "Prot-Mitigation",
		GearFile:             gearFileProtMitigation,
		Model:                model.Model_PallyProtMitigation(),
		IncludeInFirstPass:   true,
		RequestRatingPercent: 0.53,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	protMitigation.AddExtraItems([]uint32{
		86979, // heroic impaling treads
		87015, // heroic clawfeet

		86957, // heroic bladed tempest ring
		//      86946, // ruby signet heroic
		95140, // shado assault band
		95513, // scaled tyrant normal

		96373, // cloudbreaker belt heroic
		//      86955, // heroic overwhelm assault belt

		87026, // heroic peacock cloak
		//      86325, // daybreak
		94942, // hydra bloodcloak

		95535, // normal lightning legs
		//      94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		95291, // prot tier15 hand normal
		95920, // prot tier15 chest celestial
		95292, // prot tier15 head normal
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		//      95178, // lootraptor amulet
		//      94776, // primal turtle amulet

		96182, // ultimate prot of the emperor thunder

		87145, // defiled earth
		//      89934, // soul bracer
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
	})
	protMitigation.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal

	ret := multi.MultiSetParam{
		Label:                "Ret",
		GearFile:             gearFileRet,
		Model:                model.Model_PallyRet(),
		IncludeInFirstPass:   false,
		RequestRatingPercent: 0.02,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	ret.AddExtraItems([]uint32{
		//      87026, // heroic peacock cloak
		94942, // hydra bloodcloak

		95140, // shado assault band
		86957, // heroic bladed tempest ring
		95513, // scaled tyrant normal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		87024, // null greathelm
		//      86955, // heroic overwhelm assault belt
		96373, // cloudbreaker belt heroic

		95535, // normal lightning legs
		//      94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		85340, // ret tier14 legs
		//                        87101, // ret tier14 head [would need to regem, AVOID]
		85339, // ret tier14 shoulder
		85343, // ret tier14 chest
		87100, // ret tier14 hands

		//      95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		//      95912, // ret tier15 celestial leg/head (don't have yet)
		//      95913, // ret tier15 celestial leg/head (don't have yet)
		95914, // ret tier15 shoulder celestial

		95142, // striker's battletags
		95205, // terra-cotta neck
		94776, // primal turtle amulet

		//      87145, // defiled earth
		//      89934, // soul bracer
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic
	})
	ret.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protMitigation)

	// job.SuggestCulls(500, 10)
	// job.SuggestCulls(5000, 10)

	job.FindTopAndPassToSim(50000, 30, simulate.RunSize_Medium)
}
