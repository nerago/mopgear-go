package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
)

func commonComboCurrent() map[items.ItemId]stats.ReforgeRecipe {
	common := make(map[items.ItemId]stats.ReforgeRecipe)
	// 	&&&&&&&&&&&&& 4e4fe1e5-68b3-4520-910a-b780020da379
	// COMMON_COMBO baseline and fillout REVISED 94527=true
	// COMMON { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
	// COMMON { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
	// COMMON { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
	// COMMON { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
	// COMMON { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
	// COMMON { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
	// COMMON { Trinket "Ji-Kun's Rising Winds (expert->haste)" id=94527 lvl=536 {haste=668 expert=1004} }
	// COMMON { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
	// COMMON { Back "Oxhorn Bladebreaker (parry->hit)" id=98146 lvl=608 {str=1855 stam=2902 hit=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
	// COMMON { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
	// COMMON { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
	// COMMON { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
	// COMMON { Head "Lightning Emperor's Faceguard (expert->hit)" id=96666 lvl=543 {str=1655 stam=2843 hit=320 expert=480 master=1361} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
	// COMMON { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
	// COMMON { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
	// COMMON { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
	// COMMON { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
	// COMMON { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
	// COMMON { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
	// COMMON { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
	// COMMON { Shoulder "Lightning Emperor's Pauldrons (expert->dodge)" id=96658 lvl=543 {str=1248 stam=2112 expert=446 dodge=297 master=927} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
	// COMMON { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
	// COMMON { Head "Lightning Emperor's Helmet (haste->expert)" id=95282 lvl=530 {str=1439 stam=2519 hit=1056 haste=559 expert=372} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
	common[96398] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[96666] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Hit)
	common[98147] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Dodge)
	common[96478] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Mastery)
	common[96500] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[86979] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Crit)
	common[96481] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Mastery)
	common[96668] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Haste)
	common[96534] = stats.ReforgeRecipe_empty
	common[96658] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Dodge)
	common[96394] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Mastery)
	common[95282] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Expertise)
	common[96657] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
	common[95140] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Expertise)
	common[94945] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[94529] = stats.ReforgeRecipe_empty
	common[96793] = stats.ReforgeRecipe_empty
	common[96373] = stats.ReforgeRecipe_empty
	common[94527] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[96420] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
	common[98146] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Hit)
	common[96447] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[96533] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Mastery)
	return common
}

