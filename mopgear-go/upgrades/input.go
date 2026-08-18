package upgrades

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type FindUpgrades_BasicInputs struct {
	IncludeCelestial    bool
	IncludeNormal       bool
	IncludeHeroic       bool
	PositiveResultsOnly bool
	WeightType          weight_types.WeightType
	TargetUpgradeLevel  items.UpgradeLevel
	IgnoredItems        []items.ItemId
	SolverTimeout       int
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
