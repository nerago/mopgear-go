package files

import "paladin_gearing_go/stats"

const (
	WowSimDB = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\wowsimdb.json`
	BossLookup = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\bosslookup.tsv`

	LogOutputPath = `C:\Users\nicholas\Dropbox\prog\paladin_gearing_go\output\`
	ProfileDir    = `C:\Users\nicholas\Dropbox\prog\paladin_gearing_go\`

	WeightMitiFile = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\weight\PaladinProtMitigation.txt`
	WeightDpsFile  = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\weight\PaladinProtDps.txt`
	WeightRetFile  = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\weight\PaladinRet.txt`

	BagsFilename = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\bags-gear-bags.json`

	GearFileProtMitigation = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
	GearFileProtDps        = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-dps.json`
	GearFileRet            = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-ret.json`

	SimProtMitigation = "C:\\Users\\nicholas\\Dropbox\\prog\\wow-sim-mop\\example-prot-miti.json"
	SimProtDps        = "C:\\Users\\nicholas\\Dropbox\\prog\\wow-sim-mop\\example-prot-dps.json"
	SimRet            = "C:\\Users\\nicholas\\Dropbox\\prog\\wow-sim-mop\\example-ret.json"
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
