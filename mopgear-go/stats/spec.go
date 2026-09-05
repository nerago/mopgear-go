package stats

type SpecType int8

const (
	Spec_PaladinProt       SpecType = iota
	Spec_PaladinRet        SpecType = iota
	Spec_PaladinHoly       SpecType = iota
	Spec_WarriorProt       SpecType = iota
	Spec_WarriorArms       SpecType = iota
	Spec_DruidBear         SpecType = iota
	Spec_DruidTree         SpecType = iota
	Spec_DruidBoom         SpecType = iota
	Spec_DruidFeral        SpecType = iota
	Spec_MageFrost         SpecType = iota
	Spec_PriestShadow      SpecType = iota
	Spec_PriestHoly        SpecType = iota
	Spec_Rogue             SpecType = iota
	Spec_Warlock           SpecType = iota
	Spec_ShamanRestoration SpecType = iota
	Spec_ShamanElemental   SpecType = iota
	Spec_ShamanEnhance     SpecType = iota
	Spec_Hunter            SpecType = iota
	Spec_MonkBrewmaster    SpecType = iota
	Spec_MonkMistweaver    SpecType = iota
	Spec_MonkDps           SpecType = iota
	Spec_DeathKnightDps    SpecType = iota
	Spec_DeathKnightBlood  SpecType = iota
)

func (spec SpecType) Name() any {
	switch spec {
	case Spec_PaladinProt:
		return "PaladinProt"
	case Spec_PaladinRet:
		return "PaladinRet"
	case Spec_PaladinHoly:
		return "PaladinHoly"
	case Spec_WarriorProt:
		return "WarriorProt"
	case Spec_WarriorArms:
		return "WarriorArms"
	case Spec_DruidBear:
		return "DruidBear"
	case Spec_DruidTree:
		return "DruidTree"
	case Spec_DruidBoom:
		return "DruidBoom"
	case Spec_DruidFeral:
		return "DruidFeral"
	case Spec_MageFrost:
		return "MageFrost"
	case Spec_PriestShadow:
		return "PriestShadow"
	case Spec_PriestHoly:
		return "PriestHoly"
	case Spec_Rogue:
		return "Rogue"
	case Spec_Warlock:
		return "Warlock"
	case Spec_ShamanRestoration:
		return "ShamanRestoration"
	case Spec_ShamanElemental:
		return "ShamanElemental"
	case Spec_ShamanEnhance:
		return "ShamanEnhance"
	case Spec_Hunter:
		return "Hunter"
	case Spec_MonkBrewmaster:
		return "MonkBrewmaster"
	case Spec_MonkMistweaver:
		return "MonkMistweaver"
	case Spec_MonkDps:
		return "MonkDps"
	case Spec_DeathKnightDps:
		return "DeathKnightDps"
	case Spec_DeathKnightBlood:
		return "DeathKnightBlood"
	default:
		return "unknown"
	}
}

func (spec SpecType) ClassName() string {
	switch spec {
	case Spec_PaladinProt:
		return "Paladin"
	case Spec_PaladinRet:
		return "Paladin"
	case Spec_PaladinHoly:
		return "Paladin"
	case Spec_WarriorProt:
		return "Warrior"
	case Spec_WarriorArms:
		return "Warrior"
	case Spec_DruidBear:
		return "Druid"
	case Spec_DruidTree:
		return "Druid"
	case Spec_DruidBoom:
		return "Druid"
	case Spec_DruidFeral:
		return "Druid"
	case Spec_MageFrost:
		return "Mage"
	case Spec_PriestShadow:
		return "Priest"
	case Spec_PriestHoly:
		return "Priest"
	case Spec_Rogue:
		return "Rogue"
	case Spec_Warlock:
		return "Warlock"
	case Spec_ShamanRestoration:
		return "Shaman"
	case Spec_ShamanElemental:
		return "Shaman"
	case Spec_ShamanEnhance:
		return "Shaman"
	case Spec_Hunter:
		return "Hunter"
	case Spec_MonkBrewmaster:
		return "Monk"
	case Spec_MonkMistweaver:
		return "Monk"
	case Spec_MonkDps:
		return "Monk"
	case Spec_DeathKnightDps:
		return "DeathKnight"
	case Spec_DeathKnightBlood:
		return "DeathKnight"
	default:
		return "unknown"
	}
}
