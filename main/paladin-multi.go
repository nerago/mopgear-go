package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
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
		86946, // ruby signet heroic
		95513, // scaled tyrant normal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		94726, // cloudbreaker belt
		87024, // null greathelm
		94942, // hydra bloodcloak

		87026, // heroic peacock cloak
		86955, // heroic overwhelm assault belt
		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		95914, // ret tier15 shoulder celestial

		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		87036, // soulgrasp heroic
		94776, // primal turtle amulet

		96182, // ultimate prot of the emperor thunder

		87145, // defiled earth
		89934, // soul bracer
		94820, // caustic spike bracers

		96376, // worldbreaker weapon})
	})
	protDps.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal
	job.AddSetParam(protDps)

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
		86946, // ruby signet heroic
		95140, // shado assault band
		95513, // scaled tyrant normal

		94726, // cloudbreaker belt
		86955, // heroic overwhelm assault belt

		87026, // heroic peacock cloak
		86325, // daybreak
		94942, // hydra bloodcloak

		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		95291, // prot tier15 hand normal
		95920, // prot tier15 chest celestial
		95292, // prot tier15 head normal
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		95178, // lootraptor amulet
		94776, // primal turtle amulet

		96182, // ultimate prot of the emperor thunder

		87145, // defiled earth
		89934, // soul bracer
		94820, // caustic spike bracers

		96376, // worldbreaker weapon
	})
	protMitigation.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal
	job.AddSetParam(protMitigation)

	ret := multi.MultiSetParam{
		Label:                "Ret",
		GearFile:             gearFileRet,
		Model:                model.Model_PallyRet(),
		IncludeInFirstPass:   false,
		RequestRatingPercent: 0.02,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	ret.AddExtraItems([]uint32{
		87026, // heroic peacock cloak
		94942, // hydra bloodcloak

		95140, // shado assault band
		86957, // heroic bladed tempest ring
		95513, // scaled tyrant normal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		87024, // null greathelm
		86955, // heroic overwhelm assault belt
		94726, // cloudbreaker belt

		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		85340, // ret tier14 legs
		// ret tier14 head [would need to regem, AVOID]
		85339, // ret tier14 shoulder
		85343, // ret tier14 chest
		87100, // ret tier14 hands

		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		// ret tier15 celestial (don't have yet)
		// ret tier15 celestial (don't have yet)
		95914, // ret tier15 shoulder celestial

		95142, // striker's battletags
		95205, // terra-cotta neck
		94776, // primal turtle amulet

		87145, // defiled earth
		89934, // soul bracer
		94820, // caustic spike bracers
	})
	ret.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal
	job.AddSetParam(ret)

	job.SuggestCulls(10000)
}
