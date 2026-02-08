package model

import (
	. "paladin_gearing_go/types/common"
	// . "paladin_gearing_go/types/stats"
	. "paladin_gearing_go/model/ratings"
)

type Model struct {
	spec             SpecType
	statRatings      StatRatings
	statRequirements StatRequirements
	reforgeRules     ReforgeRules
	enchantChoice    EnchantChoice
	gemChoice        GemChoice
	setBonus         SetBonus
}
