package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearItemSetupEx struct {
	requireRows   stats.StatTypeMap[*util_highs.ConstraintRow] // constrains values for the hit/expertise/etc of each item
	statTotalRows stats.StatTypeMap[*util_highs.ConstraintRow]
}

func (site *gearItemSetupEx) addItem(item *items.SolvableItem, require *stats.StatTypeMap[weight_types.StatRangeFloat], columnIndex util_highs.ColumnIndex) {
	// add to stats via a summation condition
	for statType, value := range item.Total().SeqPairInt() {
		if value != 0 {
			site.statTotalRows.GetOrPanic(statType).Add(columnIndex, float64(value))
		}
	}

	// specific hit/expertise/etc values for hi/lo limits
	for statType := range require.SeqKey() {
		value := item.Total().GetFloat(statType)
		if value != 0 {
			site.requireRows.GetOrPanic(statType).Add(columnIndex, value)
		}
	}
}

func (site *gearItemSetupEx) prepareRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	for statType := range require.SeqKey() {
		site.requireRows.Put(statType, &util_highs.ConstraintRow{Debug: "require " + statType.Name()})
	}
}

func (site *gearItemSetupEx) prepareStatTotals() {
	for _, statType := range stats.StatType_List {
		site.statTotalRows.Put(statType, &util_highs.ConstraintRow{Debug: "statTotal " + statType.Name()})
	}
}

func (site *gearItemSetupEx) finishRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat], build *util_highs.LinearBuilder) {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range require.SeqKeyValue() {
		row := site.requireRows.GetOrPanic(statType)
		row.Build(build, hilo.Minimum, hilo.Maximum)
	}
}

func (site *gearItemSetupEx) finishStatTotals(build *util_highs.LinearBuilder) (statTotalColumns *stats.StatTypeMap[*columnInfo]) {
	statTotalColumns = new(stats.StatTypeMap[*columnInfo])
	// constrain: total sum of each stat for input to weights
	for _, statType := range stats.StatType_List {
		entry := columnInfo{entryType: entry_stat_total, statType: statType}
		entry.columnIndex = build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), util_highs.DebugText("statTotal "+statType.Name()))
		statTotalColumns.Put(statType, &entry)

		row := site.statTotalRows.GetOrPanic(statType)
		row.Add(entry.columnIndex, -1)
	}
	return statTotalColumns
}
