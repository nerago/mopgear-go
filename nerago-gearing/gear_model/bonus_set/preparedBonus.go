package bonus_set

import (
	"paladin_gearing_go/stats"
	"slices"
)

type PreparedBonus struct {
	BonusLookup
	flatBonus BonusByCountFlat
	simBonus  BonusByCountBySim
}

func preparedBonusMake(common dataEntry, variant dataEntryVariant) PreparedBonus {
	pb := PreparedBonus{
		BonusLookup: BonusLookup{
			name:  variant.name,
			items: common.items,
		},
		flatBonus: [6]float64{
			0: 1.0,
			1: 1.0,
			2: variant.bonus2,
			3: variant.bonus2,
			4: variant.bonus4,
			5: variant.bonus4,
		},
	}
	if extendedData, hasExtended := g_extendedData[variant.name]; hasExtended {
		bonus2 := convertToSimMap(extendedData[0])
		bonus4 := convertToSimMap(extendedData[1])
		pb.simBonus[2] = bonus2
		pb.simBonus[3] = bonus2
		pb.simBonus[4] = bonus4
		pb.simBonus[5] = bonus4
	}
	return pb
}

func convertToSimMap(ex map[stats.SimType]float64) *stats.SimTypeMap[float64] {
	simMap := &stats.SimTypeMap[float64]{}
	for simType, value := range ex {
		simMap.Put(simType, value)
	}
	return simMap
}

func (pb *PreparedBonus) Equals(other *PreparedBonus) bool {
	return pb.name == other.name && slices.Equal(pb.items, other.items)
}

func (pb *PreparedBonus) BonusByCount() BonusByCountFlat {
	return pb.flatBonus
}

func (pb *PreparedBonus) BonusByCountBySim() BonusByCountBySim {
	return pb.simBonus
}
