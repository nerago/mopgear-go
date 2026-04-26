package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
)

func commonComboCurrent() map[items.ItemId]stats.ReforgeRecipe {
	common := make(map[items.ItemId]stats.ReforgeRecipe)
	// &&&&&&&&&&&&& 99e43610-d27a-41b3-b728-a571d43a7857
	common[96394] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Mastery)
	common[96534] = stats.ReforgeRecipe_empty
	common[96533] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Mastery)
	common[94945] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[96447] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[94529] = stats.ReforgeRecipe_empty
	common[96657] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
	common[96793] = stats.ReforgeRecipe_empty
	common[95292] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[96668] = stats.ReforgeRecipe_empty
	common[96481] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Mastery)
	common[95282] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Crit)
	common[96478] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Mastery)
	common[96398] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[95140] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[96420] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
	common[96373] = stats.ReforgeRecipe_empty
	common[96500] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[98147] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Hit)
	common[95535] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Expertise)
	common[98146] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Hit)
	common[96658] = stats.ReforgeRecipe_of(stats.Stat_Mastery, stats.Stat_Hit)
	return common
}


// &&&&&&&&&&&&& 99e43610-d27a-41b3-b728-a571d43a7857
// COMMON_COMBO baseline and fillout REVISED 94527=false 
// COMMON { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
// COMMON { Shoulder "Lightning Emperor's Shoulderguards" id=96668 lvl=543 {str=1248 stam=2112 hit=859 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// COMMON { Head "Lightning Emperor's Helmet (hit->crit)" id=95282 lvl=530 {str=1439 stam=2519 hit=634 crit=422 haste=931} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// COMMON { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// COMMON { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// COMMON { Ring "Band of the Shado-Pan Assault (haste->master)" id=95140 lvl=530 {str=936 stam=1403 crit=503 haste=417 master=278} }
// COMMON { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// COMMON { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// COMMON { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Back "Tigerclaw Cape (haste->hit)" id=98147 lvl=608 {str=1855 stam=2902 hit=396 crit=992 haste=596 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Leg "Legplates of the Lightning Throne (crit->expert)" id=95535 lvl=536 {str=1535 stam=2663 crit=743 haste=771 expert=494} ENCHANT {str=465 crit=165 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// COMMON { Back "Oxhorn Bladebreaker (parry->hit)" id=98146 lvl=608 {str=1855 stam=2902 hit=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Shoulder "Lightning Emperor's Pauldrons (master->hit)" id=96658 lvl=543 {str=1248 stam=2112 hit=370 expert=743 master=557} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// COMMON { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// COMMON 94527 forbidden
// COMMON { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// COMMON { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// COMMON { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// common[96394] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Mastery)
// common[96534] = stats.ReforgeRecipe_empty
// common[96533] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Mastery)
// common[94945] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
// common[96447] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
// common[94529] = stats.ReforgeRecipe_empty
// common[96657] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
// common[96793] = stats.ReforgeRecipe_empty
// common[95292] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
// common[96668] = stats.ReforgeRecipe_empty
// common[96481] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Mastery)
// common[95282] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Crit)
// common[96478] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Mastery)
// common[96398] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
// common[95140] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
// common[96420] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
// common[96373] = stats.ReforgeRecipe_empty
// common[96500] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
// common[98147] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Hit)
// common[95535] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Expertise)
// common[98146] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Hit)
// common[96658] = stats.ReforgeRecipe_of(stats.Stat_Mastery, stats.Stat_Hit)
// ---------------- Ret ----------------
// b3b81624-213e-42ac-87b4-43c900e63303
// SET rating 34559588
// BONUS Battlegear of the Lightning Emperor => 4
// STATS {str=20103 agi=80 stam=27435 int=80 spi=80 hit=2693 crit=5452 haste=12667 expert=2812 master=7980}
// { Head "Lightning Emperor's Helmet (hit->crit)" id=95282 lvl=530 {str=1439 stam=2519 hit=634 crit=422 haste=931} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Pauldrons (master->hit)" id=96658 lvl=543 {str=1248 stam=2112 hit=370 expert=743 master=557} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (haste->hit)" id=98147 lvl=608 {str=1855 stam=2902 hit=396 crit=992 haste=596 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Lightning Emperor's Battleplate (haste->master)" id=95910 lvl=510 {str=1154 stam=2091 haste=417 expert=877 master=278} ENCHANT {str=340 agi=80 stam=200 int=80 spi=80 haste=640} GEMS {str=80 haste=160}{haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->haste)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 haste=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Impaling Treads (haste->master)" id=86979 lvl=517 {str=1025 stam=1657 hit=614 haste=446 master=296} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (haste->master)" id=95140 lvl=530 {str=936 stam=1403 crit=503 haste=417 master=278} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// { Weapon2H "Shin'ka, Execution of Dominion (crit->haste)" id=86386 lvl=504 {str=1318 stam=1977 crit=528 haste=351 master=879} ENCHANT {str=500 haste=320} GEMS {str=500}{haste=320} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76699],"id":95282,"reforging":137,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":165,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":151,"upgrade_step":2},{"enchant":4419,"gems":[76669,76699,76588],"id":95910,"reforging":154,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":154,"upgrade_step":2},{"id":95140,"reforging":154,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54926},{"spellID":54922},{"spellID":146957}],"minor":[{"spellID":57954},{"spellID":57947},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76699],"id":95282,"reforging":137,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":165,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":151,"upgrade_step":2},{"enchant":4419,"gems":[76669,76699,76588],"id":95910,"reforging":154,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":154,"upgrade_step":2},{"id":95140,"reforging":154,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"retribution","talents":"113323","unit":"player","version":"v3.2.1"}
// DPS	226250.24
// TPS	212234.81
// DTPS	9127.75
// HPS	0.00
// TMI	0.00
// DEATH	100.00
// ---------------- Prot-Damage ----------------
// 8b474395-e00d-4f01-9bac-eed92d6ee78d
// SET rating 65648920
// BONUS Battlegear of the Lightning Emperor => 1
// STATS {str=20217 agi=80 stam=28537 int=80 spi=80 hit=2715 crit=7988 haste=13517 expert=2796 parry=556 master=5656}
// { Head "Lightning Emperor's Helmet (hit->crit)" id=95282 lvl=530 {str=1439 stam=2519 hit=634 crit=422 haste=931} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// { Neck "Amulet of the Primal Turtle (hit->haste)" id=94776 lvl=536 {str=909 stam=1484 hit=277 crit=712 haste=184} ENCHANT {stam=120 crit=60 haste=160} GEMS {stam=120 haste=160} }
// { Shoulder "Shoulderguards of Centripetal Destruction" id=94773 lvl=536 {str=1239 stam=1979 crit=962 haste=628} ENCHANT {str=200 crit=100 haste=380} GEMS {haste=320} }
// { Back "Tigerclaw Cape (haste->hit)" id=98147 lvl=608 {str=1855 stam=2902 hit=396 crit=992 haste=596 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Talonrender Chestplate (hit->haste)" id=96468 lvl=549 {str=1844 stam=3006 hit=789 crit=1157 haste=526} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (crit->master)" id=96533 lvl=549 {str=1329 stam=2233 crit=606 haste=736 master=404} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Legplates of the Lightning Throne (crit->expert)" id=95535 lvl=536 {str=1535 stam=2663 crit=743 haste=771 expert=494} ENCHANT {str=465 crit=165 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Foot "Jasper Clawfeet (master->haste)" id=87015 lvl=510 {str=955 stam=1553 crit=702 haste=222 master=335} ENCHANT {crit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (haste->master)" id=95140 lvl=530 {str=936 stam=1403 crit=503 haste=417 master=278} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Ultimate Protection of the Emperor (parry->expert)" id=96182 lvl=536 {str=909 stam=1484 expert=257 parry=386 master=563} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76699],"id":95282,"reforging":137,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":151,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"reforging":146,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":87015,"reforging":167,"upgrade_step":2},{"id":95140,"reforging":154,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":96182,"reforging":132,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76699],"id":95282,"reforging":137,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":151,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"reforging":146,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":87015,"reforging":167,"upgrade_step":2},{"id":95140,"reforging":154,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":96182,"reforging":132,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	383737.77
// TPS	2444651.10
// DTPS	42297.39
// HPS	54003.66
// TMI	70.08
// DEATH	3.24
// ---------------- Prot-Compromise ----------------
// ee80cd66-7b81-46a7-a8db-4bb0b1bdfbaf
// SET rating 10063473664
// BONUS Battlegear of the Lightning Emperor => 3
// STATS {str=20927 stam=30333 hit=2721 crit=4421 haste=11948 expert=2948 dodge=2673 parry=763 master=9314}
// { Head "Lightning Emperor's Helmet (hit->crit)" id=95282 lvl=530 {str=1439 stam=2519 hit=634 crit=422 haste=931} ENCHANT {crit=504 haste=320} GEMS {Capacitive}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Pauldrons (master->hit)" id=96658 lvl=543 {str=1248 stam=2112 hit=370 expert=743 master=557} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (haste->hit)" id=98147 lvl=608 {str=1855 stam=2902 hit=396 crit=992 haste=596 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
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
// {"class":"paladin","gear":{"items":[{"gems":[95346,76699],"id":95282,"reforging":137,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":165,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":151,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76699],"id":95282,"reforging":137,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"reforging":165,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":151,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	1053405.04
// TPS	6549551.40
// DTPS	145319.46
// HPS	162434.72
// TMI	290.36
// DEATH	76.71
// ---------------- Prot-Mitigation-NoSet ----------------
// 47e09821-ccee-4153-a69a-be9b52079832
// SET rating 117048192
// BONUS Plate of the Lightning Emperor Prot Damage => 2, Battlegear of the Lightning Emperor => 1
// STATS {str=18995 stam=31247 hit=2576 crit=2103 haste=10697 expert=2620 dodge=3765 parry=1659 master=12699}
// { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Shoulderguards" id=96668 lvl=543 {str=1248 stam=2112 hit=859 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
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
// {"class":"paladin","gear":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":147,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":145,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	973284.54
// TPS	6073553.95
// DTPS	91207.12
// HPS	158433.64
// TMI	185.20
// DEATH	30.58
// ---------------- Prot-Mitigation-WithSet ----------------
// 7e48369c-25d9-4c7e-9075-2ffe55cb8883
// SET rating 182592368
// BONUS Plate of the Lightning Emperor Prot Damage => 4
// STATS {str=18138 stam=31375 hit=3133 crit=580 haste=9533 expert=2971 dodge=4906 parry=3009 master=11798}
// { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {Indomitable}{haste=320} }
// { Neck "Talisman of Angry Spirits (crit->haste)" id=96420 lvl=549 {str=1116 stam=1675 crit=460 haste=306 master=707} }
// { Shoulder "Lightning Emperor's Shoulderguards" id=96668 lvl=543 {str=1248 stam=2112 hit=859 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Oxhorn Bladebreaker (parry->hit)" id=98146 lvl=608 {str=1855 stam=2902 hit=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Lightning Emperor's Handguards (expert->haste)" id=95291 lvl=530 {str=1167 stam=1871 haste=233 expert=351 dodge=916} ENCHANT {haste=480 expert=160 dodge=60 master=170} GEMS {haste=160 expert=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legguards (hit->master)" id=96667 lvl=543 {str=1735 stam=2843 hit=557 parry=1350 master=371} ENCHANT {stam=730 haste=480 dodge=165} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=96500 lvl=549 {str=1036 stam=1675 hit=619 haste=447 master=298} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// { Weapon1H "Qon's Flaming Scimitar" id=96534 lvl=549 {str=779 stam=1288 expert=533 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4433,"gems":[76667,76699],"id":95291,"reforging":160,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4824,"gems":[76699,76654],"id":96667,"reforging":140,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":115738}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"id":96420,"reforging":145,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":129,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4433,"gems":[76667,76699],"id":95291,"reforging":160,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4824,"gems":[76699,76654],"id":96667,"reforging":140,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":96500,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	528250.23
// TPS	3379233.28
// DTPS	43885.55
// HPS	87914.67
// TMI	80.45
// DEATH	7.25