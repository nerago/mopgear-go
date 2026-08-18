package multi_types

import (
	"iter"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type CommonOptions struct {
	byRef   util_collection.MapSlice[items.ItemRef, items.FullItem]
	idToRef util_collection.MapMap[items.ItemId, items.ItemRef, bool]
}

func CommonOptions_Make() CommonOptions {
	common := CommonOptions{}
	common.byRef.Init(32)
	common.idToRef.Init(32)
	return common
}

func (co *CommonOptions) Get(ref items.ItemRef) ([]items.FullItem, bool) {
	slice, found := co.byRef.GetInternalSliceOptional(ref)
	return slice, found
}

func (co *CommonOptions) IncludesItemId(itemId items.ItemId) bool {
	return co.idToRef.HasKey1(itemId)
}

func (co *CommonOptions) Size() int {
	count := 0
	for range co.byRef.SeqValuesPointer() {
		count++
	}
	return count
}

func (co *CommonOptions) SeqGroups() iter.Seq2[items.ItemRef, []items.FullItem] {
	return co.byRef.SeqGroupsInternalSlice()
}

func (co *CommonOptions) ApplyToSlicesByItemRef(itemRef items.ItemRef, apply func(prior []items.FullItem) []items.FullItem) {
	co.byRef.MapInternalSlice(itemRef, func(slice []items.FullItem) []items.FullItem {
		newSlice := apply(slice)
		co.afterModifySlice(itemRef, newSlice)
		return newSlice
	})
}

func (co *CommonOptions) ApplyToSlicesByItemId(itemId items.ItemId, apply func(options []items.FullItem) []items.FullItem) {
	for itemRef := range co.idToRef.SeqKey2ValueWithKey1(itemId) {
		co.ApplyToSlicesByItemRef(itemRef, apply)
	}
}

func (co *CommonOptions) ApplyToAllSlices(apply func([]items.FullItem) []items.FullItem) {
	co.byRef.MapInternalSlicesAll(func(itemRef items.ItemRef, slice []items.FullItem) []items.FullItem {
		newSlice := apply(slice)
		co.afterModifySlice(itemRef, newSlice)
		return newSlice
	})
}

func (co *CommonOptions) afterModifySlice(itemRef items.ItemRef, newSliceForItemRef []items.FullItem) {
	if len(newSliceForItemRef) > 0 {
		co.idToRef.Put(itemRef.ItemId, itemRef, true)
	} else {
		co.idToRef.Delete(itemRef.ItemId, itemRef)
	}
}

func (co *CommonOptions) RemoveByItemRef(itemRef items.ItemRef) {
	co.byRef.RemoveAllForKey(itemRef)
}
