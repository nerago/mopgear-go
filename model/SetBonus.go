package model

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/stats"
)

type SetBonus struct {
	activeSets []setInfo
	itemToSet  map[uint32]int
}

func SetBonus_Named(names ...string) SetBonus {
	sets := SetBonus{}
	for _, name := range names {
		for _, info := range g_setData {
			if info.name == name {
				sets.activeSets = append(sets.activeSets, info)
			}
		}
	}
	sets.initMap()
	return sets
}

func SetBonus_ForSpec(spec SpecType) SetBonus {
	sets := SetBonus{}
	for _, info := range g_setData {
		if info.spec == spec {
			sets.activeSets = append(sets.activeSets, info)
		}
	}
	sets.initMap()
	return sets
}

func SetBonus_Empty() SetBonus {
	return SetBonus{}
}

func (sets *SetBonus) initMap() {
	sets.itemToSet = make(map[uint32]int)
	for index, info := range sets.activeSets {
		for _, itemId := range info.items {
			sets.itemToSet[itemId] = index
		}
	}
}

func (sets *SetBonus) CalcAndMultiply(itemSet *FullEquipMap, value uint64) uint64 {
	size := len(sets.activeSets)
	switch size {
	case 0:
		return value
	case 1:
		var count uint8 = 0
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Head])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Shoulder])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Chest])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Hand])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Leg])
		return bonusValue(&sets.activeSets[0], count, value)
	default:
		counts := make([]uint8, size)
		addIfInEachSet(&counts, sets.itemToSet, itemSet[Equip_Head])
		addIfInEachSet(&counts, sets.itemToSet, itemSet[Equip_Shoulder])
		addIfInEachSet(&counts, sets.itemToSet, itemSet[Equip_Chest])
		addIfInEachSet(&counts, sets.itemToSet, itemSet[Equip_Hand])
		addIfInEachSet(&counts, sets.itemToSet, itemSet[Equip_Leg])
		return bonusValueEach(&sets.activeSets, &counts, value)
	}
}

func (sets *SetBonus) CalcAndMultiplySolve(itemSet *SolvableEquipMap, value uint64) uint64 {
	size := len(sets.activeSets)
	switch size {
	case 0:
		return value
	case 1:
		var count uint8 = 0
		addIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Head])
		addIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Shoulder])
		addIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Chest])
		addIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Hand])
		addIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Leg])
		return bonusValue(&sets.activeSets[0], count, value)
	default:
		counts := make([]uint8, size)
		addIfInEachSetSolve(&counts, sets.itemToSet, itemSet[Equip_Head])
		addIfInEachSetSolve(&counts, sets.itemToSet, itemSet[Equip_Shoulder])
		addIfInEachSetSolve(&counts, sets.itemToSet, itemSet[Equip_Chest])
		addIfInEachSetSolve(&counts, sets.itemToSet, itemSet[Equip_Hand])
		addIfInEachSetSolve(&counts, sets.itemToSet, itemSet[Equip_Leg])
		return bonusValueEach(&sets.activeSets, &counts, value)
	}
}

func (sets *SetBonus) CountInAnySet(itemSet *FullEquipMap) uint8 {
	size := len(sets.activeSets)
	switch size {
	case 0:
		return 0
	default:
		var count uint8 = 0
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Head])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Shoulder])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Chest])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Hand])
		addIfInAnySet(&count, sets.itemToSet, itemSet[Equip_Leg])
		return count
	}
}

func addIfInAnySet(count *uint8, itemToSet map[uint32]int, item *FullItem) {
	if item != nil {
		_, include := itemToSet[item.ItemId()]
		if include {
			*count++
		}
	}
}

func addIfInEachSet(counts *[]uint8, itemToSet map[uint32]int, item *FullItem) {
	if item != nil {
		index, include := itemToSet[item.ItemId()]
		if include {
			(*counts)[index]++
		}
	}
}

func addIfInAnySetSolve(count *uint8, itemToSet map[uint32]int, item *SolvableItem) {
	if item != nil {
		_, include := itemToSet[item.ItemId]
		if include {
			*count++
		}
	}
}

