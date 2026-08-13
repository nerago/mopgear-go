package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
)

const g_decimateTargetItemsPerSlot = 3

// normally decimate done as part of regular jobs
func (job *MultiSetJob) TestDecimate() {
	job.prepareItems()
	job.prepareWorking()
	job.runDecimate()
}

func (job *MultiSetJob) runDecimate() {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(job.working.Size())
	defer tracker.SetDone()

	workChannel := util_async.SeqToChannel(job.working.SeqValues())
	util_async.ForEach_Channel(c_decimateThreadCount, workChannel, func(work *specWorking) {
		work.runDecimateWork(tracker.NewChild(), job.printer, job.input.TimeLimitEachSolve)
	})
}

func (work *specWorking) runDecimateWork(tracker *util.TrackProgress, printer *util.PrintRecorder, timeout int) {
	solveOptions := items.SolvableOptionsMap_of(&work.itemPrep.itemOptions)
	solverModel := solve_highs_types.SolverModelBuild(&work.itemPrep.model, work.weightType, nil)

	futureBaseItemSet := solver.LaunchSolve(new(solveOptions), solverModel, printer, work.weightType, timeout)

	if baseItemSet, gotBase := futureBaseItemSet.WaitForResult(); gotBase {
		estimateSteps := uint64(g_decimateTargetItemsPerSlot - 1)
		currentStep := new(uint64)
		tracker.RunFromInt(currentStep, estimateSteps)

		bestBySlot := items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]]{}

		for i, item := range baseItemSet.Items() {
			if item != nil {
				slotEquip := items.SlotEquip(i)
				bestBySlot.Put(slotEquip, new(util_collection.SetComparable[items.ItemId]))
				bestBySlot.GetOrPanic(slotEquip).AddIfMissing(item.ItemId())
				work.decimateFindSlotBestN(&bestBySlot, slotEquip, &solveOptions, solverModel, printer, timeout)
				work.decimateRestoreSetBonusItems(&bestBySlot, slotEquip, &solveOptions, solverModel)
				tracker.NewChild().SetDone()
				*currentStep++
			}
		}

		work.decimateApply(&bestBySlot)
	} else {
		printer.Println("Decimate initial set failed " + work.itemPrep.label)
	}
	tracker.SetDone()
}

func (work *specWorking) decimateFindSlotBestN(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) {
	for bestBySlot.GetOrPanic(slotEquip).Size() < g_decimateTargetItemsPerSlot {
		restrictedOptions, isValid := work.decimatePrepareItemOptions(bestBySlot, slotEquip, solveOptionsBase)
		if !isValid {
			return
		}

		futureCheckSet := solver.LaunchSolve(&restrictedOptions, solverModel, printer, work.weightType, timeout)

		if checkSet, gotCheck := futureCheckSet.WaitForResult(); gotCheck {
			nextBestItem := checkSet.Items().Get(slotEquip)
			if nextBestItem == nil {
				return
			}
			bestBySlot.GetOrPanic(slotEquip).AddIfMissing(nextBestItem.ItemId())
		} else {
			return // set was probably infeasible without removed items, bail out
		}
	}
}

func (work *specWorking) decimatePrepareItemOptions(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap) (items.SolvableOptionsMap, bool) {
	restrictedOptions := solveOptionsBase.Clone()
	for removeId := range bestBySlot.GetOrNilValue(slotEquip).SeqValues() {
		if restrictedOptions.RemoveItemIdFromAll(removeId) {
			//slot got emptied (either this or paired), bail out
			return items.SolvableOptionsMap{}, false
		}
	}
	return restrictedOptions, true
}

func (work *specWorking) decimateRestoreSetBonusItems(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slot items.SlotEquip, optionsMap *items.SolvableOptionsMap, model *solve_highs_types.SolverModel) {
	for _, item := range optionsMap.Get(slot) {
		_, itemHasBonus := model.SetBonusIndexForItem(item.ItemId())
		if itemHasBonus {
			bestBySlot.GetOrPanic(slot).AddIfMissing(item.ItemId())
		}
	}
}

func (work *specWorking) decimateApply(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]]) {
	work.itemOptionsWork = work.itemPrep.itemOptions.Clone()
	for slot, idSet := range bestBySlot.SeqKeyValue() {
		work.itemOptionsWork.FilterSlot(slot, func(item *items.FullItem) bool {
			return idSet.HasValue(item.ItemId())
		})
	}
}
