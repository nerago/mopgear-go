package main

//func findUpgrades_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
//	simRunSize := simulate.RunSize_QuickDirty
//	goal := stats.OptimiseGoal_Mitigation
//	model := model_factory.Model_PallyProtSurvival()
//	gearFile := files.GearFileProtSurvival
//	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
//	// upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
//	upgradeItems := []loaders.ItemFoundRef{{ItemId: 96436, UpgradeLevel: 2}} // tortos shell heroic
//	input := upgrades.FindUpgrades_SimInputs{
//		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
//			IncludeNormal:      true,
//			IncludeHeroic:      true,
//			IgnoredItems:       mygear.IgnoredItems,
//			TargetUpgradeLevel: 0,
//			SolverTimeout:      c_upgradeDefaultTimeout,
//		},
//		SimSizeBaseline:    simRunSize,
//		SimSizeItemInitial: simRunSize}
//	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, mygear.SubstituteItemsProt, printer)
//}

//func findUpgrades_T5_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
//	simRunSize := simulate.RunSize_QuickDirty
//	goal := stats.OptimiseGoal_Mitigation
//	model := model_factory.Model_PallyProtSurvival()
//	gearFile := files.GearFileProtSurvival
//	upgradeItems := loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Normal)
//	// upgradeItems := loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic)
//	input := upgrades.FindUpgrades_SimInputs{
//		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
//			IncludeNormal:      true,
//			IncludeHeroic:      true,
//			IgnoredItems:       mygear.IgnoredItems,
//			TargetUpgradeLevel: 0,
//			SolverTimeout:      c_upgradeDefaultTimeout,
//		},
//		SimSizeBaseline:    simRunSize,
//		SimSizeItemInitial: simRunSize}
//	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, mygear.SubstituteItemsProt, printer)
//}

//func findUpgrades_Sim_PaladinDps_Run(printer *util.PrintRecorder) {
//	simRunSize := simulate.RunSize_QuickDirty
//	goal := stats.OptimiseGoal_Dps
//	model := model_factory.Model_PallyProtDamage()
//	gearFile := files.GearFileProtDamage
//	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
//	upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
//	input := upgrades.FindUpgrades_SimInputs{
//		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
//			IncludeNormal:      true,
//			IncludeHeroic:      true,
//			IgnoredItems:       mygear.IgnoredItems,
//			TargetUpgradeLevel: 0,
//			SolverTimeout:      c_upgradeDefaultTimeout,
//		},
//		SimSizeBaseline:    simRunSize,
//		SimSizeItemInitial: simRunSize}
//	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, mygear.SubstituteItemsProt, printer)
//}