func addIfInEachSetSolve(counts *[]uint8, itemToSet map[uint32]int, item *SolvableItem) {
	if item != nil {
		index, include := itemToSet[item.ItemId]
		if include {
			(*counts)[index]++
		}
	}
}

func bonusValue(setInfo *setInfo, count uint8, value uint64) uint64 {
	if count >= 4 {
		return value * setInfo.bonus2 * setInfo.bonus4 / (denominator * denominator)
	} else if count >= 2 {
		return value * setInfo.bonus2 / denominator
	} else {
		return value
	}
}

func bonusValueEach(sets *[]setInfo, counts *[]uint8, value uint64) uint64 {
	for index, count := range *counts {
		setInfo := &(*sets)[index]
		if count >= 4 {
			return value * setInfo.bonus2 * setInfo.bonus4 / (denominator * denominator)
		} else if count >= 2 {
			value = value * setInfo.bonus2 / denominator
		}
	}
	return value
}

type setInfo struct {
	spec   SpecType
	name   string
	bonus2 uint64
	bonus4 uint64
	items  []uint32
}

const (
	defaultBonus = 1025
	denominator  = 1000

	white_tiger_battlegear_2      = 1032
	white_tiger_battlegear_4      = 1024
	white_tiger_battlegear_4_tank = 1035 // gives 2% to dps, a bit more overall

	plate_lightning_bonus_2_miti  = 1013 // 1.3% bonus applies to death chance only, from sim
	plate_lightning_bonus_4_miti  = 1050 // compromise number, it's situational after all
	plate_lightning_bonus_4_dps   = 1027 // sim result for horridon h10, might not always apply
	plate_lightning_bonus_4_death = 1250 // actual result of sim for horridon h10
)

var g_setData = buildSets()

