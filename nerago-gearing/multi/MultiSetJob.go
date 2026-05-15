package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

const (
	simThreadCount   = 4
	highsThreadCount = 10
)

type MultiSetJob struct {
	printer             *util.PrintRecorder
	params              []multiSetParamInternal
	fixedForge          map[items.ItemId]stats.ReforgeRecipe
	distinctUsageGroups map[items.ItemId]distinctUsageGroups
	bagsGear            loaders.EquippedArray
	simRunSize          simulate.WowSim_RunSize
}

type distinctUsageGroups struct {
	groupAIndexes []int
	groupBIndexes []int
}

func MultiSetJob_Create(printer *util.PrintRecorder, simRunSize simulate.WowSim_RunSize) MultiSetJob {
	return MultiSetJob{
		printer:    printer,
		simRunSize: simRunSize,
	}
}

func (job *MultiSetJob) AddSetParam(param multi_types.MultiSetParam) {
	job.params = append(job.params, multiSetParamInternal{MultiSetParam: param, job: job})
	job.params[len(job.params)-1].init()
}

func (job *MultiSetJob) AddFixedForge(itemId items.ItemId, reforge stats.ReforgeRecipe) {
	if job.fixedForge == nil {
		job.fixedForge = make(map[items.ItemId]stats.ReforgeRecipe)
	}
	job.fixedForge[itemId] = reforge
}

func (job *MultiSetJob) AddItemDistinctUsageGroups(itemId items.ItemId, groupA []multi_types.MultiSetParam, groupB []multi_types.MultiSetParam) {
	usageGroups := distinctUsageGroups{}
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
	job.distinctUsageGroups[itemId] = usageGroups
}

func findLabelInParams(label string, group []multi_types.MultiSetParam) bool {
	for _, p := range group {
		if p.Label == label {
			return true
		}
	}
	return false
}
