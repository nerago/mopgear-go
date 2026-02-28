package multi

import (
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"slices"
)

func (job *MultiSetJob) prepareInitial() {
	job.printer.Println("LOADING BAGS")
	job.bagsGear = loaders.BagsFileReader_Read()

	job.printer.Println("PREPARING STARTING GEAR")
	for _, param := range job.params {
		param.prepareStartingGear()
	}

	job.printer.Println("PREPARING EXTRA ITEMS")
	for _, param := range job.params {
		param.prepareExtraItems()
	}

	job.printer.Println("RESTRICTING ANY FIXED SLOTS")
	for _, param := range job.params {
		param.restrictFixed()
	}

	job.validateMultiSetAlignItemSlots()

	for _, param := range job.params {
		job.printer.Printf("BASELINE for %s\n", param.Label)
		param.runBaseline()
	}

	job.prepareRatingMultipliers()
}

func (param *MultiSetParam) prepareStartingGear() {
	param.job.printer.Println(param.Label)

	equipped := loaders.GearFileReader_Read(param.GearFile)
	param.exactEquippedGear = setup.OptionsSetup_ExactEquippedOnly(equipped, &param.Model, &param.job.printer)
	param.itemOptions = setup.OptionsSetup_FromEquipped(equipped, &param.Model, &param.job.printer)
}

func (param *MultiSetParam) prepareExtraItems() {
	for _, itemId := range param.extraItems {
		param.includeExtra(itemId)
	}
	// TODO fixed slots
}

func (param *MultiSetParam) includeExtra(itemId uint32) {
	if param.itemOptions.IncludesItemId(itemId) {
		param.job.printer.Printf("EXTRA already included %d\n", itemId)
		return
	}

	if param.extraFromOtherSpec(itemId) {
		return
	}

	if param.extraFromBags(itemId) {
		return
	}

	param.extraLoadAndGenerate(itemId)
}

func (param *MultiSetParam) extraFromOtherSpec(itemId uint32) bool {
	options := make([]items.FullItem, 0)
	for _, otherParam := range param.job.params {
		more := otherParam.itemOptions.FindItemId(itemId)
		options = slices.AppendSeq(options, more)
	}

	options = util.RemoveDuplicatesFunc(options, (*items.FullItem).Equals)

	if len(options) > 1 {
		panic("expected multiple options")
	} else if len(options) == 1 {
		param.itemOptions.AddOneOption(options[0])
		param.job.printer.Printf("OPTION from other spec %s\n", options[0].String())
		return true
	} else {
		return false
	}
}

func (param *MultiSetParam) extraFromBags(itemId uint32) bool {
	for _, equipped := range param.job.bagsGear {
		if equipped.ItemId == itemId {
			// bags file doesn't have upgrade steps
			equipped.UpgradeStep = param.ExtraUpgradeLevel

			options, baseItem := setup.OptionsSetup_FromEquipped_Single(equipped, &param.Model, &param.job.printer)
			param.itemOptions.AddSeveralOptions(baseItem.Slot, options)
			param.job.printer.Printf("OPTION from bags %s\n", baseItem.String())
			return true
		}
	}
	return false
}

func (param *MultiSetParam) extraLoadAndGenerate(itemId uint32) {
	options, baseItem := setup.OptionsSetup_FromIdOnlyUseAllDefaults(itemId, param.ExtraUpgradeLevel, &param.Model, &param.job.printer)
	param.itemOptions.AddSeveralOptions(baseItem.Slot, options)
	param.job.printer.Printf("OPTION %s\n", baseItem.String())
}

func (param *MultiSetParam) restrictFixed() {
	for slot, itemId := range param.fixedSlots {
		if !param.itemOptions.Has(slot) {
			panic("restricting slot but already empty")
		}

		param.itemOptions.MapSlot(slot, func(options []items.FullItem) []items.FullItem {
			return util.FilterSlice(options, func(x *items.FullItem) bool {
				return x.ItemId() == itemId
			})
		})

		if !param.itemOptions.Has(slot) {
			panic("restricting slot leaves slot empty")
		}
	}
}

func (job *MultiSetJob) validateMultiSetAlignItemSlots() {
	seen := make(map[uint32]items.SlotEquip)
	for _, param := range job.params {
		for slot, item := range param.itemOptions.AllItemsWithSlot() {
			seenSlot, found := seen[item.ItemId()]
			if found && seenSlot != slot {
				panic("duplicate in non-matching slot " + item.String())
			} else if !found {
				seen[item.ItemId()] = slot
			}
		}
	}
}

func (param *MultiSetParam) runBaseline() {
	param.baselineResult = solver.Solver(&param.itemOptions, &param.Model, param.PhasedAcceptable)
	param.job.printer.AppendOther(&param.baselineResult.Printer)
	if !param.baselineResult.Success {
		panic("failed to find baseline for " + param.Label)
	}
	param.seenInSolutions.Add(&param.baselineResult.FullSet)
}

func (job *MultiSetJob) prepareRatingMultipliers() {
	var totalPercent float64
	for _, param := range job.params {
		param.prepareRatingMultiplier()
		totalPercent += param.RequestRatingPercent
	}

	if totalPercent < 0.99 || totalPercent > 1.01 {
		panic("percents don't add to one")
	}
}

func (param *MultiSetParam) prepareRatingMultiplier() {
	var targetCombined float64 = 1000000000000000000.0
	baselineRating := float64(param.baselineResult.ResultRating)
	if baselineRating > targetCombined/100.0 {
		panic("need bigger ratings")
	}

	targetForThis := targetCombined * param.RequestRatingPercent
	multiplyRatingsBy := targetForThis / baselineRating
	param.ratingMultiply = uint64(math.Round(multiplyRatingsBy))

	param.job.printer.Printf("MULTIPLIERS %d base=%d mult=%d value=%d", param.Label, param.baselineResult.ResultRating, param.ratingMultiply, uint64(math.Round(baselineRating*float64(param.ratingMultiply))))
}
