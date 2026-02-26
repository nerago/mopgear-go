package stats

type SpecType int32

const (
	Spec_PaladinProtMitigation SpecType = iota
	Spec_PaladinProtDps        SpecType = iota
	Spec_PaladinRet            SpecType = iota
	Spec_PaladinHoly           SpecType = iota
	Spec_WarriorProt           SpecType = iota
	Spec_WarriorArms           SpecType = iota
	Spec_DruidBear             SpecType = iota
	Spec_DruidTree             SpecType = iota
	Spec_DruidBoom             SpecType = iota
	Spec_DruidFeral            SpecType = iota
	Spec_MageFrost             SpecType = iota
	Spec_PriestShadow          SpecType = iota
	Spec_PriestHoly            SpecType = iota
	Spec_Rogue                 SpecType = iota
	Spec_Warlock               SpecType = iota
	Spec_ShamanRestoration     SpecType = iota
	Spec_ShamanElemental       SpecType = iota
	Spec_ShamanEnhance         SpecType = iota
	Spec_Hunter                SpecType = iota
	Spec_MonkBrewmaster        SpecType = iota
	Spec_MonkMistweaver        SpecType = iota
	Spec_MonkDps               SpecType = iota
	Spec_DeathKnightDps        SpecType = iota
	Spec_DeathKnightBlood      SpecType = iota
)
