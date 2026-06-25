package files

import "paladin_gearing_go/stats"

const (
	WowSimDB   = `wowsim-external/assets/database/db.json`
	BossLookup = `nerago-gearing/files/bosslookup.tsv`

	LogOutputPath = `output/`
	ProfileDir    = `profile/`

	WeightMitiNoSetFile   = `nerago-gearing/files/weight/PaladinProtMitigationNoSet.txt`
	WeightMitiWithSetFile = `nerago-gearing/files/weight/PaladinProtMitigationWithSet.txt`
	WeightCompromiseFile  = `nerago-gearing/files/weight/PaladinProtCompromise.txt`
	WeightDpsFile         = `nerago-gearing/files/weight/PaladinProtDps.txt`
	WeightHealFile        = `nerago-gearing/files/weight/PaladinProtHeal.txt`
	WeightRetFile         = `nerago-gearing/files/weight/PaladinRet.txt`

	BagsFilename = `nerago-gearing/files/gear/bags-gear.json`

	GearFileProtMitigationWithSet = `nerago-gearing/files/gear/gear-prot-miti-set.json`
	GearFileProtMitigationNoSet   = `nerago-gearing/files/gear/gear-prot-miti-noset.json`
	GearFileProtCompromise        = `nerago-gearing/files/gear/gear-prot-compromise.json`
	GearFileProtDps               = `nerago-gearing/files/gear/gear-prot-dps.json`
	GearFileProtHeal              = `nerago-gearing/files/gear/gear-prot-heal.json`
	GearFileRet                   = `nerago-gearing/files/gear/gear-ret.json`

	SimProtHorridon   = `nerago-gearing/files/cli/example-prot-horridon.json`
	SimProtJuggernaut = `nerago-gearing/files/cli/example-prot-juggernaut.json`
	SimRet            = `nerago-gearing/files/cli/example-ret.json`

	PaladinProtRotation = `wowsim-external/ui/paladin/protection/apls/horridon.apl.json`
	PaladinRetRotation  = `wowsim-external/ui/paladin/retribution/apls/default.apl.json`
)

func SimFileFor(spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight) string {
	switch spec {
	case stats.Spec_PaladinProt:
		switch fight {
		case stats.Fight_Horridon_HighHeal, stats.Fight_Horridon_LowHeal, stats.Fight_Animus:
			return SimProtHorridon
		case stats.Fight_Juggernaut_HighHeal, stats.Fight_Juggernaut_LowHeal:
			return SimProtJuggernaut
		default:
			panic("unknown spec/fight")
		}
	case stats.Spec_PaladinRet:
		return SimRet
	default:
		panic("spec not supported")
	}
}
