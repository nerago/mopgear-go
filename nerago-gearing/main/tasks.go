package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

func basicReforge(itemOptions *items.FullOptionsMap, model *model.Model, printer *util.PrintRecorder) {
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         itemOptions,
		Model:               model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Long,
		Printer:             nil})
	output.Report(printer)
}

func testSim(printer *util.PrintRecorder) {
	testSimA(printer)
	testSimB(printer)
}
func testSimA(printer *util.PrintRecorder) {
	model := model.Model_PallyProtDps()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &model, setup.MissingEnchant_Panic, printer)
	// itemOptionsMit := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &model, setup.MissingEnchant_Panic, printer)
	// itemOptions[items.Equip_Trinket2] = itemOptionsMit[items.Equip_Trinket2]
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Medium,
		Printer:             printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute(simulate.RunSize_QuickDirty, model.Spec, output.FullSet.Items(), &model, nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}
func testSimB(printer *util.PrintRecorder) {
	model := model.Model_PallyProtMitigation()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &model, setup.MissingEnchant_Panic, printer)
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Medium,
		Printer:             printer})
	printer.Println("Running sim")
	resultStats := simulate.WowSim_Execute(simulate.RunSize_QuickDirty, model.Spec, output.FullSet.Items(), &model, nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}

func slotRating(itemArray []items.FullItem, model *model.Model, printer *util.PrintRecorder) {
	printer.Println("RATINGS")
	// printer.Println(model.StatRatings.(ratings.StatRatingsWeights).Weights())
	printer.Println(model.StatRatings.Weights().CreateString())
	printer.Println0()

	best := util_rank.BestCollector1[items.FullItem]{}
	for _, item := range itemArray {
		rate := model.CalcRatingFullItem(&item)
		printer.Println(item.CreateString())
		printer.Printf("%d\n\n", rate)
		best.Offer(&item, rate)
	}

	printer.Println0()
	printer.Println("BEST")
	printer.Println(best.BestObject.CreateString())
}

