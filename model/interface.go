package model

import (
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/types/stats"
)

type StatRequirements interface {
	CheckSet(block *StatBlock) bool
	CheckSetSkinny(set *SkinnyItemSet) bool
	ToSkinny(item *SolvableItem) SkinnyItem
	SkinnyMatch(skinny *SkinnyItem, item *SolvableItem) bool
}

type StatRatings interface {
	CalcRating(block *StatBlock) uint64
}
