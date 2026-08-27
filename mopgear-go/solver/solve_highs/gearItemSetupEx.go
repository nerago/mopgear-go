package solve_highs

import (
	"errors"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearItemSetupEx struct {
	requireRows   stats.StatTypeMap[*util_highs.ConstraintRow] // constrains values for the hit/expertise/etc of each item
	statTotalRows stats.StatTypeMap[*util_highs.ConstraintRow]
}

func (site *gearItemSetupEx) addItem(item *items.SolvableItem, require *stats.StatTypeMap[weight_types.StatRangeFloat], columnIndex util_highs.ColumnIndex) error {
	// add to stats via a summation condition
	for statType, value := range item.Total().SeqPairInt() {
		if value != 0 {
			if row, hasRow := site.statTotalRows.Get(statType); hasRow {
				row.Add(columnIndex, float64(value))
			} else {
				return errors.New("missing stat row")
			}
		}
	}

	// specific hit/expertise/etc values for hi/lo limits
	for statType := range require.SeqKey() {
		value := item.Total().GetFloat(statType)
		if value != 0 {
			if row, hasRow := site.requireRows.Get(statType); hasRow {
				row.Add(columnIndex, value)
			} else {
				return errors.New("missing require row")
			}
		}
	}

	return nil
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

func (site *gearItemSetupEx) finishRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat], build *util_highs.LinearBuilder) error {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range require.SeqKeyValue() {
		if row, hasRow := site.requireRows.Get(statType); hasRow {
			row.Build(build, hilo.Minimum, hilo.Maximum)
		} else {
			return errors.New("missing requireRow")
		}
	}
	return nil
}

func (site *gearItemSetupEx) finishStatTotals(build *util_highs.LinearBuilder) (*stats.StatTypeMap[*columnInfo], error) {
	statTotalColumns := new(stats.StatTypeMap[*columnInfo])
	// constrain: total sum of each stat for input to weights
	for _, statType := range stats.StatType_List {
		entry := &columnInfo{entryType: entry_stat_total, statType: statType}
		entry.columnIndex = build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), entry)
		statTotalColumns.Put(statType, entry)

		if row, hasRow := site.statTotalRows.Get(statType); hasRow {
			row.Add(entry.columnIndex, -1)
			row.Build(build, 0, 0)
		} else {
			return nil, errors.New("missing statTotalRow")
		}
	}
	return statTotalColumns, nil
}
