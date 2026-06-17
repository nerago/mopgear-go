package model

import (
	"maps"
	. "paladin_gearing_go/db"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/stats"
)

type EnchantChoice map[SlotItem]GemInfo

func (enchant EnchantChoice) Equals(other EnchantChoice) bool {
	return maps.Equal(enchant, other)
}

func (enchant EnchantChoice) GetChoice(slot SlotItem) *GemInfo {
	info, ok := enchant[slot]
	if ok {
		return &info
	} else {
		return nil
	}
}

func EnchantChoice_ForSpec(spec SpecType, goal OptimiseGoal) map[SlotItem]GemInfo {
	result := make(map[SlotItem]GemInfo)
	switch spec {

	case Spec_PaladinProt:
		switch goal {
		case OptimiseGoal_Mitigation, OptimiseGoal_HalfMitiHeal:
			result[Item_Shoulder] = EnchantData_ById(4805)
			result[Item_Back] = EnchantData_ById(4422)
			result[Item_Chest] = EnchantData_ById(4420)
			result[Item_Wrist] = EnchantData_ById(4411)
			result[Item_Hand] = EnchantData_ById(4433)
			result[Item_Leg] = EnchantData_ById(4824)
			result[Item_Foot] = EnchantData_ById(4429)
			result[Item_Weapon1H] = EnchantData_ById(4444)
			result[Item_Offhand] = EnchantData_ById(4993)

		case OptimiseGoal_Dps, OptimiseGoal_HalfMitiDps:
			result[Item_Shoulder] = EnchantData_ById(4803)
			result[Item_Back] = EnchantData_ById(4424)
			result[Item_Chest] = EnchantData_ById(4419)
			result[Item_Wrist] = EnchantData_ById(4415)
			result[Item_Hand] = EnchantData_ById(4432)
			result[Item_Leg] = EnchantData_ById(4823)
			result[Item_Foot] = EnchantData_ById(4429)
			result[Item_Weapon1H] = EnchantData_ById(4444)
			result[Item_Offhand] = EnchantData_ById(4993)

		default:
			panic("EnchantChoice not known")
		}

	case Spec_PaladinRet:
		result[Item_Shoulder] = EnchantData_ById(4803)
		result[Item_Back] = EnchantData_ById(4424)
		result[Item_Chest] = EnchantData_ById(4419)
		result[Item_Wrist] = EnchantData_ById(4415)
		result[Item_Hand] = EnchantData_ById(4432)
		result[Item_Leg] = EnchantData_ById(4823)
		result[Item_Foot] = EnchantData_ById(4429)
		result[Item_Weapon2H] = EnchantData_ById(4444)

	default:
		panic("EnchantChoice not known")
	}
	return result
}
