package multi_types

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
)

type JobInputs struct {
	Param                []SpecParam
	WeightTypeList       []weight_types.WeightType
	SimRunSize           simulate.WowSim_RunSize
	WriteBestToGearFiles bool
	ItemInput            ItemInputShared
	// TODO consider a TimeLimitTotal
	TimeLimitEachSolve int
	RunDecimate        bool
}

type ItemInputShared struct {
	FixedForge                   map[items.ItemId]stats.ReforgeRecipe
	DistinctUsageGroups          map[items.ItemId]DistinctUsageGroups
	AlternateUpgradeChoices      [][]items.ItemId
	AlternateGemming             []stats.GemInfo
	RandomVariantItems           []RandomVariantItem
	MinimumExtraItemLevel        uint16
	AlternateGemsEnableAsPermute bool
	ReforgingAllowNonCommon      bool
	PermuteOnItemCountOptions    bool
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

func (ji *JobInputs) AddSetParam(param SpecParam) {
	for _, other := range ji.Param {
		if other.Label == param.Label {
			panic("duplicate label")
		}
	}
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
	if ji.ItemInput.FixedForge == nil {
		ji.ItemInput.FixedForge = make(map[items.ItemId]stats.ReforgeRecipe)
	}
	ji.ItemInput.FixedForge[itemId] = reforge
}

func (ji *JobInputs) AddItemDistinctUsageGroups(itemId items.ItemId, forceTryInEachParam bool, groupA []SpecParam, groupB []SpecParam) {
	usageGroups := DistinctUsageGroups{ForceTryInEachParam: forceTryInEachParam}
	for param := range util_collection.ForPointer(ji.Param) {
		inA := findLabelInParams(param.Label, groupA)
		inB := findLabelInParams(param.Label, groupB)
		if inA && inB {
			panic("in duplicate groups")
		} else if inA {
			usageGroups.GroupALabels = append(usageGroups.GroupALabels, param.Label)
		} else if inB {
			usageGroups.GroupBLabels = append(usageGroups.GroupBLabels, param.Label)
		} else {
			panic("in no groups")
		}
	}

	if ji.ItemInput.DistinctUsageGroups == nil {
		ji.ItemInput.DistinctUsageGroups = make(map[items.ItemId]DistinctUsageGroups)
	}
	ji.validateAddUsageGroups(itemId, usageGroups)
	ji.ItemInput.DistinctUsageGroups[itemId] = usageGroups
}

func (ji *JobInputs) validateAddUsageGroups(addItemId items.ItemId, addGroup DistinctUsageGroups) {
	if !addGroup.ForceTryInEachParam {
		// only have trouble with duplicate forces
		return
	}

	addItem := db.WowSimDB_LoadItemById(addItemId, 0)
	for otherItemId, otherGroup := range ji.ItemInput.DistinctUsageGroups {
		if otherGroup.ForceTryInEachParam {
			otherItem := db.WowSimDB_LoadItemById(otherItemId, 0)
			if addItem.SlotItem() == otherItem.SlotItem() {
				if anyInCommon(addGroup.GroupALabels, otherGroup.GroupALabels, otherGroup.GroupBLabels) ||
					anyInCommon(addGroup.GroupBLabels, otherGroup.GroupALabels, otherGroup.GroupBLabels) {
					panic("same slot forced in multiple items/groups, try forceTryInEachParam=false")
				}
			}
		}
	}
}

func anyInCommon(checkSlice []string, otherASlice []string, otherBSlice []string) bool {
	for _, check := range checkSlice {
		for _, a := range otherASlice {
			if check == a {
				return true
			}
		}
		for _, b := range otherBSlice {
			if check == b {
				return true
			}
		}
	}
	return false
}

func (ji *JobInputs) AddAlternateUpgradeChoices(itemIdList ...items.ItemId) {
	ji.ItemInput.AlternateUpgradeChoices = append(ji.ItemInput.AlternateUpgradeChoices, itemIdList)
}

func findLabelInParams(label string, group []SpecParam) bool {
	for _, p := range group {
		if p.Label == label {
			return true
		}
	}
	return false
}

func (ji *JobInputs) AddAlternateGemming(block stats.StatBlock) {
	gem := db.GemData_ByStat(&block)
	ji.ItemInput.AlternateGemming = append(ji.ItemInput.AlternateGemming, gem)
}

type RandomVariantItem struct {
	ItemId           items.ItemId
	UpgradeLevel     items.UpgradeLevel
	RandomSuffixList []items.RandomSuffix
}

// so initial processing is based on itemId only and we might have multiple versions of a rolled item
// force duplicate the item late in itemOptions build
func (ji *JobInputs) MakeRandomVariants(itemId items.ItemId, upgradeLevel items.UpgradeLevel, randomSuffix ...items.RandomSuffix) {
	ji.ItemInput.RandomVariantItems = append(ji.ItemInput.RandomVariantItems,
		RandomVariantItem{itemId, upgradeLevel, randomSuffix},
	)
}

func (ji *JobInputs) SetMinimumExtraItemLevel(itemLevel uint16) {
	ji.ItemInput.MinimumExtraItemLevel = itemLevel
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

func (ji *JobInputs) ActivateAlternateGemsEnableAsPermute() {
	ji.ItemInput.AlternateGemsEnableAsPermute = true
}

func (ji *JobInputs) SetWeightTypes(weightTypeList ...weight_types.WeightType) {
	ji.WeightTypeList = weightTypeList
}

func (ji *JobInputs) SetReforgingAllowNonCommon(reforgingAllowNonCommon bool) {
	ji.ItemInput.ReforgingAllowNonCommon = reforgingAllowNonCommon
}

func (ji *JobInputs) EnablePermuteOnItemCountOptions() {
	ji.ItemInput.PermuteOnItemCountOptions = true
}
