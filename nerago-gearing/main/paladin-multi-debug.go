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


func PaladinMultiRunDebug(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer, simulate.RunSize_QuickDirty)

	protDps := multi_types.MultiSetParam{
		Label:    "Prot-Damage",
		GearFile: files.GearFileProtDps,
		Model:    model.Model_PallyProtDps(),
		RequestRatingPercent:      0.60,
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

		trinketZandSpark,
		trinketPrimRage,
	})
	// blockHelmetsWithoutCapacitance(&protDps)
	// protDps.ForceSingleSlot(items.Equip_Ring2, 96500)               // scaled tyrant heroic
	// protDps.ForceSingleSlot(items.Equip_Back, preLegendMeleeCloak)  // pre-legend strength dps
	// protDps.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark) // zandalar trinket
	// protDps.ForceSingleSlot(items.Equip_Trinket2, trinketPrimRage)

	protCompromise := multi_types.MultiSetParam{
		Label:    "Prot-Compromise",
		GearFile: files.GearFileProtCompromise,
		Model:    model.Model_PallyProtCompromise(),
		RequestRatingPercent:      0.40,
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

		trinketZandSpark,
		trinketSoulBarrier,
	})
	// blockHelmetsWithoutCapacitance(&protCompromise)
	// protCompromise.ForceSingleSlot(items.Equip_Trinket1, trinketZandSpark)
	// protCompromise.ForceSingleSlot(items.Equip_Trinket2, trinketSoulBarrier)
	// protCompromise.AddReportVariant(items.Equip_Trinket2, trinketPrimRage)

	job.AddSetParam(protDps)
	job.AddSetParam(protCompromise)

	job.FindHighsResult_Sample(1)
}
