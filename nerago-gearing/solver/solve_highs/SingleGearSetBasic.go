package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_single_basic_scaled_ratings = 1.0e-8 // try to make highs happier
	// example rating 138082172 1.38e8
	c_single_basic_ratings_high_range = 1.0e10 * c_single_basic_scaled_ratings
)

func SingleGearSetMain(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	sb := makeGearSetBasic(&build)
	outputVar := sb.createOutputVariableForSeparateRun()
	sb.setup(model, itemOptions, outputVar)
	return sb.runForFutureResult(itemOptions, model, printer)
}

func makeGearSetBasic(build *util_highs.LinearBuilder) *singleGearSetBasic {
	return &singleGearSetBasic{
		singleGearSetShared: singleGearSetShared{
			build:          build,
			ratingPreScale: c_single_basic_scaled_ratings,
		},
	}
}

func (sb *singleGearSetBasic) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, outputVar *columnInfo) {
	sb.prepareCommon(model, itemOptions)
	sb.prepareRequire1(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		sb.addItem(slot, item, model.SetBonusIndexForItem, model.CalcRatingItem)
	}

	sb.finishItemsCommon(itemOptions)
	sb.finishRequire1(&model.StatRequirements)
	sb.calcRatingSum()

	sb.multiplyByActiveCombo(sb.baseRatingSumVar, outputVar, c_single_basic_ratings_high_range,
		func(combo *bonusCombo) float64 {
			return combo.totalFlatMultiplier()
		})
}

type singleGearSetBasic struct {
	singleGearSetShared

	hitValueRow      util_highs.ConstraintRow // constrains values for the hits of each item
	expertValueRow   util_highs.ConstraintRow // constrains values for the expertise of each item
	minimumValueType stats.StatType
	minimumValueRow  util_highs.ConstraintRow // when an extra minimum is specified

	baseRatingSumRow util_highs.ConstraintRow // values for the ratings of each item
	baseRatingSumVar *columnInfo              // sum of values for the ratings of selected items
}

func (sb *singleGearSetBasic) calcRatingSum() {
	entry := columnInfo{entryType: entry_sum_rating}

	// sum of individual selected item ratings
	// doesen't go directly into output rating
	entry.columnIndex = sb.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), &entry)

	// main action of this variable: derive value to match rest of row sum
	sb.baseRatingSumRow.Debug = "baseRatingSumRow"
	sb.baseRatingSumRow.Add(entry.columnIndex, -1)
	sb.baseRatingSumRow.Build(sb.build, 0, 0)

	// save reference
	sb.baseRatingSumVar = &entry
	sb.allColumns = append(sb.allColumns, &entry)
}

func (sb *singleGearSetBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, activeSet func(id items.ItemId) (int, bool), calcRating func(item *items.SolvableItem) float64) util_highs.ColumnIndex {
	columnIndex := sb.addItemCommon(itemSlot, item, activeSet)

	// add rating via a summation condition
	// scale down ratings to keep numbers small for solver stability
	rating := calcRating(item) * c_single_basic_scaled_ratings
	sb.baseRatingSumRow.Add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	sb.hitValueRow.Add(columnIndex, item.Total().GetFloat(stats.Stat_Hit))
	sb.expertValueRow.Add(columnIndex, item.Total().GetFloat(stats.Stat_Expertise))

	// additional minimum value (e.g. haste)
	if sb.minimumValueType != stats.Stat_Invalid {
		sb.minimumValueRow.Add(columnIndex, item.Total().GetFloat(sb.minimumValueType))
	}

	return columnIndex
}

func (sb *singleGearSetBasic) prepareRequire1(statRequirements *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	sb.minimumValueType = stats.Stat_Invalid
	for statType := range statRequirements.SeqKey() {
		if statType != stats.Stat_Hit && statType != stats.Stat_Expertise {
			if sb.minimumValueType == stats.Stat_Invalid {
				sb.minimumValueType = statType
			} else {
				panic("multiple additional required stats not supported in basic weights mode")
			}
		}
	}
}

func (sb *singleGearSetBasic) finishRequire1(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	// constrain: total sum of hit/exp are within requested limits
	if hitRange, hasHit := require.Get(stats.Stat_Hit); hasHit {
		sb.hitValueRow.Debug = "hitValueRow"
		sb.hitValueRow.Build(sb.build, hitRange.Minimum, hitRange.Maximum)
	}

	if expertRange, hasExpert := require.Get(stats.Stat_Expertise); hasExpert {
		sb.expertValueRow.Debug = "expertValueRow"
		sb.expertValueRow.Build(sb.build, expertRange.Minimum, expertRange.Maximum)
	}

	// constrain: additional minimum value if specified has required minimum
	if sb.minimumValueType != stats.Stat_Invalid {
		otherRange := require.GetOrPanic(sb.minimumValueType)
		sb.minimumValueRow.Build(sb.build, otherRange.Minimum, otherRange.Maximum)
	}
}
