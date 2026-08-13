package multi

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"slices"
)

type specItemPrep struct {
	label             string
	model             gear_model.SpecModel
	exactEquippedGear items.FullEquipMap
	itemOptions       items.FullOptionsMap
	addedFromBags     []items.ItemId
	inputs            *multi_types.ItemInputs
	seenInSolutions   *seenMap
}

func (job *MultiSetJob) prepareItems() {
	job.printer.Println("LOADING BAGS")
	job.bagsGear = loaders.BagsFileReader_Read()

	job.printer.Println("PREPARING STARTING GEAR")
	params := job.input.Param
	itemPrepSlice := util_collection.MapSliceAsNew(params, func(param *multi_types.SpecParam) specItemPrep {
		prep := specItemPrep{
			label:           param.Label,
			model:           param.Model,
			inputs:          &param.ItemInputs,
			seenInSolutions: &seenMap{content: make(map[items.ItemId]uint32)},
		}
		prep.prepareStartingGear(&param.ItemInputs, &param.Model, job.printer)
		return prep
	})

	job.itemPrep = util_collection.SliceToMap(itemPrepSlice,
		func(p *specItemPrep) string { return p.label },
		util_collection.IdentityFunc[*specItemPrep])

	refs := &extraRefs{
		bagsGear:   &job.bagsGear,
		shared:     &job.input.ItemInput,
		otherPreps: job.itemPrep,
	}

	job.printer.Println("PREPARING EXTRA ITEMS")
	for i := range itemPrepSlice {
		itemPrepSlice[i].prepareExtraItems(&params[i].ItemInputs, refs, job.printer)
	}

	job.printer.Println("ADDITIONAL ITEM SETUP")
	for i := range itemPrepSlice {
		prep := &itemPrepSlice[i]
		input := &params[i].ItemInputs
		prep.removeBlocked(input, job.printer)
		prep.restrictFixed(input, job.printer)
		prep.makeRandomVariantItems(job.input.ItemInput.RandomVariantItems, input, job.printer)
		prep.setupAlternateGems(job.input.ItemInput.AlternateGemming, job.printer)
	}
}

func (job *MultiSetJob) paramOrderSlice() []string {
	return util_collection.MapSliceAsNew(job.input.Param, func(param *multi_types.SpecParam) string {
		return param.Label
	})
}

type extraRefs struct {
	bagsGear   *loaders.EquippedArray
	shared     *multi_types.ItemInputShared
	otherPreps map[string]*specItemPrep
}

func (prep *specItemPrep) prepareStartingGear(input *multi_types.ItemInputs, model *gear_model.SpecModel, printer *util.PrintRecorder) {
	printer.Println(prep.label)

	equipped := loaders.GearFileReader_Read(input.GearFile)
	prep.exactEquippedGear = setup.OptionsSetup_ExactEquippedOnly(equipped, model, input.MissingEnchant, printer)
	prep.itemOptions = setup.OptionsSetup_FromEquipped(equipped, model, input.MissingEnchant, printer)

	setup.UpgradeAllOptionsToLevel2(&prep.itemOptions, input.ForceUpgradeExistingItems, model, printer)
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

	if prep.copyExtraFromOtherSpec(itemId, refs, printer) {
		return
	}

	if prep.copyExtraFromBags(itemId, refs, printer) {
		return
	}

	prep.extraLoadAndGenerate(itemId, items.NO_RANDOM_SUFFIX, input, printer)
}

