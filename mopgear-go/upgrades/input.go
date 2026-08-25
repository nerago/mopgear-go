package upgrades

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type InputSettings struct {
	IncludeCelestial             bool
	IncludeNormal                bool
	IncludeHeroic                bool
	PositiveResultsOnly          bool
	WeightType                   weight_types.WeightType
	TargetUpgradeLevel           items.UpgradeLevel
	IgnoredItems                 []items.ItemId
	SolverTimeout                int
	SimSizeBaseline              simulate.WowSim_RunSize
	SimSizeItemInitial           simulate.WowSim_RunSize
	ExtraSimForTopResultsCount   int
	ExtraSimForTopResultsSimSize simulate.WowSim_RunSize
}

type FindUpgradesMultiSpec struct {
	Settings InputSettings
	Specs    []SpecInput
}

type SpecInput struct {
	Label                   string
	Model                   gear_model.SpecModel
	GearFile                string
	ItemFinder              func(stats.Difficulty) []loaders.ItemFoundRef
	SubstituteItems         []items.ItemId
	SubstituteEmptySlotOnly map[items.SlotItem]items.ItemId
}
