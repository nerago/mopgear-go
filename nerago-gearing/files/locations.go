package files

import "github.com/nerago/mopgear-go/stats"

const (
	WowSimDB   = `wowsim-external/assets/database/db.json`
	BossLookup = `nerago-gearing/files/bosslookup.tsv`

	LogOutputPath = `output/`
	ProfileDir    = `profile/`
	TempPath      = `tempdata/`

	WeightPath           = `../paladin_gearing_weights/paladin-weight/`
	WeightMitigationFile = WeightPath + `PaladinProtMitigationNoSet.txt`
	WeightSurvivalFile   = WeightPath + `PaladinProtMitigationWithSet.txt`
	WeightBalancedFile   = WeightPath + `PaladinProtCompromise.txt`
	WeightDamageFile     = WeightPath + `PaladinProtDps.txt`
	WeightHealFile       = WeightPath + `PaladinProtHeal.txt`
	WeightRetFile        = WeightPath + `PaladinRet.txt`

	GearPath     = `../paladin_gearing_weights/paladin-gear/`
	BagsFilename = GearPath + `bags-gear.json`

	GearFileProtSurvival   = GearPath + `gear-prot-miti-set.json`
	GearFileProtMitigation = GearPath + `gear-prot-miti-noset.json`
	GearFileProtBalanced   = GearPath + `gear-prot-compromise.json`
	GearFileProtDamage     = GearPath + `gear-prot-dps.json`
	GearFileProtHeal       = GearPath + `gear-prot-heal.json`
	GearFileRet            = GearPath + `gear-ret.json`

	SimProtHorridon   = `nerago-gearing/files/cli/example-prot-horridon.json`
	SimProtJuggernaut = `nerago-gearing/files/cli/example-prot-juggernaut.json`
	SimRet            = `nerago-gearing/files/cli/example-ret.json`

	PaladinProtRotation    = `wowsim-external/ui/paladin/protection/apls/iron_juggernaut.apl.json`
	PaladinProtRotationT15 = `nerago-gearing/files/rotation/mixT16withWordGlory.apl`
	PaladinRetRotation     = `wowsim-external/ui/paladin/retribution/apls/default.apl.json`
)

func SimFileFor(spec stats.SpecType, fight stats.WowSim_Fight) string {
	switch spec {
	case stats.Spec_PaladinProt:
		switch fight {
		case stats.Fight_Horridon_HighHeal, stats.Fight_Horridon_LowHeal, stats.Fight_Animus:
			return SimProtHorridon
		case stats.Fight_Juggernaut_HighHeal, stats.Fight_Juggernaut_NoExternalHeal, stats.Fight_Juggernaut_SelfWordGlory, stats.Fight_Juggernaut_OffHealer:
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

func ToWeight2(base string) string {
	return base + ".v2"
}
func ToWeight3(base string) string {
	return base + ".v3"
}