func buildSets() []setInfo {
	sets := make([]setInfo, 0)
	sets = append(sets, setInfo{Spec_PaladinProtMitigation, "White Tiger Battlegear Prot Mitigation", denominator, white_tiger_battlegear_4_tank, []uint32{86681, 86679, 86683, 86682, 86680, 85341, 85339, 85343, 85342, 85340, 87101, 87103, 87099, 87100, 87102}})
	sets = append(sets, setInfo{Spec_PaladinProtMitigation, "White Tiger Plate", defaultBonus, defaultBonus, []uint32{85319, 85320, 85321, 85322, 85323, 86659, 86660, 86661, 86662, 86663, 87109, 87110, 87111, 87112, 87113}})

	sets = append(sets, setInfo{Spec_PaladinProtMitigation, "Plate of the Lightning Emperor Prot Mitigation", plate_lightning_bonus_2_miti, plate_lightning_bonus_4_miti, []uint32{95290, 95291, 95292, 95293, 95294, 95920, 95921, 95922, 95923, 95924, 96664, 96665, 96666, 96667, 96668}})
	sets = append(sets, setInfo{Spec_PaladinProtDps, "Plate of the Lightning Emperor Prot Damage", denominator, plate_lightning_bonus_4_dps, []uint32{95290, 95291, 95292, 95293, 95294, 95920, 95921, 95922, 95923, 95924, 96664, 96665, 96666, 96667, 96668}})

	sets = append(sets, setInfo{Spec_PaladinProtMitigation, "Plate of Winged Triumph", defaultBonus, defaultBonus, []uint32{99026, 99027, 99028, 99029, 99031, 99126, 99127, 99128, 99129, 99130, 99364, 99368, 99369, 99370, 99371, 99593, 99594, 99595, 99596, 99598}})
	sets = append(sets, setInfo{Spec_PaladinRet, "White Tiger Battlegear", white_tiger_battlegear_2, white_tiger_battlegear_4, []uint32{85339, 85340, 85341, 85342, 85343, 86679, 86680, 86681, 86682, 86683, 87099, 87100, 87101, 87102, 87103}})
	sets = append(sets, setInfo{Spec_PaladinRet, "Battlegear of the Lightning Emperor", defaultBonus, defaultBonus, []uint32{95280, 95281, 95282, 95283, 95284, 95910, 95911, 95912, 95913, 95914, 96654, 96655, 96656, 96657, 96658}})
	sets = append(sets, setInfo{Spec_PaladinRet, "Battlegear of Winged Triumph", defaultBonus, defaultBonus, []uint32{98985, 98986, 98987, 99002, 99052, 99132, 99136, 99137, 99138, 99139, 99372, 99373, 99379, 99380, 99387, 99566, 99625, 99651, 99661, 99662}})
	sets = append(sets, setInfo{Spec_PaladinHoly, "White Tiger Vestments", defaultBonus, defaultBonus, []uint32{85344, 85345, 85346, 85347, 85348, 86684, 86685, 86686, 86687, 86688, 87104, 87105, 87106, 87107, 87108}})
	sets = append(sets, setInfo{Spec_PaladinHoly, "Vestments of the Lightning Emperor", defaultBonus, defaultBonus, []uint32{95285, 95286, 95287, 95288, 95289, 95915, 95916, 95917, 95918, 95919, 96659, 96660, 96661, 96662, 96663}})
	sets = append(sets, setInfo{Spec_PaladinHoly, "Vestments of Winged Triumph", defaultBonus, defaultBonus, []uint32{98979, 98980, 98982, 99003, 99076, 99124, 99125, 99133, 99134, 99135, 99374, 99375, 99376, 99377, 99378, 99626, 99648, 99656, 99665, 99666}})

	sets = append(sets, setInfo{Spec_WarriorArms, "Battleplate of Resounding Rings", defaultBonus, defaultBonus, []uint32{85329, 85330, 85331, 85332, 85333, 86669, 86670, 86671, 86672, 86673, 87192, 87193, 87194, 87195, 87196}})
	sets = append(sets, setInfo{Spec_WarriorArms, "Battleplate of the Last Mogu", defaultBonus, defaultBonus, []uint32{95330, 95331, 95332, 95333, 95334, 95986, 95987, 95988, 95989, 95990, 96730, 96731, 96732, 96733, 96734}})
	sets = append(sets, setInfo{Spec_WarriorArms, "Battleplate of the Prehistoric Marauder", defaultBonus, defaultBonus, []uint32{99034, 99035, 99036, 99046, 99047, 99197, 99198, 99199, 99200, 99206, 99411, 99412, 99413, 99414, 99418, 99559, 99560, 99561, 99602, 99603}})
	sets = append(sets, setInfo{Spec_WarriorProt, "Plate of Resounding Rings", defaultBonus, defaultBonus, []uint32{85324, 85325, 85326, 85327, 85328, 86664, 86665, 86666, 86667, 86668, 87197, 87198, 87199, 87200, 87201}})
	sets = append(sets, setInfo{Spec_WarriorProt, "Plate of the Last Mogu", defaultBonus, defaultBonus, []uint32{95335, 95336, 95337, 95338, 95339, 95991, 95992, 95993, 95994, 95995, 96735, 96736, 96737, 96738, 96739}})
	sets = append(sets, setInfo{Spec_WarriorProt, "Plate of the Prehistoric Marauder", defaultBonus, defaultBonus, []uint32{99030, 99032, 99033, 99037, 99038, 99195, 99196, 99201, 99202, 99203, 99407, 99408, 99409, 99410, 99415, 99557, 99558, 99562, 99563, 99597}})

	sets = append(sets, setInfo{Spec_Hunter, "Battlegear of the Saurok Stalker", defaultBonus, defaultBonus, []uint32{95255, 95256, 95257, 95258, 95259, 95882, 95883, 95884, 95885, 95886, 96626, 96627, 96628, 96629, 96630}})
	sets = append(sets, setInfo{Spec_Hunter, "Battlegear of the Unblinking Vigil", defaultBonus, defaultBonus, []uint32{99080, 99081, 99082, 99085, 99086, 99157, 99158, 99159, 99167, 99168, 99402, 99403, 99404, 99405, 99406, 99573, 99574, 99577, 99578, 99660}})
	sets = append(sets, setInfo{Spec_Hunter, "Yaungol Slayer Battlegear", defaultBonus, defaultBonus, []uint32{85294, 85295, 85296, 85297, 85298, 86634, 86635, 86636, 86637, 86638, 87002, 87003, 87004, 87005, 87006}})
	sets = append(sets, setInfo{Spec_Rogue, "Barbed Assassin Battlegear", defaultBonus, defaultBonus, []uint32{99006, 99007, 99008, 99009, 99010, 99112, 99113, 99114, 99115, 99116, 99348, 99349, 99350, 99355, 99356, 99629, 99630, 99631, 99634, 99635}})
	sets = append(sets, setInfo{Spec_Rogue, "Battlegear of the Thousandfold Blades", defaultBonus, defaultBonus, []uint32{85299, 85300, 85301, 85302, 85303, 86639, 86640, 86641, 86642, 86643, 87124, 87125, 87126, 87127, 87128}})
	sets = append(sets, setInfo{Spec_Rogue, "Nine-Tail Battlegear", defaultBonus, defaultBonus, []uint32{95305, 95306, 95307, 95308, 95309, 95935, 95936, 95937, 95938, 95939, 96679, 96680, 96681, 96682, 96683}})
	sets = append(sets, setInfo{Spec_PriestShadow, "Regalia of Ternion Glory", defaultBonus, defaultBonus, []uint32{99004, 99005, 99019, 99020, 99021, 99110, 99111, 99121, 99122, 99123, 99359, 99360, 99361, 99362, 99363, 99586, 99587, 99588, 99627, 99628}})
	sets = append(sets, setInfo{Spec_PriestShadow, "Regalia of the Exorcist", defaultBonus, defaultBonus, []uint32{95300, 95301, 95302, 95303, 95304, 95930, 95931, 95932, 95933, 95934, 96674, 96675, 96676, 96677, 96678}})
	sets = append(sets, setInfo{Spec_PriestShadow, "Regalia of the Guardian Serpent", defaultBonus, defaultBonus, []uint32{85364, 85365, 85366, 85367, 85368, 86704, 86705, 86706, 86707, 86708, 87119, 87120, 87121, 87122, 87123}})
	sets = append(sets, setInfo{Spec_PriestHoly, "Vestments of Ternion Glory", defaultBonus, defaultBonus, []uint32{99017, 99018, 99023, 99024, 99025, 99117, 99118, 99119, 99120, 99131, 99357, 99358, 99365, 99366, 99367, 99584, 99585, 99590, 99591, 99592}})
	sets = append(sets, setInfo{Spec_PriestHoly, "Vestments of the Exorcist", defaultBonus, defaultBonus, []uint32{95295, 95296, 95297, 95298, 95299, 95925, 95926, 95927, 95928, 95929, 96669, 96670, 96671, 96672, 96673}})
	sets = append(sets, setInfo{Spec_PriestHoly, "Vestments of the Guardian Serpent", defaultBonus, defaultBonus, []uint32{85359, 85360, 85361, 85362, 85363, 86699, 86700, 86701, 86702, 86703, 87114, 87115, 87116, 87117, 87118}})
	sets = append(sets, setInfo{Spec_DeathKnightDps, "Battlegear of the Lost Catacomb", defaultBonus, defaultBonus, []uint32{85334, 85335, 85336, 85337, 85338, 86674, 86675, 86676, 86677, 86678, 86913, 86914, 86915, 86916, 86917}})
	sets = append(sets, setInfo{Spec_DeathKnightDps, "Battleplate of Cyclopean Dread", defaultBonus, defaultBonus, []uint32{99057, 99058, 99059, 99066, 99067, 99186, 99187, 99192, 99193, 99194, 99335, 99336, 99337, 99338, 99339, 99571, 99572, 99608, 99609, 99639}})
	sets = append(sets, setInfo{Spec_DeathKnightDps, "Battleplate of the All-Consuming Maw", defaultBonus, defaultBonus, []uint32{95225, 95226, 95227, 95228, 95229, 95825, 95826, 95827, 95828, 95829, 96569, 96570, 96571, 96572, 96573}})
	sets = append(sets, setInfo{Spec_DeathKnightBlood, "Plate of Cyclopean Dread", defaultBonus, defaultBonus, []uint32{99039, 99040, 99048, 99049, 99060, 99179, 99188, 99189, 99190, 99191, 99323, 99324, 99325, 99330, 99331, 99564, 99604, 99605, 99640, 99652}})
	sets = append(sets, setInfo{Spec_DeathKnightBlood, "Plate of the All-Consuming Maw", defaultBonus, defaultBonus, []uint32{95230, 95231, 95232, 95233, 95234, 95830, 95831, 95832, 95833, 95834, 96574, 96575, 96576, 96577, 96578}})
	sets = append(sets, setInfo{Spec_DeathKnightBlood, "Plate of the Lost Catacomb", defaultBonus, defaultBonus, []uint32{85314, 85315, 85316, 85317, 85318, 86654, 86655, 86656, 86657, 86658, 86918, 86919, 86920, 86921, 86922}})
	sets = append(sets, setInfo{Spec_ShamanEnhance, "Battlegear of the Firebird", defaultBonus, defaultBonus, []uint32{85284, 85285, 85286, 85287, 85288, 86624, 86625, 86626, 86627, 86628, 87134, 87135, 87136, 87137, 87138}})
	sets = append(sets, setInfo{Spec_ShamanEnhance, "Battlegear of the Witch Doctor", defaultBonus, defaultBonus, []uint32{95315, 95316, 95317, 95318, 95319, 95945, 95946, 95947, 95948, 95949, 96689, 96690, 96691, 96692, 96693}})
	sets = append(sets, setInfo{Spec_ShamanEnhance, "Celestial Harmony Battlegear", defaultBonus, defaultBonus, []uint32{98977, 98983, 98984, 98992, 98993, 99101, 99102, 99103, 99104, 99105, 99340, 99341, 99342, 99343, 99347, 99615, 99616, 99649, 99650, 99663}})
	sets = append(sets, setInfo{Spec_ShamanElemental, "Celestial Harmony Regalia", defaultBonus, defaultBonus, []uint32{99087, 99088, 99089, 99090, 99091, 99092, 99093, 99094, 99095, 99106, 99332, 99333, 99334, 99344, 99345, 99579, 99580, 99645, 99646, 99647}})
	sets = append(sets, setInfo{Spec_ShamanRestoration, "Celestial Harmony Vestment", defaultBonus, defaultBonus, []uint32{98988, 98989, 98990, 98991, 99011, 99099, 99100, 99107, 99108, 99109, 99346, 99351, 99352, 99353, 99354, 99611, 99612, 99613, 99614, 99636}})
	sets = append(sets, setInfo{Spec_ShamanElemental, "Regalia of the Firebird", defaultBonus, defaultBonus, []uint32{85289, 85290, 85291, 85292, 85293, 86629, 86630, 86631, 86632, 86633, 87139, 87140, 87141, 87142, 87143}})
	sets = append(sets, setInfo{Spec_ShamanElemental, "Regalia of the Witch Doctor", defaultBonus, defaultBonus, []uint32{95320, 95321, 95322, 95323, 95324, 95950, 95951, 95952, 95953, 95954, 96694, 96695, 96696, 96697, 96698}})
	sets = append(sets, setInfo{Spec_ShamanRestoration, "Vestments of the Firebird", defaultBonus, defaultBonus, []uint32{85349, 85350, 85351, 85352, 85353, 86689, 86690, 86691, 86692, 86693, 87129, 87130, 87131, 87132, 87133}})
	sets = append(sets, setInfo{Spec_ShamanRestoration, "Vestments of the Witch Doctor", defaultBonus, defaultBonus, []uint32{95310, 95311, 95312, 95313, 95314, 95940, 95941, 95942, 95943, 95944, 96684, 96685, 96686, 96687, 96688}})
	sets = append(sets, setInfo{Spec_MageFrost, "Chronomancer Regalia", defaultBonus, defaultBonus, []uint32{99077, 99078, 99079, 99083, 99084, 99152, 99153, 99160, 99161, 99162, 99397, 99398, 99399, 99400, 99401, 99575, 99576, 99657, 99658, 99659}})
	sets = append(sets, setInfo{Spec_MageFrost, "Regalia of the Burning Scroll", defaultBonus, defaultBonus, []uint32{85374, 85375, 85376, 85377, 85378, 86714, 86715, 86716, 86717, 86718, 87007, 87008, 87009, 87010, 87011}})
	sets = append(sets, setInfo{Spec_MageFrost, "Regalia of the Chromatic Hydra", defaultBonus, defaultBonus, []uint32{95260, 95261, 95262, 95263, 95264, 95890, 95891, 95892, 95893, 95894, 96634, 96635, 96636, 96637, 96638}})
	sets = append(sets, setInfo{Spec_Warlock, "Regalia of the Horned Nightmare", defaultBonus, defaultBonus, []uint32{99045, 99053, 99054, 99055, 99056, 99096, 99097, 99098, 99204, 99205, 99416, 99417, 99424, 99425, 99426, 99567, 99568, 99569, 99570, 99601}})
	sets = append(sets, setInfo{Spec_Warlock, "Regalia of the Thousandfold Hells", defaultBonus, defaultBonus, []uint32{95325, 95326, 95327, 95328, 95329, 95981, 95982, 95983, 95984, 95985, 96725, 96726, 96727, 96728, 96729}})
	sets = append(sets, setInfo{Spec_Warlock, "Sha-Skin Regalia", defaultBonus, defaultBonus, []uint32{85369, 85370, 85371, 85372, 85373, 86709, 86710, 86711, 86712, 86713, 87187, 87188, 87189, 87190, 87191}})
	sets = append(sets, setInfo{Spec_MonkBrewmaster, "Armor of Seven Sacred Seals", defaultBonus, defaultBonus, []uint32{99050, 99051, 99063, 99064, 99065, 99140, 99141, 99142, 99143, 99144, 99382, 99383, 99384, 99385, 99386, 99565, 99606, 99607, 99643, 99644}})
	sets = append(sets, setInfo{Spec_MonkBrewmaster, "Armor of the Red Crane", defaultBonus, defaultBonus, []uint32{85384, 85385, 85386, 85387, 85388, 86724, 86725, 86726, 86727, 86728, 87094, 87095, 87096, 87097, 87098}})
	sets = append(sets, setInfo{Spec_MonkDps, "Battlegear of Seven Sacred Seals", defaultBonus, defaultBonus, []uint32{99071, 99072, 99073, 99074, 99075, 99145, 99146, 99154, 99155, 99156, 99392, 99393, 99394, 99395, 99396, 99555, 99556, 99653, 99654, 99655}})
	sets = append(sets, setInfo{Spec_MonkDps, "Battlegear of the Red Crane", defaultBonus, defaultBonus, []uint32{85394, 85395, 85396, 85397, 85398, 86734, 86735, 86736, 86737, 86738, 87084, 87085, 87086, 87087, 87088}})
	sets = append(sets, setInfo{Spec_MonkBrewmaster, "Fire-Charm Armor", defaultBonus, defaultBonus, []uint32{95275, 95276, 95277, 95278, 95279, 95905, 95906, 95907, 95908, 95909, 96649, 96650, 96651, 96652, 96653}})
	sets = append(sets, setInfo{Spec_MonkDps, "Fire-Charm Battlegear", defaultBonus, defaultBonus, []uint32{95265, 95266, 95267, 95268, 95269, 95895, 95896, 95897, 95898, 95899, 96639, 96640, 96641, 96642, 96643}})
	sets = append(sets, setInfo{Spec_MonkMistweaver, "Fire-Charm Vestments", defaultBonus, defaultBonus, []uint32{95270, 95271, 95272, 95273, 95274, 95900, 95901, 95902, 95903, 95904, 96644, 96645, 96646, 96647, 96648}})
	sets = append(sets, setInfo{Spec_MonkMistweaver, "Vestments of Seven Sacred Seals", defaultBonus, defaultBonus, []uint32{99061, 99062, 99068, 99069, 99070, 99147, 99148, 99149, 99150, 99151, 99381, 99388, 99389, 99390, 99391, 99552, 99553, 99554, 99641, 99642}})
	sets = append(sets, setInfo{Spec_MonkMistweaver, "Vestments of the Red Crane", defaultBonus, defaultBonus, []uint32{85389, 85390, 85391, 85392, 85393, 86729, 86730, 86731, 86732, 86733, 87089, 87090, 87091, 87092, 87093}})
	sets = append(sets, setInfo{Spec_DruidBear, "Armor of the Eternal Blossom", defaultBonus, defaultBonus, []uint32{85379, 85380, 85381, 85382, 85383, 86719, 86720, 86721, 86722, 86723, 86938, 86939, 86940, 86941, 86942}})
	sets = append(sets, setInfo{Spec_DruidBear, "Armor of the Haunted Forest", defaultBonus, defaultBonus, []uint32{95250, 95251, 95252, 95253, 95254, 95850, 95851, 95852, 95853, 95854, 96594, 96595, 96596, 96597, 96598}})
	sets = append(sets, setInfo{Spec_DruidBear, "Armor of the Shattered Vale", defaultBonus, defaultBonus, []uint32{98978, 98981, 98999, 99000, 99001, 99163, 99164, 99165, 99166, 99170, 99419, 99420, 99421, 99422, 99423, 99610, 99622, 99623, 99624, 99664}})
	sets = append(sets, setInfo{Spec_DruidFeral, "Battlegear of the Eternal Blossom", defaultBonus, defaultBonus, []uint32{85309, 85310, 85311, 85312, 85313, 86649, 86650, 86651, 86652, 86653, 86923, 86924, 86925, 86926, 86927}})
	sets = append(sets, setInfo{Spec_DruidFeral, "Battlegear of the Haunted Forest", defaultBonus, defaultBonus, []uint32{95235, 95236, 95237, 95238, 95239, 95835, 95836, 95837, 95838, 95839, 96579, 96580, 96581, 96582, 96583}})
	sets = append(sets, setInfo{Spec_DruidFeral, "Battlegear of the Shattered Vale", defaultBonus, defaultBonus, []uint32{99022, 99041, 99042, 99043, 99044, 99180, 99181, 99182, 99183, 99184, 99322, 99326, 99327, 99328, 99329, 99589, 99599, 99600, 99632, 99633}})
	sets = append(sets, setInfo{Spec_DruidBoom, "Regalia of the Eternal Blossom", defaultBonus, defaultBonus, []uint32{85304, 85305, 85306, 85307, 85308, 86644, 86645, 86646, 86647, 86648, 86933, 86934, 86935, 86936, 86937}})
	sets = append(sets, setInfo{Spec_DruidBoom, "Regalia of the Haunted Forest", defaultBonus, defaultBonus, []uint32{95245, 95246, 95247, 95248, 95249, 95845, 95846, 95847, 95848, 95849, 96589, 96590, 96591, 96592, 96593}})
	sets = append(sets, setInfo{Spec_DruidBoom, "Regalia of the Shattered Vale", defaultBonus, defaultBonus, []uint32{98994, 98995, 98996, 98997, 98998, 99169, 99174, 99175, 99176, 99177, 99427, 99428, 99432, 99433, 99434, 99617, 99618, 99619, 99620, 99621}})
	sets = append(sets, setInfo{Spec_DruidTree, "Vestments of the Eternal Blossom", defaultBonus, defaultBonus, []uint32{85354, 85355, 85356, 85357, 85358, 86694, 86695, 86696, 86697, 86698, 86928, 86929, 86930, 86931, 86932}})
	sets = append(sets, setInfo{Spec_DruidTree, "Vestments of the Haunted Forest", defaultBonus, defaultBonus, []uint32{95240, 95241, 95242, 95243, 95244, 95840, 95841, 95842, 95843, 95844, 96584, 96585, 96586, 96587, 96588}})
	sets = append(sets, setInfo{Spec_DruidTree, "Vestments of the Shattered Vale", defaultBonus, defaultBonus, []uint32{99012, 99013, 99014, 99015, 99016, 99171, 99172, 99173, 99178, 99185, 99429, 99430, 99431, 99435, 99436, 99581, 99582, 99583, 99637, 99638}})
	return sets
}
