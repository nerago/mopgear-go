package solve_highs

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearItemSetupBasic struct {
	baseRatingSumRow util_highs.ConstraintRow // values for the ratings of each item

	hitValueRow    util_highs.ConstraintRow // constrains values for the hits of each item
	expertValueRow util_highs.ConstraintRow // constrains values for the expertise of each item

	minimumValueType stats.StatType
	minimumValueRow  util_highs.ConstraintRow // when an extra minimum is specified
}

func (sbb *gearItemSetupBasic) addItem(item *items.SolvableItem, calcRating func(item *items.SolvableItem) float64, columnIndex util_highs.ColumnIndex) {
	// add rating via a summation condition
	rating := calcRating(item)
	sbb.baseRatingSumRow.Add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	sbb.hitValueRow.Add(columnIndex, item.Total().GetFloat(stats.Stat_Hit))
	sbb.expertValueRow.Add(columnIndex, item.Total().GetFloat(stats.Stat_Expertise))

	// additional minimum value (e.g. haste)
	if sbb.minimumValueType != stats.Stat_Invalid {
		sbb.minimumValueRow.Add(columnIndex, item.Total().GetFloat(sbb.minimumValueType))
	}
}

func (sbb *gearItemSetupBasic) prepareRequire(statRequirements *stats.StatTypeMap[weight_types.StatRangeFloat]) error {
	sbb.minimumValueType = stats.Stat_Invalid
	for statType := range statRequirements.SeqKey() {
		if statType != stats.Stat_Hit && statType != stats.Stat_Expertise {
			if sbb.minimumValueType == stats.Stat_Invalid {
				sbb.minimumValueType = statType
			} else {
				return util.ErrorTracedNew("multiple additional required stats not supported in basic weights mode")
			}
		}
	}
	return nil
}

func (sbb *gearItemSetupBasic) finishRequire1(require *stats.StatTypeMap[weight_types.StatRangeFloat], build *util_highs.LinearBuilder) error {
	// constrain: total sum of hit/exp are within requested limits
	if hitRange, hasHit := require.Get(stats.Stat_Hit); hasHit {
		sbb.hitValueRow.Debug = "hitValueRow"
		sbb.hitValueRow.Build(build, hitRange.Minimum, hitRange.Maximum)
	}

	if expertRange, hasExpert := require.Get(stats.Stat_Expertise); hasExpert {
		sbb.expertValueRow.Debug = "expertValueRow"
		sbb.expertValueRow.Build(build, expertRange.Minimum, expertRange.Maximum)
	}

	// constrain: additional minimum value if specified has required minimum
	if sbb.minimumValueType != stats.Stat_Invalid {
		if otherRange, hasRange := require.Get(sbb.minimumValueType); hasRange {
			sbb.minimumValueRow.Build(build, otherRange.Minimum, otherRange.Maximum)
		} else {
			return util.ErrorTracedNew("missing require value")
		}
	}

	return nil
}

func (sbb *gearItemSetupBasic) finishRatingSum(build *util_highs.LinearBuilder, scaleOffset weight_types.ScaleAndOffset) (baseRatingSumVar *columnInfo) {
	sumColumn := &columnInfo{entryType: entry_sum_rating}

	// sum of individual selected item ratings
	sumColumn.columnIndex = build.CreateColumnGeneral(highs.Continuous, 0, c_basic_ratingsMax, sumColumn)

	// main action of this variable: derive value to match rest of row sum
	// apply scale and offset factors given by the weights
	sbb.baseRatingSumRow.Debug = "baseRatingSumRow"
	offset := -scaleOffset.Offset
	ratingScale := -1.0 / scaleOffset.Scale
	sbb.baseRatingSumRow.Add(sumColumn.columnIndex, ratingScale)
	sbb.baseRatingSumRow.Build(build, offset, offset)

	// save reference
	return sumColumn
}
