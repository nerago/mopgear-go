package upgrades

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
)

type FindUpgrades_BasicInputs struct {
	IncludeCelestial    bool
	IncludeNormal       bool
	IncludeHeroic       bool
	PositiveResultsOnly bool
	WeightType          int
	TargetUpgradeLevel  items.UpgradeLevel
	IgnoredItems        []items.ItemId
}

type FindUpgrades_SimInputs struct {
	FindUpgrades_BasicInputs
	SimSizeBaseline              simulate.WowSim_RunSize
	SimSizeItemInitial           simulate.WowSim_RunSize
	ExtraSimForTopResultsCount   int
	ExtraSimForTopResultsSimSize simulate.WowSim_RunSize
}

type FindUpgrades_MultiSpec struct {
	FindUpgrades_BasicInputs
	Specs []FindUpgrades_Spec
}

type FindUpgrades_MultiSpec_Sim struct {
	FindUpgrades_SimInputs
	Specs []FindUpgrades_Spec
}

type FindUpgrades_Spec struct {
	Label                   string
	Model                   gear_model.SpecModel
	GearFile                string
	ItemFinder              func(stats.Difficulty) []loaders.ItemFoundRef
	SubstituteItems         []items.ItemId
	SubstituteEmptySlotOnly map[items.SlotItem]items.ItemId
}