func (prep *specItemPrep) copyExtraFromOtherSpec(itemId items.ItemId, refs *extraRefs, printer *util.PrintRecorder) bool {
	usageGroups, hasUsageGroups := refs.shared.DistinctUsageGroups[itemId]

	options := make([]items.FullItem, 0)
	if !hasUsageGroups {
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
		} else if slices.Contains(usageGroups.GroupALabels, prep.label) {
			for _, otherLabel := range usageGroups.GroupBLabels {
				otherPrep := refs.otherPreps[otherLabel]
				more := otherPrep.itemOptions.FindItemId(itemId)
				options = slices.AppendSeq(options, more)
			}
		} else {
			panic("expected param label to be in one of the groups")
		}
	}

	util_collection.RemoveDuplicatesFunc_InPlace(&options, (*items.FullItem).Equals)

	// NOTE these may not copy with the model's reforge preferences etc

	if len(options) > 0 {
		prep.addItemOptionsWithValidate(options[0].SlotItem(), options)
		printer.Printf("OPTION from other spec %s\n", options[0].CreateString())
		return true
	} else {
		return false
	}
}

func (prep *specItemPrep) copyExtraFromBags(itemId items.ItemId, refs *extraRefs, printer *util.PrintRecorder) bool {
	equipped := refs.bagsGear.GetWithItemId(itemId)
	if equipped != nil {
		options, example := setup.OptionsSetup_Single_FromEquipped(*equipped, &prep.model, setup.MissingEnchant_Fix, printer)
		prep.addItemOptionsWithValidate(example.SlotItem(), options)
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

	if prep.copyExtraFromOtherSpec(equipped.ItemId, refs, printer) {
		return
	}

	options, example := setup.OptionsSetup_Single_FromEquipped(*equipped, &prep.model, setup.MissingEnchant_Fix, printer)

	added := false
	example.SlotItem().ForEachEquip(func(slot items.SlotEquip) {
		if prep.itemOptions.CouldAddUpgrade_EquipSlot(slot, example, printer, prep.model.SpecificIncompatibleList) == items.CanUpgrade_Yes {
			printer.Printf("ADDITIONAL EXTRA OPTION from bags %s\n", example.CreateString())
			prep.addItemOptionsSpecificWithValidate(slot, options)
			added = true
		}
	})

	if added {
		prep.addedFromBags = append(prep.addedFromBags, equipped.ItemId)
	}
}

func (prep *specItemPrep) extraLoadAndGenerate(itemId items.ItemId, randomSuffix items.RandomSuffix, input *multi_types.ItemInputs, printer *util.PrintRecorder) {
	options, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, input.ExtraUpgradeLevel, randomSuffix, &prep.model, printer)
	prep.addItemOptionsWithValidate(example.SlotItem(), options)
	printer.Printf("OPTION %s\n", example.CreateString())
}

func (prep *specItemPrep) restrictFixed(input *multi_types.ItemInputs, printer *util.PrintRecorder) {
	printer.Println(prep.label)

	for slot, itemIdList := range input.SemiFixedSlots {
		if !prep.itemOptions.Has(slot) {
			panic("restricting slot but already empty")
		} else if len(itemIdList) == 0 {
			panic("empty restrict list")
		}

		if len(itemIdList) == 1 {
			prep.itemOptions.ForceSlotOnlySpecifiedItemId(slot, itemIdList[0])
		} else {
			for _, itemId := range itemIdList {
				if !prep.itemOptions.IncludesItemIdInSlot(itemId, slot) {
					panic("item included in slot restrictions but not actually available option " + prep.label + " " + itemId.String())
				}
			}

			prep.itemOptions.FilterSlot(slot, func(x *items.FullItem) bool { return slices.Contains(itemIdList, x.ItemId()) })
		}
	}
}

func (prep *specItemPrep) removeBlocked(input *multi_types.ItemInputs, printer *util.PrintRecorder) {
	// remove blocked items
	for _, itemId := range input.BlockedItems {
		printer.Printf("BLOCKING ITEM %d\n", itemId)
		prep.itemOptions.RemoveItemIdFromAll(itemId)
	}
}

func (prep *specItemPrep) setupAlternateGems(alternateGemList []stats.GemInfo, printer *util.PrintRecorder) {
	if len(alternateGemList) > 0 {
		for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
			existing := prep.itemOptions.Get(slot)
			for _, item := range existing {
				for _, alternateGem := range alternateGemList {
					alternateItem := prep.reGemAlternate(item, alternateGem, printer)
					prep.addItemOptionsWithValidate_WhereNotExist(slot, []items.FullItem{alternateItem})
				}

				defaultItem := prep.reGemDefault(item, printer)
				prep.addItemOptionsWithValidate_WhereNotExist(slot, []items.FullItem{defaultItem})
			}
		}
		prep.itemOptions.RemoveDuplicates()
	}
}