// &&&&&&&&&&&&& ffb430ce-e65b-4ae7-97be-4658d37bce7a
// COMMON_COMBO overflow REVISED 94527=false
// COMMON { Foot "Impaling Treads" id=86979 lvl=517 {str=1025 stam=1657 hit=614 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// COMMON { Ring "Band of the Scaled Tyrant (hit->master)" id=95513 lvl=536 {str=909 stam=1484 hit=329 haste=653 master=218} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// COMMON 94527 forbidden
// COMMON { Chest "Talonrender Chestplate (hit->haste)" id=96468 lvl=549 {str=1844 stam=3006 hit=789 crit=1157 haste=526} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Belt "Cloudbreaker Greatbelt (master->crit)" id=96373 lvl=549 {str=1329 stam=2233 crit=371 haste=888 master=557} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// COMMON { Wrist "Frozen Warlord's Bracers (haste->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=479 expert=653 master=319} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// COMMON { Ring "Band of the Shado-Pan Assault (crit->dodge)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 dodge=201} }
// common[96398] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Mastery}
// common[96468] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Haste}
// common[96373] = stats.ReforgeRecipe{From: stats.Stat_Mastery, To: stats.Stat_Crit}
// common[96394] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Mastery}
// common[95140] = stats.ReforgeRecipe{From: stats.Stat_Crit, To: stats.Stat_Dodge}
// common[86979] = stats.ReforgeRecipe_empty
// common[95513] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Mastery}
// ---------------- Ret ----------------
// 88cebb22-8e0c-48aa-ad6a-e782a1a31db1
// SET rating 32759518
// BONUS 1.06
// STATS {str=18191 agi=80 stam=27397 int=80 spi=80 hit=2617 crit=7764 haste=11227 expert=2828 dodge=201 master=7329}
// { Head "White Tiger Helmet (hit->haste)" id=87101 lvl=517 {str=1248 stam=2232 hit=548 haste=364 master=832} ENCHANT {str=180 stam=444 haste=160} GEMS {stam=324}{stam=120 haste=160} }
// { Neck "Amulet of the Primal Turtle (hit->haste)" id=94776 lvl=536 {str=909 stam=1484 hit=277 crit=712 haste=184} ENCHANT {stam=120 crit=60 haste=160} GEMS {stam=120 haste=160} }
// { Shoulder "Lightning Emperor's Pauldrons" id=96658 lvl=535 {str=1147 stam=1960 expert=684 master=855} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (master->expert)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=992 expert=396 master=596} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Talonrender Chestplate (hit->haste)" id=96468 lvl=549 {str=1844 stam=3006 hit=789 crit=1157 haste=526} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (haste->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=479 expert=653 master=319} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "White Tiger Gauntlets (expert->master)" id=87100 lvl=517 {str=1105 stam=1657 crit=700 expert=455 master=303} ENCHANT {str=170 haste=320} GEMS {haste=320} }
// { Belt "Cloudbreaker Greatbelt (master->crit)" id=96373 lvl=549 {str=1329 stam=2233 crit=371 haste=888 master=557} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates" id=96657 lvl=543 {str=1735 stam=2843 crit=1253 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Impaling Treads" id=86979 lvl=517 {str=1025 stam=1657 hit=614 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (crit->dodge)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 dodge=201} }
// { Ring "Band of the Scaled Tyrant (hit->master)" id=95513 lvl=536 {str=909 stam=1484 hit=329 haste=653 master=218} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Primordius' Talisman of Rage (crit->master)" id=94519 lvl=536 {crit=1004 master=668} }
// { Weapon2H "Shin'ka, Execution of Dominion (crit->haste)" id=86386 lvl=504 {str=1318 stam=1977 crit=528 haste=351 master=879} ENCHANT {str=500 haste=320} GEMS {str=500}{haste=320} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76654],"id":87101,"reforging":138,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"upgrade_step":0},{"enchant":4424,"gems":[76667],"id":98147,"reforging":168,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":154,"upgrade_step":2},{"enchant":4432,"gems":[76633],"id":87100,"reforging":161,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"reforging":166,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"upgrade_step":2},{"id":95140,"reforging":142,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":140,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94519,"reforging":147,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54926},{"spellID":54922},{"spellID":146957}],"minor":[{"spellID":57954},{"spellID":57947},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76654],"id":87101,"reforging":138,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"upgrade_step":0},{"enchant":4424,"gems":[76667],"id":98147,"reforging":168,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":154,"upgrade_step":2},{"enchant":4432,"gems":[76633],"id":87100,"reforging":161,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"reforging":166,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"upgrade_step":2},{"id":95140,"reforging":142,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":140,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94519,"reforging":147,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"retribution","talents":"113323","unit":"player","version":"v3.2.1"}
// DPS	207799.41
// TPS	193979.46
// DTPS	9301.11
// HPS	0.00
// TMI	0.00
// DEATH	100.00
// ---------------- Prot-Damage ----------------
// cc067acb-9769-40bb-9605-5a29e8f3f8ec
// SET rating 10178234368
// BONUS 1.00
// STATS {str=19875 agi=80 stam=28261 int=80 spi=80 hit=2674 crit=8513 haste=14435 expert=2658 dodge=201 parry=556 master=3317}
// { Head "Nullification Greathelm" id=87024 lvl=510 {str=1154 stam=2090 crit=882 haste=671} ENCHANT {stam=120 crit=504 haste=160} GEMS {crit=324}{stam=120 haste=160} }
// { Neck "Amulet of the Primal Turtle (hit->haste)" id=94776 lvl=536 {str=909 stam=1484 hit=277 crit=712 haste=184} ENCHANT {stam=120 crit=60 haste=160} GEMS {stam=120 haste=160} }
// { Shoulder "Shoulderguards of Centripetal Destruction" id=94773 lvl=536 {str=1239 stam=1979 crit=962 haste=628} ENCHANT {str=200 crit=100 haste=380} GEMS {haste=320} }
// { Back "Tigerclaw Cape (master->expert)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=992 expert=396 master=596} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Talonrender Chestplate (hit->haste)" id=96468 lvl=549 {str=1844 stam=3006 hit=789 crit=1157 haste=526} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (haste->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=479 expert=653 master=319} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->expert)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 expert=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt (master->crit)" id=96373 lvl=549 {str=1329 stam=2233 crit=371 haste=888 master=557} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Legplates of the Lightning Throne" id=95535 lvl=536 {str=1535 stam=2663 crit=1237 haste=771} ENCHANT {str=465 crit=165 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Foot "Impaling Treads" id=86979 lvl=517 {str=1025 stam=1657 hit=614 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (crit->dodge)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 dodge=201} }
// { Ring "Band of the Scaled Tyrant (hit->master)" id=95513 lvl=536 {str=909 stam=1484 hit=329 haste=653 master=218} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// { Weapon1H "Worldbreaker's Stormscythe (expert->crit)" id=96376 lvl=549 {str=779 stam=1288 hit=605 crit=163 expert=245} ENCHANT {str=60 stam=120 haste=480} GEMS {stam=120 haste=160}{haste=320} }
// { Offhand "Ultimate Protection of the Emperor (parry->haste)" id=96182 lvl=536 {str=909 stam=1484 haste=257 parry=386 master=563} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76588],"id":87024,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":168,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":154,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":146,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"reforging":166,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"upgrade_step":2},{"id":95140,"reforging":142,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":140,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76588,76699],"id":96376,"reforging":159,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":96182,"reforging":131,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76588],"id":87024,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":168,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":154,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":146,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"reforging":166,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"upgrade_step":2},{"id":95140,"reforging":142,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":140,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76588,76699],"id":96376,"reforging":159,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":96182,"reforging":131,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	378457.25
// TPS	2410379.64
// DTPS	46544.34
// HPS	54467.30
// TMI	79.69
// DEATH	7.70
// ---------------- Prot-Mitigation ----------------
// f69fe2fc-660c-41bc-94ae-fba2012fc1c2
// SET rating 9615011840
// BONUS 1.06
// STATS {str=17804 stam=30993 hit=2596 crit=491 haste=9916 expert=3472 dodge=5159 parry=2469 master=11776}
// { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {stam=324}{haste=320} }
// { Neck "Necklace of the Terra-Cotta Vanquisher (hit->haste)" id=95205 lvl=536 {str=909 stam=1484 hit=401 haste=266 master=539} ENCHANT {haste=160 expert=220} GEMS {haste=160 expert=160} }
// { Shoulder "Lightning Emperor's Shoulderguards (hit->dodge)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 dodge=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Oxhorn Bladebreaker (parry->haste)" id=98146 lvl=608 {str=1855 stam=2902 haste=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (haste->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=479 expert=653 master=319} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Lightning Emperor's Handguards (dodge->master)" id=95291 lvl=530 {str=1167 stam=1871 expert=584 dodge=550 master=366} ENCHANT {haste=480 expert=160 dodge=60 master=170} GEMS {haste=160 expert=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt (master->crit)" id=96373 lvl=549 {str=1329 stam=2233 crit=371 haste=888 master=557} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legguards (parry->master)" id=96667 lvl=543 {str=1735 stam=2843 hit=928 parry=810 master=540} ENCHANT {stam=730 haste=480 dodge=165} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (hit->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=422 dodge=691 master=280} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (hit->master)" id=95513 lvl=536 {str=909 stam=1484 hit=329 haste=653 master=218} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// { Weapon1H "Qon's Flaming Scimitar (expert->haste)" id=96534 lvl=549 {str=779 stam=1288 haste=213 expert=320 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"gems":[76667],"id":95205,"reforging":138,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":135,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":131,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":154,"upgrade_step":2},{"enchant":4433,"gems":[76667,76699],"id":95291,"reforging":126,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"reforging":166,"upgrade_step":2},{"enchant":4824,"gems":[76699,76654],"id":96667,"reforging":133,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":140,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":140,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"gems":[76667],"id":95205,"reforging":138,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":135,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":131,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":154,"upgrade_step":2},{"enchant":4433,"gems":[76667,76699],"id":95291,"reforging":126,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"reforging":166,"upgrade_step":2},{"enchant":4824,"gems":[76699,76654],"id":96667,"reforging":133,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":140,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":140,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	307678.58
// TPS	1970379.54
// DTPS	24850.75
// HPS	50522.60
// TMI	38.48
// DEATH	1.25
