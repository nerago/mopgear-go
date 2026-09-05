package files

const (
	WowSimDB   = `wowsim-external/assets/database/db.json`
	BossLookup = `mopgear-go/files/bosslookup.tsv`

	LogOutputPath = `output/`
	ProfileDir    = `profile/`
	TempData      = `tempdata/`
	TempLog       = `templog/`

	WeightPath           = `../paladin_gearing_weights/paladin-weight/`
	WeightProtMitigation = WeightPath + `PaladinProtMitigation.txt`
	WeightProtSurvival   = WeightPath + `PaladinProtSurvival.txt`
	WeightProtBalanced   = WeightPath + `PaladinProtBalanced.txt`
	WeightProtDamage     = WeightPath + `PaladinProtDamage.txt`
	WeightProtHeal       = WeightPath + `PaladinProtHeal.txt`
	WeightRet            = WeightPath + `PaladinRet.txt`

	GearPath     = `../paladin_gearing_weights/paladin-gear/`
	BagsFilename = GearPath + `bags-gear.json`

	GearFileProtSurvival   = GearPath + `gear-prot-survival.json`
	GearFileProtMitigation = GearPath + `gear-prot-mitigation.json`
	GearFileProtBalanced   = GearPath + `gear-prot-balanced.json`
	GearFileProtDamage     = GearPath + `gear-prot-damage.json`
	GearFileProtHeal       = GearPath + `gear-prot-heal.json`
	GearFileRet            = GearPath + `gear-ret.json`

	SimProtHorridon   = `mopgear-go/files/cli/example-prot-horridon.json`
	SimProtJuggernaut = `mopgear-go/files/cli/example-prot-juggernaut.json`
	SimRet            = `mopgear-go/files/cli/example-ret.json`

	PaladinProtRotation    = `wowsim-external/ui/paladin/protection/apls/iron_juggernaut.apl.json`
	PaladinProtRotationT15 = `mopgear-go/files/rotation/mixT16withWordGlory.apl`
	PaladinRetRotation     = `wowsim-external/ui/paladin/retribution/apls/default.apl.json`
)

func NameToWeight2(base string) string {
	return base + ".v2"
}
func NameToWeight3(base string) string {
	return base + ".v3"
}
