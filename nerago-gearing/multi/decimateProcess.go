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
		work.runDecimateWork(tracker.NewChild(), job.printer)
	})
}

func (work *specWorking) runDecimateWork(tracker *util.TrackProgress, printer *util.PrintRecorder) {
	solveOptions := items.SolvableOptionsMap_of(&work.itemPrep.itemOptions)
	solverModel := solve_highs_types.SolverModelBuild(&work.itemPrep.model, work.weightType, nil)

	futureBaseItemSet := solver.LaunchSolve(new(solveOptions), solverModel, printer, work.weightType)

	if baseItemSet, gotBase := futureBaseItemSet.WaitForResult(); gotBase {
		estimateSteps := uint64(g_decimateTargetItemsPerSlot - 1)
		currentStep := new(uint64)
		tracker.RunFromInt(currentStep, estimateSteps)

		bestBySlot := util_collection.MapSlice[items.SlotEquip, items.ItemId]{}

		for i, item := range baseItemSet.Items() {
			if item != nil {
				slotEquip := items.SlotEquip(i)
				bestBySlot.Add(slotEquip, item.ItemId())
				work.decimateFindSlotBestN(&bestBySlot, slotEquip, &solveOptions, solverModel, printer)
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

func (work *specWorking) decimateFindSlotBestN(bestBySlot *util_collection.MapSlice[items.SlotEquip, items.ItemId], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel, printer *util.PrintRecorder) {
	for bestBySlot.CountForKey(slotEquip) < g_decimateTargetItemsPerSlot {
		restrictedOptions := solveOptionsBase.Clone()
		for removeId := range bestBySlot.ValuesForKeyAsSeq(slotEquip) {
			if restrictedOptions.RemoveItemIdFromAll(removeId) {
				//slot got emptied (either this or paired), bail out
				return
			}
		}

		futureCheckSet := solver.LaunchSolve(&restrictedOptions, solverModel, printer, work.weightType)

		if checkSet, gotCheck := futureCheckSet.WaitForResult(); gotCheck {
			nextBestItem := checkSet.Items().Get(slotEquip)
			bestBySlot.Add(slotEquip, nextBestItem.ItemId())
		} else {
			return // set was probably infeasible without removed items, bail out
		}
	}
}

func (work *specWorking) decimateApply(bestBySlot *util_collection.MapSlice[items.SlotEquip, items.ItemId]) {
	work.itemOptionsWork = work.itemPrep.itemOptions.Clone()
	//for slot, idSlice := range bestBySlot.SeqGroupsInternalSlice() {
	//
	//}
}
