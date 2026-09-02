package multi

import (
	"slices"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type specItemPrep struct {
	label             string
	model             *gear_model.SpecModel
	exactEquippedGear items.FullEquipMap
	itemOptions       items.FullOptionsMap
	addedFromBags     []items.ItemId
	inputs            *multi_types.ItemInputs
	seenInSolutions   *seenMap
}

func (job *MainJob) prepareItems() error {
	job.printer.Println("LOADING BAGS")
	job.bagsGear = loaders.BagsFileReader_Read()

	job.printer.Println("PREPARING STARTING GEAR")
	params := job.input.Param
	itemPrepSlice, err := util_collection.MapSliceAsNew_PassError(params, func(param *multi_types.SpecParam) (specItemPrep, error) {
		prep := specItemPrep{
			label:           param.Label,
			model:           &param.Model,
			inputs:          &param.ItemInputs,
			seenInSolutions: &seenMap{content: make(map[items.ItemId]uint32)},
		}
		err := prep.prepareStartingGear(&param.ItemInputs, &param.Model, job.printer)
		return prep, err
	})
	if err != nil {
		return err
	}

	job.itemPrep = util_collection.SliceToMap(itemPrepSlice,
		func(p *specItemPrep) string { return p.label },
		util_collection.IdentityFunc[*specItemPrep])

	refs := &extraRefs{
		bagsGear:   &job.bagsGear,
		shared:     &job.input.Shared,
		otherPreps: job.itemPrep,
		tasks:      job.tasks,
	}

	job.printer.Println("PREPARING EXTRA ITEMS")
	for i := range itemPrepSlice {
		itemPrepSlice[i].prepareExtraItems(&params[i].ItemInputs, refs, job.printer)
	}

	job.printer.Println("ADDITIONAL ITEM SETUP")
	for i := range itemPrepSlice {
		prep := &itemPrepSlice[i]
		input := &params[i].ItemInputs
		if err := prep.removeBlocked(input, job.printer); err != nil {
			return err
		}
		if err := prep.restrictFixed(input, job.printer); err != nil {
			return err
		}
		if err := prep.makeRandomVariantItems(job.input.Shared.RandomVariantItems, input, job.printer); err != nil {
			return err
		}
	}

	return nil
}

func (job *MainJob) paramOrderSlice() []string {
	return util_collection.MapSliceAsNew(job.input.Param, func(param *multi_types.SpecParam) string {
		return param.Label
	})
}

type extraRefs struct {
	bagsGear   *loaders.EquippedArray
	shared     *multi_types.ItemShared
	otherPreps map[string]*specItemPrep
	tasks      []multi_types.JobInputTask
}

func (ref extraRefs) findDistinctUsageGroupsForItem(itemId items.ItemId) *multi_types.DistinctUsageGroups {
	for _, task := range ref.tasks {
		if group, hasGroup := task.Permute.DistinctUsageGroups[itemId]; hasGroup {
			return group
		}
	}
	return nil
}

func (prep *specItemPrep) prepareStartingGear(input *multi_types.ItemInputs, model *gear_model.SpecModel, printer *util.PrintRecorder) (err error) {
	printer.Println(prep.label)

	equipped := loaders.GearFileReader_Read(input.GearFile)
	prep.exactEquippedGear, err = setup.OptionsSetup_FromEquipped_OriginalForgeOnly(equipped, model, input.MissingEnchant, printer)
	if err != nil {
		return err
	}
	prep.itemOptions, err = setup.OptionsSetup_FromEquipped(equipped, model, input.MissingEnchant, printer)
	if err != nil {
		return err
	}

	return setup.UpgradeAllOptionsToLevel2(&prep.itemOptions, input.ForceUpgradeExistingItems, model, printer)
}

func (prep *specItemPrep) prepareExtraItems(input *multi_types.ItemInputs, refs *extraRefs, printer *util.PrintRecorder) {
	printer.Println(prep.label)

	for _, itemId := range input.ExtraItems {
		prep.includeExtra(itemId, input, refs, printer)
	}

	if input.ExtraFromBags {
		for _, item := range *refs.bagsGear {
			prep.tryAddExtraFromBags(&item, input, refs, printer)
		}
	}

	prep.replicatePairedOptions()
}

func (prep *specItemPrep) replicatePairedOptions() {
	// risky compared to old algorithms, but now we have better unique equipped maybe ok?
	prep.replicatePairedSlots(items.Equip_Ring1, items.Equip_Ring2)
	prep.replicatePairedSlots(items.Equip_Trinket1, items.Equip_Trinket2)
	prep.itemOptions.RemoveDuplicates()
}

func (prep *specItemPrep) replicatePairedSlots(slotA items.SlotEquip, slotB items.SlotEquip) {
	optA := prep.itemOptions.Get(slotA)
	optB := prep.itemOptions.Get(slotB)
	combined := slices.Concat(optA, optB)
	prep.itemOptions[slotA] = combined
	prep.itemOptions[slotB] = combined
}

func (prep *specItemPrep) includeExtra(itemId items.ItemId, input *multi_types.ItemInputs, refs *extraRefs, printer *util.PrintRecorder) {
	if prep.itemOptions.IncludesItemId(itemId) {
		printer.Printf("EXTRA already included %d\n", itemId)
		return
	}

	if slices.Contains(input.BlockedItems, itemId) {
		printer.Printf("BLOCKED %d\n", itemId)
		return
	}

	basicVersion := db.WowSimDB_LoadItemById(itemId, 0)
	if prep.itemOptions.CouldAddUpgrade_ItemSlot(basicVersion.SlotItem(), basicVersion, printer, prep.model.SpecificIncompatibleList) == items.CanUpgrade_InvalidAlways {
		return
	}

	copyOk, err := prep.copyExtraFromOtherSpec(itemId, refs, printer)
	if err != nil {
		printer.Printf("EXTRA include %d ERROR: %v\n", itemId, err)
	} else if copyOk {
		return
	}

	if prep.copyExtraFromBags(itemId, refs, input.ExtraUpgradeLevel, printer) {
		return
	}

	prep.extraLoadAndGenerate(itemId, items.NO_RANDOM_SUFFIX, input, printer)
}

func (prep *specItemPrep) copyExtraFromOtherSpec(itemId items.ItemId, refs *extraRefs, printer *util.PrintRecorder) (bool, error) {
	usageGroups := refs.findDistinctUsageGroupsForItem(itemId)

	options := make([]items.FullItem, 0)
	if usageGroups == nil {
		for _, other := range refs.otherPreps {
			more := other.itemOptions.FindItemId(itemId)
			options = slices.AppendSeq(options, more)
		}
	} else {
		if slices.Contains(usageGroups.GroupALabels, prep.label) {
			for _, otherLabel := range usageGroups.GroupALabels {
				otherPrep := refs.otherPreps[otherLabel]
				more := otherPrep.itemOptions.FindItemId(itemId)
				options = slices.AppendSeq(options, more)
			}
		} else if slices.Contains(usageGroups.GroupBLabels, prep.label) {
			for _, otherLabel := range usageGroups.GroupBLabels {
				otherPrep := refs.otherPreps[otherLabel]
				more := otherPrep.itemOptions.FindItemId(itemId)
				options = slices.AppendSeq(options, more)
			}
		} else {
			return false, util.ErrorTracedNew("expected param label to be in one of the groups")
		}
	}

	util_collection.RemoveDuplicatesFunc_InPlace(&options, (*items.FullItem).Equals)

	// NOTE these may not copy with the model's reforge preferences etc

	if len(options) > 0 {
		if err := prep.addItemOptionsWithValidate(options[0].SlotItem(), options); err != nil {
			return false, err
		}
		printer.Printf("OPTION from other spec %s\n", options[0].CreateString())
		return true, nil
	} else {
		return false, nil
	}
}

func (prep *specItemPrep) copyExtraFromBags(itemId items.ItemId, refs *extraRefs, requestedUpgrade items.UpgradeLevel, printer *util.PrintRecorder) bool {
	equipped := refs.bagsGear.GetWithItemId(itemId)
	if equipped != nil {
		if requestedUpgrade > 0 {
			equipped.UpgradeStepOrItemLevel = int32(requestedUpgrade)
		}
		options, example, err := setup.OptionsSetup_OneItem_FromEquipped_AllForges(*equipped, prep.model, setup.MissingEnchant_Fix, printer)
		if err != nil {
			printer.Printf("OPTION from bags ERROR: %v\n", err)
			return false
		}
		if err := prep.addItemOptionsWithValidate(example.SlotItem(), options); err != nil {
			printer.Printf("OPTION from bags ERROR: %v\n", err)
			return false
		}
		printer.Printf("OPTION from bags %s\n", example.CreateString())
		return true
	}
	return false
}

func (prep *specItemPrep) tryAddExtraFromBags(equipped *loaders.EquippedItem, input *multi_types.ItemInputs, refs *extraRefs, printer *util.PrintRecorder) {
	if !db.WowSimDB_HasItemId(equipped.ItemId) {
		printer.Printf("UNKNOWN itemid IN bags %d\n", equipped.ItemId)
		return
	} else if prep.itemOptions.IncludesItemId(equipped.ItemId) {
		printer.Printf("EXTRA already included %d\n", equipped.ItemId)
		return
	} else if slices.Contains(input.BlockedItems, equipped.ItemId) {
		//printer.Printf("BLOCKED %d\n", equipped.ItemId)
		return
	}

	// bail early before considering full item stats/enchants/etc that might not fit spec
	basicVersion := db.WowSimDB_LoadItemById(equipped.ItemId, 0)
	if basicVersion.ItemLevel() < refs.shared.MinimumExtraItemLevel {
		return
	}
	if prep.itemOptions.CouldAddUpgrade_ItemSlot(basicVersion.SlotItem(), basicVersion, printer, prep.model.SpecificIncompatibleList) != items.CanUpgrade_Yes {
		return
	}

	copyOk, err := prep.copyExtraFromOtherSpec(equipped.ItemId, refs, printer)
	if err != nil {
		printer.Printf("EXTRA add ERROR: %v\n", err)
	} else if copyOk {
		return
	}

	options, example, err := setup.OptionsSetup_OneItem_FromEquipped_AllForges(*equipped, prep.model, setup.MissingEnchant_Fix, printer)
	if err != nil {
		printer.Printf("EXTRA add ERROR: %v\n", err)
		return
	}

	added := false
	example.SlotItem().ForEachEquip(func(slot items.SlotEquip) {
		if prep.itemOptions.CouldAddUpgrade_EquipSlot(slot, example, printer, prep.model.SpecificIncompatibleList) == items.CanUpgrade_Yes {
			err := prep.addItemOptionsSpecificWithValidate(slot, options)
			if err != nil {
				printer.Printf("EXTRA add ERROR: %v\n", err)
			} else {
				added = true
			}
		}
	})

	if added {
		printer.Printf("ADDITIONAL EXTRA OPTION from bags %s\n", example.CreateString())
		prep.addedFromBags = append(prep.addedFromBags, equipped.ItemId)
	}
}

func (prep *specItemPrep) extraLoadAndGenerate(itemId items.ItemId, randomSuffix items.RandomSuffix, input *multi_types.ItemInputs, printer *util.PrintRecorder) {
	options, example, err := setup.OptionsSetup_OneItem_FromItemId_AllForges(itemId, input.ExtraUpgradeLevel, randomSuffix, prep.model, printer)
	if err == nil {
		err = prep.addItemOptionsWithValidate(example.SlotItem(), options)
	}
	if err == nil {
		printer.Printf("OPTION %s\n", example.CreateString())
	} else {
		printer.Printf("OPTION add %d ERROR: %v\n", itemId, err)
	}
}

func (prep *specItemPrep) restrictFixed(input *multi_types.ItemInputs, printer *util.PrintRecorder) error {
	printer.Println(prep.label)

	for slot, itemIdList := range input.SemiFixedSlots {
		if !prep.itemOptions.Has(slot) {
			panic("restricting slot but already empty")
		} else if len(itemIdList) == 0 {
			panic("empty restrict list")
		}

		if len(itemIdList) == 1 {
			err := prep.itemOptions.ForceSlotOnlySpecifiedItemId(slot, itemIdList[0])
			if err != nil {
				return err
			}
		} else {
			for _, itemId := range itemIdList {
				if !prep.itemOptions.IncludesItemIdInSlot(itemId, slot) {
					return util.ErrorTracedNew("item included in slot restrictions but not actually available option " + prep.label + " " + itemId.String())
				}
			}

			err := prep.itemOptions.FilterSlot(slot, func(x *items.FullItem) bool { return slices.Contains(itemIdList, x.ItemId()) })
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (prep *specItemPrep) removeBlocked(input *multi_types.ItemInputs, printer *util.PrintRecorder) error {
	// remove blocked items
	for _, itemId := range input.BlockedItems {
		printer.Printf("BLOCKING ITEM %d\n", itemId)
		err := prep.itemOptions.RemoveItemIdFromAll(itemId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (work *specWorker) setupAlternateGems(alternateGemList []stats.GemInfo, printer *util.PrintRecorder) error {
	if len(alternateGemList) > 0 {
		for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
			existing := work.itemOptionsWork.Get(slot)
			for _, item := range existing {
				for _, alternateGem := range alternateGemList {
					alternateItem, err := work.reGemAlternate(item, alternateGem, printer)
					if err != nil {
						return err
					}
					if err := work.addItemOptionsWithValidate_WhereNotExist(slot, []items.FullItem{alternateItem}); err != nil {
						return err
					}
				}

				defaultItem, err := work.reGemDefault(item, printer)
				if err != nil {
					return err
				}
				if err := work.addItemOptionsWithValidate_WhereNotExist(slot, []items.FullItem{defaultItem}); err != nil {
					return err
				}
			}
		}
		work.itemOptionsWork.RemoveDuplicates()
	}
	return nil
}

func (work *specWorker) reGemAlternate(item items.FullItem, alternateGem stats.GemInfo, printer *util.PrintRecorder) (items.FullItem, error) {
	alternateEquipItem := loaders.EquippedItem_FromFull(&item)
	for i := range alternateEquipItem.GemChoice {
		socket := item.SocketSlots()[i]
		if socket.IsStandard() {
			alternateEquipItem.GemChoice[i] = alternateGem.Id
		}
	}
	alternateItem, err := setup.OptionsSetup_OneItem_FromEquipped_OriginalForgeOnly(alternateEquipItem, setup.MissingEnchant_Panic, work.model, util.PrintRecorder_Nop())
	if err != nil {
		return items.FullItem{}, err
	}

	alternateItem.SetNameTag(items.ReGem_GemAlternate)
	return alternateItem, nil
}

func (work *specWorker) reGemDefault(item items.FullItem, printer *util.PrintRecorder) (items.FullItem, error) {
	defaultEquipItem := loaders.EquippedItem_FromFull(&item)
	defaultEquipItem.GemChoice = nil
	defaultItem, err := setup.OptionsSetup_OneItem_FromEquipped_OriginalForgeOnly(defaultEquipItem, setup.MissingEnchant_Fix, work.model, util.PrintRecorder_Nop())
	if err != nil {
		return items.FullItem{}, err
	}

	defaultItem.SetNameTag(items.ReGem_GemDefault)
	return defaultItem, nil
}

func (prep *specItemPrep) makeRandomVariantItems(variantItems []multi_types.RandomVariantItem, input *multi_types.ItemInputs, printer *util.PrintRecorder) error {
	for variantItem := range util_collection.ForPointer(variantItems) {
		for slot := range prep.itemOptions {
			slotOptions := prep.itemOptions[slot]
			newOpt, err := prep.makeRandomVariantItem(variantItem, slotOptions, input, printer)
			if err != nil {
				return err
			}
			slotOptions = newOpt
			prep.itemOptions[slot] = slotOptions
		}
	}
	return nil
}

func (prep *specItemPrep) makeRandomVariantItem(variantItem *multi_types.RandomVariantItem, slotOptions []items.FullItem, input *multi_types.ItemInputs, printer *util.PrintRecorder) ([]items.FullItem, error) {
	hasAny := false
	hasVersion := make([]bool, len(variantItem.RandomSuffixList))
	for item := range util_collection.ForPointer(slotOptions) {
		if item.ItemId() == variantItem.ItemId {
			hasAny = true

			index := slices.Index(variantItem.RandomSuffixList, item.RandomSuffix())
			if index == -1 {
				// panic("variant item found with unexpected suffix")
				continue
			}
			hasVersion[index] = true
		}
	}

	if hasAny {
		for index, randomSuffix := range variantItem.RandomSuffixList {
			if !hasVersion[index] {
				options, example, err := setup.OptionsSetup_OneItem_FromItemId_AllForges(variantItem.ItemId, input.ExtraUpgradeLevel, randomSuffix, prep.model, printer)
				if err != nil {
					return nil, err
				}
				printer.Println("RANDOM_VARIANT adding " + example.CreateString())
				slotOptions = append(slotOptions, options...)
			}
		}

		slotOptions = util_collection.FilterSliceAsNew(slotOptions, func(x *items.FullItem) bool {
			return x.ItemId() != variantItem.ItemId || slices.Contains(variantItem.RandomSuffixList, x.RandomSuffix())
		})
	}

	return slotOptions, nil
}

func (prep *specItemPrep) addItemOptionsWithValidate(slot items.SlotItem, options []items.FullItem) error {
	if err := validateAddOptions(slot.ToSlotEquipOptions()[0], options, prep.model); err != nil {
		return err
	}
	prep.itemOptions.AddSeveralOptions(slot, options)
	return nil
}

func (prep *specItemPrep) addItemOptionsSpecificWithValidate(slot items.SlotEquip, options []items.FullItem) error {
	if err := validateAddOptions(slot, options, prep.model); err != nil {
		return err
	}
	prep.itemOptions.AddSeveralOptionsSpecific(slot, options)
	return nil
}

func (work *specWorker) addItemOptionsWithValidate_WhereNotExist(slot items.SlotEquip, options []items.FullItem) error {
	if err := validateAddOptions(slot, options, work.model); err != nil {
		return err
	}
	work.itemOptionsWork.AddSeveralOptionsSpecific_WhereNotExist(slot, options)
	return nil
}

func validateAddOptions(slot items.SlotEquip, options []items.FullItem, model *gear_model.SpecModel) error {
	if slot == items.Equip_Head {
		for item := range util_collection.ForPointer(options) {
			if err := model.GemChoice.ValidateMetaGemInItem(item); err != nil {
				return err
			}
		}
	}
	return nil
}
