package multi

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
)

const (
	simThreadCount   = 4
	highsThreadCount = 10
)

type MultiSetJob struct {
	printer                   *util.PrintRecorder
	params                    []multiSetParamInternal
	weightTypes               []int
	fixedForge                map[items.ItemId]stats.ReforgeRecipe
	distinctUsageGroups       map[items.ItemId]distinctUsageGroups
	alternateUpgradeChoices   [][]items.ItemId
	alternateGemming          []stats.GemInfo
	randomVariantItems        []randomVariantItem
	bagsGear                  loaders.EquippedArray
	simRunSize                simulate.WowSim_RunSize
	minimumExtraItemLevel     uint16
	writeBestToGearFiles      bool
	alternateGemmingAsPermute bool
}

type distinctUsageGroups struct {
	groupAIndexes       []int
	groupBIndexes       []int
	forceTryInEachParam bool
}

func MultiSetJob_Create(printer *util.PrintRecorder, simRunSize simulate.WowSim_RunSize) MultiSetJob {
	return MultiSetJob{
		printer:    printer,
		simRunSize: simRunSize,
	}
}

func (job *MultiSetJob) AddSetParam(param multi_types.MultiSetParam) {
	index := len(job.params)
	job.params = append(job.params, multiSetParamInternal{MultiSetParam: param, job: job, paramIndex: index})
	job.params[index].init()
}

func (job *MultiSetJob) AddFixedForge(itemId items.ItemId, reforge stats.ReforgeRecipe) {
	if job.fixedForge == nil {
		job.fixedForge = make(map[items.ItemId]stats.ReforgeRecipe)
	}
	job.fixedForge[itemId] = reforge
}

func (job *MultiSetJob) AddItemDistinctUsageGroups(itemId items.ItemId, forceTryInEachParam bool, groupA []multi_types.MultiSetParam, groupB []multi_types.MultiSetParam) {
	usageGroups := distinctUsageGroups{forceTryInEachParam: forceTryInEachParam}
	for paramIndex := range job.params {
		param := &job.params[paramIndex]

		inA := findLabelInParams(param.Label, groupA)
		inB := findLabelInParams(param.Label, groupB)
		if inA && inB {
			panic("in duplicate groups")
		} else if inA {
			usageGroups.groupAIndexes = append(usageGroups.groupAIndexes, paramIndex)
		} else if inB {
			usageGroups.groupBIndexes = append(usageGroups.groupBIndexes, paramIndex)
		} else {
			panic("in no groups")
		}
	}

	if job.distinctUsageGroups == nil {
		job.distinctUsageGroups = make(map[items.ItemId]distinctUsageGroups)
	}
	job.validateAddUsageGroups(itemId, usageGroups)
	job.distinctUsageGroups[itemId] = usageGroups
}

func (job *MultiSetJob) validateAddUsageGroups(addItemId items.ItemId, addGroup distinctUsageGroups) {
	if !addGroup.forceTryInEachParam {
		// only have trouble with duplicate forces
		return
	}

	addItem := db.WowSimDB_LoadItemById(addItemId, 0)
	for otherItemId, otherGroup := range job.distinctUsageGroups {
		if otherGroup.forceTryInEachParam {
			otherItem := db.WowSimDB_LoadItemById(otherItemId, 0)
			if addItem.SlotItem() == otherItem.SlotItem() {
				if anyInCommon(addGroup.groupAIndexes, otherGroup.groupAIndexes, otherGroup.groupBIndexes) ||
					anyInCommon(addGroup.groupBIndexes, otherGroup.groupAIndexes, otherGroup.groupBIndexes) {
					panic("same slot forced in multiple items/groups, try forceTryInEachParam=false")
				}
			}
		}
	}
}

func anyInCommon(checkSlice []int, otherASlice []int, otherBSlice []int) bool {
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

func (job *MultiSetJob) AddAlternateUpgradeChoices(itemIdList ...items.ItemId) {
	job.alternateUpgradeChoices = append(job.alternateUpgradeChoices, itemIdList)
}

func findLabelInParams(label string, group []multi_types.MultiSetParam) bool {
	for _, p := range group {
		if p.Label == label {
			return true
		}
	}
	return false
}

func (job *MultiSetJob) AddAlternateGemming(block stats.StatBlock) {
	gem := db.GemData_ByStat(&block)
	job.alternateGemming = append(job.alternateGemming, gem)
}

type randomVariantItem struct {
	itemId           items.ItemId
	upgradeLevel     items.UpgradeLevel
	randomSuffixList []items.RandomSuffix
}

// so initial processing is based on itemId only and we might have multiple versions of a rolled item
// force duplicate the item late in itemOptions build
func (job *MultiSetJob) MakeRandomVariants(itemId items.ItemId, upgradeLevel items.UpgradeLevel, randomSuffix ...items.RandomSuffix) {
	job.randomVariantItems = append(job.randomVariantItems,
		randomVariantItem{itemId, upgradeLevel, randomSuffix},
	)
}

func (job *MultiSetJob) SetMinimumExtraItemLevel(itemLevel uint16) {
	job.minimumExtraItemLevel = itemLevel
}

func (job *MultiSetJob) VerifyNoExtraDuplicates() {
	for param := range util_collection.ForPointer(job.params) {
		seen := make(map[items.ItemId]bool)
		for _, itemId := range param.ExtraItems {
			if seen[itemId] {
				panic("duplicate item " + itemId.String())
			} else {
				seen[itemId] = true
			}
		}
	}
}

func (job *MultiSetJob) RemoveAnyExtraDuplicates() {
	for param := range util_collection.ForPointer(job.params) {
		util_collection.RemoveDuplicatesComparable_InPlace(&param.ExtraItems)
	}
}

func (job *MultiSetJob) SetWriteBestToGearFiles() {
	job.writeBestToGearFiles = true
}

func (job *MultiSetJob) ActivateAlternateGemmingAsPermute() {
	job.alternateGemmingAsPermute = true
}

func (job *MultiSetJob) SetWeightTypes(weightTypes []int) {
	job.weightTypes = weightTypes
}
