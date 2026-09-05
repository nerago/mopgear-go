package multi_types

import (
	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type JobInputs struct {
	Param                []SpecParam
	SimRunSize           simulate.WowSim_RunSize
	WriteBestToGearFiles bool
	Shared               ItemShared
	TimeLimitEachSolve   int
}

type AlternateMode uint8

const (
	AlternateModeNone                 AlternateMode = iota
	AlternateModeReforgeBlocks        AlternateMode = iota
	AlternateModeItemAndReforgeBlocks AlternateMode = iota
)

type JobInputTask struct {
	WeightTypeList       []weight_types.WeightType
	AlsoExistingEquipped bool
	AlsoSpecOptimums     bool

	Alternates              AlternateMode
	AlternatesLimit         util_collection.Optional[int]
	IncludeInterimResults   bool
	RunDecimate             bool
	ReforgingAllowNonCommon bool
	AlternateGemList        []stats.GemInfo
	Permute                 InputPermute
}

type ItemShared struct {
	FixedForge            map[items.ItemId]stats.ReforgeRecipe
	RandomVariantItems    []RandomVariantItem
	MinimumExtraItemLevel uint16
}

type InputPermute struct {
	DistinctUsageGroups          map[items.ItemId]*DistinctUsageGroups
	AlternateUpgradeChoices      [][]items.ItemId
	AlternateAddItems            [][]items.ItemId
	PermuteOnItemCountOptions    bool
	AlternateGemsEnableAsPermute bool
}

type DistinctUsageGroups struct {
	GroupALabels        []string
	GroupBLabels        []string
	ForceTryInEachParam bool
}

func (ji *JobInputs) SetSimSize(simSize simulate.WowSim_RunSize) {
	ji.SimRunSize = simSize
}

func (ji *JobInputs) SetTimeLimitEachSolver(timeLimitSeconds int) {
	ji.TimeLimitEachSolve = timeLimitSeconds
}

func (ji *JobInputs) AddSpecParam(param SpecParam) {
	for _, other := range ji.Param {
		if other.Label == param.Label {
			panic("duplicate label")
		}
	}
	param.Model.InitDerives()
	ji.Param = append(ji.Param, param)
}

func (ji *JobInputs) GetSetParam(label string) *SpecParam {
	for param := range util_collection.ForPointer(ji.Param) {
		if param.Label == label {
			return param
		}
	}
	return nil
}

func (ji *JobInputs) AddFixedForge(itemId items.ItemId, reforge stats.ReforgeRecipe) {
	if ji.Shared.FixedForge == nil {
		ji.Shared.FixedForge = make(map[items.ItemId]stats.ReforgeRecipe)
	}
	ji.Shared.FixedForge[itemId] = reforge
}

type RandomVariantItem struct {
	ItemId           items.ItemId
	UpgradeLevel     items.UpgradeLevel
	RandomSuffixList []items.RandomSuffix
}

// so initial processing is based on itemId only and we might have multiple versions of a rolled item
// force duplicate the item late in itemOptions build
func (ji *JobInputs) MakeRandomVariants(itemId items.ItemId, upgradeLevel items.UpgradeLevel, randomSuffix ...items.RandomSuffix) {
	ji.Shared.RandomVariantItems = append(ji.Shared.RandomVariantItems,
		RandomVariantItem{itemId, upgradeLevel, randomSuffix},
	)
}

func (ji *JobInputs) SetMinimumExtraItemLevel(itemLevel uint16) {
	ji.Shared.MinimumExtraItemLevel = itemLevel
}

func (ji *JobInputs) VerifyNoExtraDuplicates() {
	for param := range util_collection.ForPointer(ji.Param) {
		seen := make(map[items.ItemId]bool)
		for _, itemId := range param.ItemInputs.ExtraItems {
			if seen[itemId] {
				panic("duplicate item " + itemId.String())
			} else {
				seen[itemId] = true
			}
		}
	}
}

func (ji *JobInputs) RemoveAnyExtraDuplicates() {
	for param := range util_collection.ForPointer(ji.Param) {
		util_collection.RemoveDuplicatesComparable_InPlace(&param.ItemInputs.ExtraItems)
	}
}

func (ji *JobInputs) SetWriteBestToGearFiles(writeBestToGearFiles bool) {
	ji.WriteBestToGearFiles = writeBestToGearFiles
}

func (task *JobInputTask) AddItemDistinctUsageGroups(itemId items.ItemId, forceTryInEachParam bool, groupA []SpecParam, groupB []SpecParam) {
	usageGroups := &DistinctUsageGroups{ForceTryInEachParam: forceTryInEachParam}
	usageGroups.GroupALabels = util_collection.MapSliceAsNew(groupA, func(x *SpecParam) string { return x.Label })
	usageGroups.GroupBLabels = util_collection.MapSliceAsNew(groupB, func(x *SpecParam) string { return x.Label })

	if task.Permute.DistinctUsageGroups == nil {
		task.Permute.DistinctUsageGroups = make(map[items.ItemId]*DistinctUsageGroups)
	}
	task.Permute.DistinctUsageGroups[itemId] = usageGroups
}

func (task *JobInputTask) AddAlternateUpgradeChoices(itemIdList ...items.ItemId) {
	task.Permute.AlternateUpgradeChoices = append(task.Permute.AlternateUpgradeChoices, itemIdList)
}

func (task *JobInputTask) AddAlternateItemsChoices(itemId ...items.ItemId) {
	task.Permute.AlternateAddItems = append(task.Permute.AlternateAddItems, itemId)
}

func (task *JobInputTask) AddAlternateGem(block stats.StatBlock) {
	gem := db.GemData_ByStat(&block)
	task.AlternateGemList = append(task.AlternateGemList, gem)
}

func (task *JobInputTask) ActivateAlternateGemsEnableAsPermute() {
	task.Permute.AlternateGemsEnableAsPermute = true
}

func (task *JobInputTask) SetWeightTypes(weightTypeList ...weight_types.WeightType) {
	task.WeightTypeList = weightTypeList
}

func (task *JobInputTask) SetReforgingAllowNonCommon(reforgingAllowNonCommon bool) {
	task.ReforgingAllowNonCommon = reforgingAllowNonCommon
}

func (task *JobInputTask) EnablePermuteOnItemCountOptions() {
	task.Permute.PermuteOnItemCountOptions = true
}
