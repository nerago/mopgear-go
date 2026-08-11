package bonus_set

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
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
		bonus0 := defaultMap()
		pb.simBonus[0] = bonus0
		pb.simBonus[1] = bonus0
		bonus2 := convertToSimMap(extendedData[0])
		pb.simBonus[2] = bonus2
		pb.simBonus[3] = bonus2
		bonus4 := convertToSimMap(extendedData[1])
		pb.simBonus[4] = bonus4
		pb.simBonus[5] = bonus4
	}
	return pb
}

func defaultMap() *stats.SimTypeMap[float64] {
	simMap := &stats.SimTypeMap[float64]{}
	for _, simType := range stats.SimTypeList {
		simMap.Put(simType, 1)
	}
	return simMap
}

func convertToSimMap(ex map[stats.SimType]float64) *stats.SimTypeMap[float64] {
	simMap := &stats.SimTypeMap[float64]{}
	for _, simType := range stats.SimTypeList {
		if value, hasValue := ex[simType]; hasValue {
			simMap.Put(simType, value)
		} else {
			simMap.Put(simType, 1)
		}
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

func (pb *PreparedBonus) deriveUpgradedFlatBonus(priority *weight_types.SimPriorityBasic) {
	for i := range pb.simBonus {
		bonusMap := pb.simBonus[i]
		if bonusMap != nil {
			sum := pb.deriveUpgradedFlatBonusSingle(pb.simBonus[i], priority)
			if sum != 0 {
				pb.flatBonus[i] = sum
			}
		}
	}
}

func (pb *PreparedBonus) deriveUpgradedFlatBonusSingle(bonusMap *stats.SimTypeMap[float64], priority *weight_types.SimPriorityBasic) float64 {
	sum := 0.0
	for simType, bonus := range bonusMap.SeqKeyValue() {
		ratio, hasRatio := priority.Get(simType)
		if hasRatio {
			sum += ratio * bonus
		}
	}
	return sum
}
