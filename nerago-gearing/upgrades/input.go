package upgrades

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
)

type UpgradeGoal int8

const (
	UpgradeGoal_Mitigation  UpgradeGoal = iota
	UpgradeGoal_Dps         UpgradeGoal = iota
	UpgradeGoal_HalfMitiDps UpgradeGoal = iota
	UpgradeGoal_Healing     UpgradeGoal = iota
)

func (up UpgradeGoal) Name() string {
	switch up {
	case UpgradeGoal_Mitigation:
		return "miti"
	case UpgradeGoal_Dps:
		return "dps"
	case UpgradeGoal_HalfMitiDps:
		return "half"
	case UpgradeGoal_Healing:
		return "heal"
	default:
		panic("unknown")
	}
}

type FindUpgrades_BasicInputs struct {
	IncludeNormal       bool
	IncludeHeroic       bool
	IncludeRaden        bool
	PositiveResultsOnly bool
	IgnoredItems        []items.ItemId
	SolveSize           solver.SolveSize
}

type FindUpgrades_SimInputs struct {
	FindUpgrades_BasicInputs
	SimSize simulate.WowSim_RunSize
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
	Goal                    UpgradeGoal
	Model                   model.Model
	GearFile                string
	ItemFinder              func(stats.Difficulty) []*items.FullItem
	SubstituteItems         []items.ItemId
	SubstituteEmptySlotOnly map[items.SlotItem]items.ItemId
}
