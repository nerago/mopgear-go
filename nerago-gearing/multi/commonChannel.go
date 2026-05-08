package multi

import (
	"maps"
	"paladin_gearing_go/items"
)

const (
	additionalSetEach     uint64 = 32
	additionalThreads     uint64 = 2
	defaultThreadSetCount        = 8
)

type forceItemMode int8

const (
	Force_UnknownTODO              forceItemMode = iota
	Force_ForbiddenTODO            forceItemMode = iota
	Force_OptionalTODO             forceItemMode = iota
	Force_FixedWhereAvailableTODO  forceItemMode = iota
	Force_RequireAtLeastOneUseTODO forceItemMode = iota
	Force_RequiredAlwaysTODO       forceItemMode = iota
)

type commonComboEntry struct {
	Item *items.FullItem
}

type commonCombo struct {
	entryMap map[items.ItemId]commonComboEntry
}

func commonCombo_Make() commonCombo {
	return commonCombo{
		make(map[items.ItemId]commonComboEntry),
	}
}

func (combo *commonCombo) addItem(itemId items.ItemId, item *items.FullItem) {
	combo.entryMap[itemId] = commonComboEntry{Item: item}
}

func (combo *commonCombo) hasItem(itemId items.ItemId) bool {
	_, has := combo.entryMap[itemId]
	return has
}

func (combo *commonCombo) clone() commonCombo {
	return commonCombo{
		maps.Clone(combo.entryMap),
	}
}
