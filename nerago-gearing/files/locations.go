package files

import "paladin_gearing_go/stats"

const (
	WowSimDB   = `wowsim-external\assets\database\db.json`
	BossLookup = `nerago-gearing\files\bosslookup.tsv`

	LogOutputPath = `output\`
	ProfileDir    = `profile\`

	WeightMitiFile = `nerago-gearing\files\weight\PaladinProtMitigation.txt`
	WeightDpsFile  = `nerago-gearing\files\weight\PaladinProtDps.txt`
	WeightRetFile  = `nerago-gearing\files\weight\PaladinRet.txt`

	BagsFilename = `nerago-gearing\files\gear\bags-gear.json`

	GearFileProtMitigation = `nerago-gearing\files\gear\gear-prot-defence.json`
	GearFileProtDps        = `nerago-gearing\files\gear\gear-prot-dps.json`
	GearFileRet            = `nerago-gearing\files\gear\gear-ret.json`

	SimProtMitigation = `nerago-gearing\files\cli\example-prot-miti.json`
	SimProtDps        = `nerago-gearing\files\cli\example-prot-dps.json`
	SimRet            = `nerago-gearing\files\cli\example-ret.json`

	PaladinProtRotation = `wowsim-external\ui\paladin\protection\apls\horridon.apl.json`
)

func SimFileFor(spec stats.SpecType) string {
	switch spec {
	case stats.Spec_PaladinProtDps:
		return SimProtDps
	case stats.Spec_PaladinProtMitigation:
		return SimProtMitigation
	case stats.Spec_PaladinRet:
		return SimRet
	default:
		panic("spec not supported")
	}
}

func GearFileFor(spec stats.SpecType) string {
	switch spec {
	case stats.Spec_PaladinProtDps:
		return GearFileProtDps
	case stats.Spec_PaladinProtMitigation:
		return GearFileProtMitigation
	case stats.Spec_PaladinRet:
		return GearFileRet
	default:
		panic("spec not supported")
	}
}