func (prep *specItemPrep) reGemAlternate(item items.FullItem, alternateGem stats.GemInfo, printer *util.PrintRecorder) items.FullItem {
	alternateEquipItem := loaders.EquippedItem_FromFull(&item)
	for i := range alternateEquipItem.GemChoice {
		socket := item.SocketSlots()[i]
		if socket.IsStandard() {
			alternateEquipItem.GemChoice[i] = alternateGem.Id
		}
	}
	alternateItem := setup.OptionsSetup_ExactEquippedOnly_Item(alternateEquipItem, setup.MissingEnchant_Panic, &prep.model, printer)

	alternateItem.SetNameTag(items.ReGem_GemAlternate)
	return alternateItem
}

func (prep *specItemPrep) reGemDefault(item items.FullItem, printer *util.PrintRecorder) items.FullItem {
	defaultEquipItem := loaders.EquippedItem_FromFull(&item)
	defaultEquipItem.GemChoice = nil
	defaultItem := setup.OptionsSetup_ExactEquippedOnly_Item(defaultEquipItem, setup.MissingEnchant_Fix, &prep.model, printer)
	defaultItem.SetNameTag(items.ReGem_GemDefault)
	return defaultItem
}

func (prep *specItemPrep) makeRandomVariantItems(variantItems []multi_types.RandomVariantItem, input *multi_types.ItemInputs, printer *util.PrintRecorder) {
	for variantItem := range util_collection.ForPointer(variantItems) {
		for slot := range prep.itemOptions {
			slotOptions := prep.itemOptions[slot]
			slotOptions = prep.makeRandomVariantItem(variantItem, slotOptions, input, printer)
			prep.itemOptions[slot] = slotOptions
		}
	}
}

func (prep *specItemPrep) makeRandomVariantItem(variantItem *multi_types.RandomVariantItem, slotOptions []items.FullItem, input *multi_types.ItemInputs, printer *util.PrintRecorder) []items.FullItem {
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
				options, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(variantItem.ItemId, input.ExtraUpgradeLevel, randomSuffix, &prep.model, printer)
				printer.Println("RANDOM_VARIANT adding " + example.CreateString())
				slotOptions = append(slotOptions, options...)
			}
		}

		slotOptions = util_collection.FilterSliceAsNew(slotOptions, func(x *items.FullItem) bool {
			return x.ItemId() != variantItem.ItemId || slices.Contains(variantItem.RandomSuffixList, x.RandomSuffix())
		})
	}

	return slotOptions
}

func (prep *specItemPrep) addItemOptionsWithValidate(slot items.SlotItem, options []items.FullItem) {
	prep.validateAddOptions(slot.ToSlotEquipOptions()[0], options)
	prep.itemOptions.AddSeveralOptions(slot, options)
}

func (prep *specItemPrep) addItemOptionsSpecificWithValidate(slot items.SlotEquip, options []items.FullItem) {
	prep.validateAddOptions(slot, options)
	prep.itemOptions.AddSeveralOptionsSpecific(slot, options)
}

func (prep *specItemPrep) addItemOptionsWithValidate_WhereNotExist(slot items.SlotEquip, options []items.FullItem) {
	prep.validateAddOptions(slot, options)
	prep.itemOptions.AddSeveralOptionsSpecific_WhereNotExist(slot, options)
}

func (prep *specItemPrep) validateAddOptions(slot items.SlotEquip, options []items.FullItem) {
	if slot == items.Equip_Head {
		for item := range util_collection.ForPointer(options) {
			prep.model.GemChoice.ValidateMetaGemInItem(item)
		}
	}
}