// &&&&&&&&&&&&& 4e4fe1e5-68b3-4520-910a-b780020da379
// COMMON_COMBO baseline and fillout REVISED 94527=true
// COMMON { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
// COMMON { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// COMMON { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// COMMON { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// COMMON { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// COMMON { Trinket "Ji-Kun's Rising Winds (expert->haste)" id=94527 lvl=536 {haste=668 expert=1004} }
// COMMON { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// COMMON { Back "Oxhorn Bladebreaker (parry->hit)" id=98146 lvl=608 {str=1855 stam=2902 hit=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// COMMON { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// COMMON { Head "Lightning Emperor's Faceguard (expert->hit)" id=96666 lvl=543 {str=1655 stam=2843 hit=320 expert=480 master=1361} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
// COMMON { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// COMMON { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// COMMON { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// COMMON { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Shoulder "Lightning Emperor's Pauldrons (expert->dodge)" id=96658 lvl=543 {str=1248 stam=2112 expert=446 dodge=297 master=927} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// COMMON { Head "Lightning Emperor's Helmet (haste->expert)" id=95282 lvl=530 {str=1439 stam=2519 hit=1056 haste=559 expert=372} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// common[96398] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
// common[96666] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Hit)
// common[98147] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Dodge)
// common[96478] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Mastery)
// common[96500] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
// common[86979] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Crit)
// common[96481] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Mastery)
// common[96668] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Haste)
// common[96534] = stats.ReforgeRecipe_empty
// common[96658] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Dodge)
// common[96394] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Mastery)
// common[95282] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Expertise)
// common[96657] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
// common[95140] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Expertise)
// common[94945] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
// common[94529] = stats.ReforgeRecipe_empty
// common[96793] = stats.ReforgeRecipe_empty
// common[96373] = stats.ReforgeRecipe_empty
// common[94527] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
// common[96420] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
// common[98146] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Hit)
// common[96447] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
// common[96533] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Mastery)
// ---------------- Ret ----------------
// e141d6c6-b2fb-4af9-8785-d772377670d9
// SET rating 34022648
// BONUS counts Battlegear of the Lightning Emperor => 4
// BONUS multiply 1.050625
// STATS {str=19896 agi=80 stam=27364 int=80 spi=80 hit=2565 crit=5386 haste=13001 expert=2738 dodge=693 master=7141}
// { Head "Lightning Emperor's Helmet (haste->expert)" id=95282 lvl=530 {str=1439 stam=2519 hit=1056 haste=559 expert=372} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// { Neck "Amulet of the Primal Turtle" id=94776 lvl=536 {str=909 stam=1484 hit=461 crit=712} ENCHANT {stam=120 crit=60 haste=160} GEMS {stam=120 haste=160} }
// { Shoulder "Lightning Emperor's Pauldrons (expert->dodge)" id=96658 lvl=543 {str=1248 stam=2112 expert=446 dodge=297 master=927} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Lightning Emperor's Battleplate (expert->master)" id=95910 lvl=510 {str=1154 stam=2091 haste=695 expert=527 master=350} ENCHANT {str=340 agi=80 stam=200 int=80 spi=80 haste=640} GEMS {str=80 haste=160}{haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// { Weapon2H "Shin'ka, Execution of Dominion (crit->haste)" id=86386 lvl=504 {str=1318 stam=1977 crit=528 haste=351 master=879} ENCHANT {str=500 haste=320} GEMS {str=500}{haste=320} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76699],"id":95282,"reforging":153,"upgrade_step":2},{"gems":[76654],"id":94776,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":156,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76669,76699,76588],"id":95910,"reforging":161,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54926},{"spellID":54922},{"spellID":146957}],"minor":[{"spellID":57954},{"spellID":57947},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76699],"id":95282,"reforging":153,"upgrade_step":2},{"gems":[76654],"id":94776,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":156,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76669,76699,76588],"id":95910,"reforging":161,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"retribution","talents":"113323","unit":"player","version":"v3.2.1"}
// DPS	223248.01
// TPS	209162.89
// DTPS	9125.92
// HPS	0.00
// TMI	0.00
// DEATH	100.00
// ---------------- Prot-Damage ----------------
// 8f220521-fa0b-47f1-9ce9-a6f023f1d1a6
// SET rating 65312172
// BONUS counts Battlegear of the Lightning Emperor => 1
// BONUS multiply 1.000000
// STATS {str=18822 agi=80 stam=28712 int=80 spi=80 hit=2893 crit=7030 haste=14264 expert=3873 dodge=1081 parry=170 master=5792}
// { Head "Lightning Emperor's Helmet (haste->expert)" id=95282 lvl=530 {str=1439 stam=2519 hit=1056 haste=559 expert=372} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Shoulderguards of Centripetal Destruction (haste->expert)" id=94773 lvl=536 {str=1239 stam=1979 crit=962 haste=377 expert=251} ENCHANT {str=200 crit=100 haste=380} GEMS {haste=320} }
// { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Talonrender Chestplate (hit->haste)" id=96468 lvl=549 {str=1844 stam=3006 hit=789 crit=1157 haste=526} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Legplates of the Lightning Throne (haste->dodge)" id=95535 lvl=536 {str=1535 stam=2663 crit=1237 haste=463 dodge=308} ENCHANT {str=465 crit=165 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Ji-Kun's Rising Winds (expert->haste)" id=94527 lvl=536 {haste=668 expert=1004} }
// { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76699],"id":95282,"reforging":153,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"reforging":153,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"reforging":149,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94527,"reforging":160,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76699],"id":95282,"reforging":153,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"reforging":153,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"reforging":149,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94527,"reforging":160,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	362885.48
// TPS	2313179.66
// DTPS	42566.80
// HPS	55987.23
// TMI	70.48
// DEATH	2.78
// ---------------- Prot-Compromise ----------------
// d6cac1e0-fd4a-4fcd-afdc-9837ea3588fc
// SET rating 9953912832
// BONUS counts Battlegear of the Lightning Emperor => 3
// BONUS multiply 1.000000
// STATS {str=21107 stam=30333 hit=2617 crit=3999 haste=10778 expert=2631 dodge=4195 parry=1123 master=9253}
// { Head "Lightning Emperor's Helmet (haste->expert)" id=95282 lvl=530 {str=1439 stam=2519 hit=1056 haste=559 expert=372} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Pauldrons (expert->dodge)" id=96658 lvl=543 {str=1248 stam=2112 expert=446 dodge=297 master=927} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Shell-Coated Wristplates (parry->hit)" id=96428 lvl=549 {str=1116 stam=1675 hit=240 dodge=829 parry=360} ENCHANT {str=180 haste=320} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76699],"id":95282,"reforging":153,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":156,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4415,"gems":[76699],"id":96428,"reforging":129,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76699],"id":95282,"reforging":153,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":156,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4415,"gems":[76699],"id":96428,"reforging":129,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	1073392.55
// TPS	6717006.79
// DTPS	139131.04
// HPS	157643.13
// TMI	284.72
// DEATH	74.74
// ---------------- Prot-Mitigation-NoSet ----------------
// cb3e807f-a9d9-4c13-a07a-ee40fabfec8d
// SET rating 118230296
// BONUS counts Plate of the Lightning Emperor => 2, Battlegear of the Lightning Emperor => 1
// BONUS multiply 1.000000
// STATS {str=19211 stam=31571 hit=2553 crit=2103 haste=10764 expert=2685 dodge=3765 parry=1659 master=12864}
// { Head "Lightning Emperor's Faceguard (expert->hit)" id=96666 lvl=543 {str=1655 stam=2843 hit=320 expert=480 master=1361} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Oxhorn Bladebreaker (parry->hit)" id=98146 lvl=608 {str=1855 stam=2902 hit=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76699],"id":96666,"reforging":158,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76699],"id":96666,"reforging":158,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	980478.46
// TPS	6119968.72
// DTPS	88870.87
// HPS	159859.80
// TMI	178.80
// DEATH	27.94
// ---------------- Prot-Mitigation-WithSet ----------------
// 45258e51-82ab-4a25-9e73-6049004c2acd
// SET rating 182059328
// BONUS counts Plate of the Lightning Emperor => 2, Battlegear of the Lightning Emperor => 1
// BONUS multiply 1.013000
// STATS {str=19211 stam=31571 hit=2553 crit=2103 haste=10764 expert=2685 dodge=3765 parry=1659 master=12864}
// { Head "Lightning Emperor's Faceguard (expert->hit)" id=96666 lvl=543 {str=1655 stam=2843 hit=320 expert=480 master=1361} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Oxhorn Bladebreaker (parry->hit)" id=98146 lvl=608 {str=1855 stam=2902 hit=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76699],"id":96666,"reforging":158,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76699],"id":96666,"reforging":158,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	541487.67
// TPS	3468105.76
// DTPS	43054.90
// HPS	87824.39
// TMI	79.73
// DEATH	9.27
