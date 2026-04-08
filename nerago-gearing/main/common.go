package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
)

func commonComboCurrent() map[items.ItemId]stats.ReforgeRecipe {
	common := make(map[items.ItemId]stats.ReforgeRecipe)
	// &&&&&&&&&&&&& 2cab5939-b35b-4f9e-a8a3-4f1703784a62
	common[95142] = stats.ReforgeRecipe_empty
	common[96373] = stats.ReforgeRecipe_empty
	common[96793] = stats.ReforgeRecipe_empty
	common[96394] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Mastery)
	common[96398] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[95292] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[96478] = stats.ReforgeRecipe_of(stats.Stat_Parry, stats.Stat_Mastery)
	common[86979] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Crit)
	common[96447] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[95513] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[94945] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[98147] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Dodge)
	common[96668] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Haste)
	common[96533] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Hit)
	common[95140] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Expertise)
	common[96657] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Expertise)
	common[96481] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Mastery)
	common[96534] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	return common
}

// &&&&&&&&&&&&& 2cab5939-b35b-4f9e-a8a3-4f1703784a62
// COMMON_COMBO baseline and fillout REVISED 94527=false 
// COMMON { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Hand "Rein-Binder's Fists (haste->hit)" id=96533 lvl=549 {str=1329 stam=2233 hit=294 crit=1010 haste=442} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// COMMON { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// COMMON { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
// COMMON { Leg "Lightning Emperor's Legplates (crit->expert)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 expert=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// COMMON { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// COMMON { Weapon1H "Qon's Flaming Scimitar (expert->haste)" id=96534 lvl=549 {str=779 stam=1288 haste=213 expert=320 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// COMMON { Neck "Striker's Battletags" id=95142 lvl=530 {str=936 stam=1403 haste=660 master=562} }
// COMMON { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// COMMON { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// COMMON { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {stam=324}{haste=320} }
// COMMON { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// COMMON { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// COMMON { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// COMMON { Ring "Band of the Scaled Tyrant (haste->master)" id=95513 lvl=536 {str=909 stam=1484 hit=547 haste=392 master=261} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// COMMON { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// COMMON { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// COMMON 94527 forbidden
// common[95142] = stats.ReforgeRecipe_empty
// common[96394] = stats.ReforgeRecipe{From: stats.Stat_Expertise, To: stats.Stat_Mastery}
// common[96398] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Mastery}
// common[95292] = stats.ReforgeRecipe{From: stats.Stat_Expertise, To: stats.Stat_Haste}
// common[96478] = stats.ReforgeRecipe{From: stats.Stat_Parry, To: stats.Stat_Mastery}
// common[86979] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Crit}
// common[96447] = stats.ReforgeRecipe{From: stats.Stat_Dodge, To: stats.Stat_Haste}
// common[95513] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Mastery}
// common[94945] = stats.ReforgeRecipe{From: stats.Stat_Dodge, To: stats.Stat_Haste}
// common[98147] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Dodge}
// common[96668] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Haste}
// common[96533] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Hit}
// common[96373] = stats.ReforgeRecipe_empty
// common[95140] = stats.ReforgeRecipe{From: stats.Stat_Crit, To: stats.Stat_Expertise}
// common[96657] = stats.ReforgeRecipe{From: stats.Stat_Crit, To: stats.Stat_Expertise}
// common[96481] = stats.ReforgeRecipe{From: stats.Stat_Dodge, To: stats.Stat_Mastery}
// common[96793] = stats.ReforgeRecipe_empty
// common[96534] = stats.ReforgeRecipe{From: stats.Stat_Expertise, To: stats.Stat_Haste}
// ---------------- Ret ----------------
// 77901dab-e4c7-4d82-80bf-479c49d08e88
// SET rating 32124628
// BONUS 1.02
// STATS {str=18288 agi=80 stam=27426 int=80 spi=80 hit=2585 crit=7599 haste=11748 expert=2578 dodge=858 master=7508}
// { Head "Nullification Greathelm" id=87024 lvl=510 {str=1154 stam=2090 crit=882 haste=671} ENCHANT {stam=120 crit=504 haste=160} GEMS {crit=324}{stam=120 haste=160} }
// { Neck "Striker's Battletags" id=95142 lvl=530 {str=936 stam=1403 haste=660 master=562} }
// { Shoulder "Lightning Emperor's Pauldrons" id=96658 lvl=535 {str=1147 stam=1960 expert=684 master=855} ENCHANT {str=200 crit=220 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Talonrender Chestplate (crit->dodge)" id=96468 lvl=549 {str=1844 stam=3006 hit=1315 crit=695 dodge=462} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (haste->hit)" id=96533 lvl=549 {str=1329 stam=2233 hit=294 crit=1010 haste=442} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->expert)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 expert=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=95513 lvl=536 {str=909 stam=1484 hit=547 haste=392 master=261} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Primordius' Talisman of Rage (crit->master)" id=94519 lvl=536 {crit=1004 master=668} }
// { Weapon2H "Shin'ka, Execution of Dominion (crit->haste)" id=86386 lvl=504 {str=1318 stam=1977 crit=528 haste=351 master=879} ENCHANT {str=500 haste=320} GEMS {str=500}{haste=320} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76588],"id":87024,"upgrade_step":2},{"id":95142,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"upgrade_step":0},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":142,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":151,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":146,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94519,"reforging":147,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54926},{"spellID":54922},{"spellID":146957}],"minor":[{"spellID":57954},{"spellID":57947},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76588],"id":87024,"upgrade_step":2},{"id":95142,"upgrade_step":2},{"enchant":4803,"gems":[76667,76699],"id":96658,"upgrade_step":0},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76699,76654],"id":96468,"reforging":142,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":151,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":146,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94519,"reforging":147,"upgrade_step":2},{"enchant":4444,"gems":[89881,76699],"id":86386,"reforging":145,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"retribution","talents":"113323","unit":"player","version":"v3.2.1"}
// DPS	212218.98
// TPS	198171.34
// DTPS	9299.52
// HPS	0.00
// TMI	0.00
// DEATH	100.00
// ---------------- Prot-Damage ----------------
// 9dced659-da8a-48b6-a456-901d48d4b23b
// SET rating 9692497920
// BONUS 1.00
// STATS {str=19550 agi=80 stam=27774 int=80 spi=80 hit=3034 crit=7471 haste=13097 expert=3047 dodge=396 parry=813 master=4069}
// { Head "Nullification Greathelm" id=87024 lvl=510 {str=1154 stam=2090 crit=882 haste=671} ENCHANT {stam=120 crit=504 haste=160} GEMS {crit=324}{stam=120 haste=160} }
// { Neck "Amulet of the Primal Turtle (hit->haste)" id=94776 lvl=536 {str=909 stam=1484 hit=277 crit=712 haste=184} ENCHANT {stam=120 crit=60 haste=160} GEMS {stam=120 haste=160} }
// { Shoulder "Shoulderguards of Centripetal Destruction" id=94773 lvl=536 {str=1239 stam=1979 crit=962 haste=628} ENCHANT {str=200 crit=100 haste=380} GEMS {haste=320} }
// { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Tyrant King Battleplate (hit->haste)" id=95153 lvl=530 {str=1519 stam=2519 hit=574 haste=382 expert=1089} ENCHANT {str=200 agi=80 stam=200 int=80 spi=80 haste=320 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (haste->hit)" id=96533 lvl=549 {str=1329 stam=2233 hit=294 crit=1010 haste=442} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Legplates of the Lightning Throne (haste->hit)" id=95535 lvl=536 {str=1535 stam=2663 hit=308 crit=1237 haste=463} ENCHANT {str=465 crit=165 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Foot "Impaling Treads (hit->crit)" id=86979 lvl=517 {str=1025 stam=1657 hit=369 crit=245 haste=742} ENCHANT {hit=60 haste=320 master=140} GEMS {haste=320} }
// { Ring "Band of the Shado-Pan Assault (crit->expert)" id=95140 lvl=530 {str=936 stam=1403 crit=302 haste=695 expert=201} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=95513 lvl=536 {str=909 stam=1484 hit=547 haste=392 master=261} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Gaze of the Twins" id=94529 lvl=536 {str=1672} }
// { Weapon1H "Worldbreaker's Stormscythe (expert->haste)" id=96376 lvl=549 {str=779 stam=1288 hit=605 haste=163 expert=245} ENCHANT {str=60 stam=120 haste=480} GEMS {stam=120 haste=160}{haste=320} }
// { Offhand "Ultimate Protection of the Emperor" id=96182 lvl=536 {str=909 stam=1484 parry=643 master=563} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95346,76588],"id":87024,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76667,76588],"id":95153,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":151,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"reforging":151,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76588,76699],"id":96376,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":96182,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95346,76588],"id":87024,"upgrade_step":2},{"gems":[76654],"id":94776,"reforging":138,"upgrade_step":2},{"enchant":4803,"gems":[76699],"id":94773,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4419,"gems":[76667,76588],"id":95153,"reforging":138,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":151,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76667,76633,76633],"id":95535,"reforging":151,"upgrade_step":2},{"enchant":4429,"gems":[76699],"id":86979,"reforging":137,"upgrade_step":2},{"id":95140,"reforging":146,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":94529,"upgrade_step":2},{"enchant":4444,"gems":[76588,76699],"id":96376,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":96182,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	376908.68
// TPS	2398435.14
// DTPS	45581.04
// HPS	53696.35
// TMI	78.09
// DEATH	6.52
// ---------------- Prot-Mitigation-NoSet ----------------
// f2e5671c-6147-4c60-a680-f0756debc005
// SET rating 9258530816
// BONUS 1.00
// STATS {str=18721 stam=30575 hit=2726 crit=3219 haste=10853 expert=3343 dodge=3169 parry=1063 master=11875}
// { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {stam=324}{haste=320} }
// { Neck "Necklace of the Terra-Cotta Vanquisher (master->expert)" id=95205 lvl=536 {str=909 stam=1484 hit=667 expert=215 master=324} ENCHANT {haste=160 expert=220} GEMS {haste=160 expert=160} }
// { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Tigerclaw Cape (haste->dodge)" id=98147 lvl=608 {str=1855 stam=2902 crit=992 haste=596 dodge=396 master=992} ENCHANT {str=60 crit=180 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Rein-Binder's Fists (haste->hit)" id=96533 lvl=549 {str=1329 stam=2233 hit=294 crit=1010 haste=442} ENCHANT {str=290 stam=120 haste=640 expert=160} GEMS {haste=160 expert=160}{stam=120 haste=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legplates (crit->expert)" id=96657 lvl=543 {str=1735 stam=2843 crit=752 expert=501 master=1038} ENCHANT {str=405 stam=120 crit=165 haste=480} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=95513 lvl=536 {str=909 stam=1484 hit=547 haste=392 master=261} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// { Weapon1H "Qon's Flaming Scimitar (expert->haste)" id=96534 lvl=549 {str=779 stam=1288 haste=213 expert=320 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"gems":[76667],"id":95205,"reforging":168,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":151,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":146,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"gems":[76667],"id":95205,"reforging":168,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4424,"gems":[76667],"id":98147,"reforging":149,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4432,"gems":[76667,76588,76699],"id":96533,"reforging":151,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4823,"gems":[76699,76654],"id":96657,"reforging":146,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	1208167.35
// TPS	7465457.23
// DTPS	118593.72
// HPS	213798.63
// TMI	215.32
// DEATH	22.22
// ---------------- Prot-Mitigation-WithSet ----------------
// b8dbf153-2c8f-4c50-88ca-7f0a078f81c6
// SET rating 9693741056
// BONUS 1.06
// STATS {str=17831 stam=30912 hit=2693 crit=120 haste=10551 expert=2991 dodge=4540 parry=2469 master=12151}
// { Head "Lightning Emperor's Faceguard (expert->haste)" id=95292 lvl=530 {str=1439 stam=2519 haste=276 expert=415 master=1196} ENCHANT {stam=324 haste=320 parry=180} GEMS {stam=324}{haste=320} }
// { Neck "Striker's Battletags" id=95142 lvl=530 {str=936 stam=1403 haste=660 master=562} }
// { Shoulder "Lightning Emperor's Shoulderguards (hit->haste)" id=96668 lvl=543 {str=1248 stam=2112 hit=516 haste=343 master=859} ENCHANT {stam=300 haste=480 expert=160 dodge=100 parry=120} GEMS {haste=160 expert=160}{haste=320} }
// { Back "Oxhorn Bladebreaker (parry->haste)" id=98146 lvl=608 {str=1855 stam=2902 haste=396 dodge=992 parry=596 master=992} ENCHANT {stam=290 haste=160 expert=160} GEMS {haste=160 expert=160} }
// { Chest "Rot-Proof Greatplate (dodge->haste)" id=96447 lvl=549 {str=1844 stam=3006 haste=502 dodge=754 master=1256} ENCHANT {stam=420 haste=480 dodge=120} GEMS {haste=320}{stam=120 haste=160} }
// { Wrist "Frozen Warlord's Bracers (expert->master)" id=96394 lvl=549 {str=1116 stam=1675 haste=798 expert=392 master=261} ENCHANT {haste=320 master=170} GEMS {haste=320} }
// { Hand "Lightning Emperor's Handguards (dodge->master)" id=95291 lvl=530 {str=1167 stam=1871 expert=584 dodge=550 master=366} ENCHANT {haste=480 expert=160 dodge=60 master=170} GEMS {haste=160 expert=160}{haste=320} }
// { Belt "Cloudbreaker Greatbelt" id=96373 lvl=549 {str=1329 stam=2233 haste=888 master=928} ENCHANT {crit=120 haste=800 expert=160} GEMS {haste=160 expert=160}{haste=320}{haste=320} }
// { Leg "Lightning Emperor's Legguards (parry->master)" id=96667 lvl=543 {str=1735 stam=2843 hit=928 parry=810 master=540} ENCHANT {stam=730 haste=480 dodge=165} GEMS {haste=320}{stam=120 haste=160} }
// { Foot "Treads of the Blind Eye (parry->master)" id=96478 lvl=549 {str=1409 stam=2233 dodge=887 parry=593 master=395} ENCHANT {stam=120 haste=160 dodge=60 master=140} GEMS {stam=120 haste=160} }
// { Ring "Durumu's Severed Tentacle (dodge->master)" id=96481 lvl=549 {str=1036 stam=1675 hit=702 dodge=415 master=276} ENCHANT {haste=160 expert=160 dodge=60} GEMS {haste=160 expert=160} }
// { Ring "Band of the Scaled Tyrant (haste->master)" id=95513 lvl=536 {str=909 stam=1484 hit=547 haste=392 master=261} ENCHANT {haste=220 expert=160} GEMS {haste=160 expert=160} }
// { Trinket "Spark of Zandalar (haste->master)" id=96398 lvl=549 {haste=1133 master=754} }
// { Trinket "Fortitude of the Zandalari" id=96793 lvl=549 {master=1887} }
// { Weapon1H "Qon's Flaming Scimitar (expert->haste)" id=96534 lvl=549 {str=779 stam=1288 haste=213 expert=320 master=533} ENCHANT {str=60 haste=480 expert=160} GEMS {haste=160 expert=160}{haste=320} }
// { Offhand "Greatshield of the Gloaming (dodge->haste)" id=94945 lvl=536 {str=909 stam=1484 haste=250 dodge=377 master=605} ENCHANT {str=60 haste=160 expert=160 parry=170} GEMS {haste=160 expert=160} }
// {"class":"paladin","gear":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"id":95142,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":131,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4433,"gems":[76667,76699],"id":95291,"reforging":126,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4824,"gems":[76699,76654],"id":96667,"reforging":133,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}],"version":"v3.2.1"},"glyphs":{"major":[{"spellID":54935},{"spellID":63222},{"spellID":54924}],"minor":[{"spellID":57947},{"spellID":57954},{"spellID":57979}]},"id":"Player-4385-05E852E3","level":90,"name":"Neravi","player":{"equipment":{"items":[{"gems":[95344,76633],"id":95292,"reforging":160,"upgrade_step":2},{"id":95142,"upgrade_step":2},{"enchant":4805,"gems":[76667,76699],"id":96668,"reforging":138,"upgrade_step":2},{"enchant":4422,"gems":[76667],"id":98146,"reforging":131,"upgrade_step":2},{"enchant":4420,"gems":[76699,76654],"id":96447,"reforging":124,"upgrade_step":2},{"enchant":4411,"gems":[76633],"id":96394,"reforging":161,"upgrade_step":2},{"enchant":4433,"gems":[76667,76699],"id":95291,"reforging":126,"tinker":4898,"upgrade_step":2},{"gems":[76667,76633,76633],"id":96373,"upgrade_step":2},{"enchant":4824,"gems":[76699,76654],"id":96667,"reforging":133,"upgrade_step":2},{"enchant":4429,"gems":[76654],"id":96478,"reforging":133,"upgrade_step":2},{"gems":[76667],"id":96481,"reforging":126,"upgrade_step":2},{"gems":[76667],"id":95513,"reforging":154,"upgrade_step":2},{"id":96398,"reforging":154,"upgrade_step":2},{"id":96793,"upgrade_step":2},{"enchant":4444,"gems":[76667,76699],"id":96534,"reforging":160,"upgrade_step":2},{"enchant":4993,"gems":[76667],"id":94945,"reforging":124,"upgrade_step":2}]}},"professions":[{"level":600,"name":"Blacksmithing"},{"level":600,"name":"Engineering"}],"race":"BloodElf","realm":"Galakras","spec":"protection","talents":"113213","unit":"player","version":"v3.2.1"}
// DPS	1150190.60
// TPS	7112097.88
// DTPS	110833.04
// HPS	213145.95
// TMI	199.48
// DEATH	15.03