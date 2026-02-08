package ratings

import (
	. "paladin_gearing_go/db"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/types/stats"
)

type StatRequirements interface {
	setMatches(itemSet SolvableItemSet) bool
	toSkinny(item SolvableItem) SkinnyItem
	skinnyMatch(skinny SkinnyItem, item SolvableItem) bool
}

type StatRatings interface {
	calcRating(block StatBlock) uint64
}
