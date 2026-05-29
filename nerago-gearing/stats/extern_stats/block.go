package extern_stats

import (
	gear_stat "paladin_gearing_go/stats"
	"strconv"

	wowsim_proto "github.com/wowsims/mop/sim/core/proto"
	wowsim_stat "github.com/wowsims/mop/sim/core/stats"
)

func GearStatBlockToSimStat(block gear_stat.StatBlock) wowsim_stat.Stats {
	stats := wowsim_stat.Stats{}
	for index := range block {
		gearStat := gear_stat.StatType(index)
		simStat := GearStatToSimStat(gearStat)
		stats[simStat] = float64(block[gearStat])
	}
	return stats
}

func GearStatBlockToUnitStats(block *gear_stat.StatBlock) *wowsim_proto.UnitStats {
	unitStats := &wowsim_proto.UnitStats{}
	unitStats.Stats = make([]float64, 12)
	for index := range block {
		gearStat := gear_stat.StatType(index)
		simStat := GearStatToSimStat(gearStat)
		unitStats.Stats[simStat] = float64(block[gearStat])
	}
	return unitStats
}

func GearStatMapToUnitStats(statMap map[gear_stat.StatType]int32) *wowsim_proto.UnitStats {
	unitStats := &wowsim_proto.UnitStats{}
	unitStats.Stats = make([]float64, 12)
	for gearStat, value := range statMap {
		simStat := GearStatToSimStat(gearStat)
		unitStats.Stats[simStat] = float64(value)
	}
	return unitStats
}

func SimStatsToGearStatBlock(stats wowsim_stat.Stats) gear_stat.StatBlock {
	block := gear_stat.StatBlock{}
	for index := range block {
		gearStat := gear_stat.StatType(index)
		simStat := GearStatToSimStat(gearStat)
		block[gearStat] = uint32(stats[simStat])
	}
	return block
}

func SimJsonArrayToGearStatBlock(input []any) gear_stat.StatBlock {
	block := gear_stat.StatBlock{}
	for indexNum, value := range input {
		simStat := wowsim_stat.Stat(indexNum)
		gearStat, inRange, supported := SimStatToGearStat(simStat)
		if inRange && supported {
			block[gearStat] = uint32(value.(float64))
		}
	}
	return block
}

func SimJsonMapToGearStatBlock(input map[string]any) gear_stat.StatBlock {
	block := gear_stat.StatBlock{}
	for indexStr, value := range input {
		indexNum, err := strconv.Atoi(indexStr)
		if err != nil {
			panic(err)
		}
		simStat := wowsim_stat.Stat(indexNum)

		gearStat, inRange, supported := SimStatToGearStat(simStat)
		if inRange && supported {
			block[gearStat] = uint32(value.(float64))
		} else if !inRange {
			panic("unknown stat index")
		}
	}
	return block
}
