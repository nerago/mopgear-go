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
	WeightRetFile         = `nerago-gearing/files/weight/PaladinRet.txt`

	BagsFilename = `nerago-gearing/files/gear/bags-gear.json`

	GearFileProtMitigationWithSet = `nerago-gearing/files/gear/gear-prot-miti-set.json`
	GearFileProtMitigationNoSet   = `nerago-gearing/files/gear/gear-prot-miti-noset.json`
	GearFileProtCompromise        = `nerago-gearing/files/gear/gear-prot-compromise.json`
	GearFileProtDps               = `nerago-gearing/files/gear/gear-prot-dps.json`
	GearFileRet                   = `nerago-gearing/files/gear/gear-ret.json`

	SimProtMitigation = `nerago-gearing/files/cli/example-prot-miti.json`
	SimProtDps        = `nerago-gearing/files/cli/example-prot-dps.json`
	SimRet            = `nerago-gearing/files/cli/example-ret.json`

	PaladinProtRotation = `wowsim-external/ui/paladin/protection/apls/horridon.apl.json`
	PaladinRetRotation  = `wowsim-external/ui/paladin/retribution/apls/default.apl.json`
)

func SimFileFor(spec stats.SpecType, goal stats.OptimiseGoal) string {
	switch spec {
	case stats.Spec_PaladinProt:
		switch goal {
		case stats.OptimiseGoal_Dps, stats.OptimiseGoal_HalfMitiDps:
			return SimProtDps
		case stats.OptimiseGoal_Mitigation:
			return SimProtMitigation
		default:
			panic("spec+goal not supported")
		}
	case stats.Spec_PaladinRet:
		return SimRet
	default:
		panic("spec not supported")
	}
}
