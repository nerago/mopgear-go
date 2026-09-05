package weight_types

import (
	"encoding/json"
	"iter"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
)

func Weight1Basic_Make_CompatibleExternal_FromBlock(ratingWeight stats.StatBlockFloat) *Weight1_CompatibleExternal {
	return &Weight1_CompatibleExternal{weight1Internal{content: ratingWeight}}
}

type Weight1_ScaledSolvable struct {
	weight1Internal
	scaleAndOffset ScaleAndOffset
}

func Weight1_Make_ScaledSolvable() *Weight1_ScaledSolvable {
	return &Weight1_ScaledSolvable{}
}

func (wbs *Weight1_ScaledSolvable) Clone() Weight1_ScaledSolvable {
	return Weight1_ScaledSolvable{
		weight1Internal: weight1Internal{content: wbs.content.Clone()},
		scaleAndOffset:  wbs.scaleAndOffset,
	}
}

func (wbs *Weight1_ScaledSolvable) ConvertToExternal(requiredStats []stats.StatType) Weight1_CompatibleExternal {
	weightExternal := Weight1_CompatibleExternal{
		weight1Internal: wbs.weight1Internal,
	}

	weightExternal.NormalizeForBase(requiredStats)

	return weightExternal
}

func (wbs *Weight1_ScaledSolvable) Equals(other *Weight1_ScaledSolvable) bool {
	return wbs.weight1Internal.Equals(&other.weight1Internal) &&
		wbs.scaleAndOffset == other.scaleAndOffset
}

func (wbs *Weight1_ScaledSolvable) CalcStatScore(stats *stats.StatBlock) float64 {
	rawValue := wbs.calcStatScoreRaw(stats)
	return wbs.scaleAndOffset.Apply(rawValue)
}

func (wbs *Weight1_ScaledSolvable) CalcStatScoreRaw(stats *stats.StatBlock) float64 {
	return wbs.calcStatScoreRaw(stats)
}

func (wbs *Weight1_ScaledSolvable) GetScaleOffset() ScaleAndOffset {
	return wbs.scaleAndOffset
}

func (wbs *Weight1_ScaledSolvable) SetScaleOffset(scaleOffset ScaleAndOffset) {
	wbs.scaleAndOffset = scaleOffset
}

func (wbs *Weight1_ScaledSolvable) UpdateScaling(sampleInputs []WeightInput) error {
	scaleOffset, err := updateScalingBasic1(sampleInputs, wbs.calcStatScoreRaw)
	if err != nil {
		return err
	} else {
		wbs.scaleAndOffset = scaleOffset
		return nil
	}
}

func (wbs *Weight1_ScaledSolvable) FinishAndValidate(sampleInputs []WeightInput) error {
	err := wbs.validateTypes()
	if err != nil {
		return err
	}

	err = wbs.verifyGoodRange(sampleInputs)
	if err != nil {
		return err
	}

	return nil
}

func (wbs *Weight1_ScaledSolvable) validateTypes() error {
	if !stats.IsUsefulWeightNumber(wbs.scaleAndOffset.Scale) || util.FloatEqualsZero(wbs.scaleAndOffset.Scale) {
		return util.ErrorTracedNewFormat("bad scale %f", wbs.scaleAndOffset.Scale)
	}
	if wbs.content.IsEmpty() {
		return util.ErrorTracedNewFormat("empty weight")
	}
	return nil
}

type Weight1_CompatibleExternal struct {
	weight1Internal
}

func (wbx *Weight1_CompatibleExternal) Clone() Weight1_CompatibleExternal {
	return Weight1_CompatibleExternal{
		weight1Internal: weight1Internal{content: wbx.content.Clone()},
	}
}

func (wbx *Weight1_CompatibleExternal) ConvertToSolvable(sampleInputs []WeightInput) (Weight1_ScaledSolvable, error) {
	weightSolve := Weight1_ScaledSolvable{
		weight1Internal: wbx.weight1Internal,
	}

	err := weightSolve.UpdateScaling(sampleInputs)
	if err != nil {
		return Weight1_ScaledSolvable{}, err
	}

	return weightSolve, nil
}

func (wbx *Weight1_CompatibleExternal) NormalizeForBase(requiredStats []stats.StatType) {
	factor := 1.0
	foundBase := false

	for _, tryBaseStat := range requiredStats {
		value := wbx.Get(tryBaseStat)
		if util.FloatNonZero(value) && value > 0 {
			factor = 1.0 / value
			foundBase = true
			break
		}
	}

	if !foundBase {
		for _, tryBaseStat := range requiredStats {
			value := wbx.Get(tryBaseStat)
			if value > 0 {
				factor = 1.0 / value
				foundBase = true
				break
			}
		}
	}

	wbx.content.MultiplyScalar(factor, &wbx.content)
}

type weight1Internal struct {
	content stats.StatBlockFloat
}

//func Weight1BasicInternal_Of(values []float64, statTypes []stats.StatType) weight1BasicInternal {
//	wr := weight1BasicInternal{}
//	for i, statType := range statTypes {
//		wr.Put(statType, values[i])
//	}
//	return wr
//}
//

func (wb *weight1Internal) IsEmpty() bool {
	return wb.content.IsEmpty()
}

func (wb *weight1Internal) IsOverlySimple() bool {
	zero, one, interesting := 0, 0, 0
	for value := range wb.content.SeqValues() {
		if stats.IsValidWeightNumber(value) {
			if util.FloatEqualsZero(value) {
				zero++
			} else if util.FloatEqualsOne(value) {
				one++
			} else {
				interesting++
			}
		}
	}
	if interesting <= 1 || (interesting+one) <= 2 {
		return true
	}

	distinct := make([]float64, 0, 4)
	for value := range wb.content.SeqValues() {
		if !slices.Contains(distinct, value) {
			distinct = append(distinct, value)
		}
	}
	return len(distinct) <= 2
}

func (wb *weight1Internal) Get(statType stats.StatType) float64 {
	return wb.content.GetFloat(statType)
}

func (wb *weight1Internal) IsZero(statType stats.StatType) bool {
	return util.FloatEqualsZero(wb.content.GetFloat(statType))
}

func (wb *weight1Internal) Put(statType stats.StatType, value float64) {
	wb.content[statType] = value
}

func (wb *weight1Internal) PlusEquals(statType stats.StatType, value float64) {
	wb.content[statType] += value
}

func (wb *weight1Internal) MinusEquals(statType stats.StatType, value float64) {
	wb.content[statType] -= value
}

func (wb *weight1Internal) MultiplyEquals(statType stats.StatType, value float64) {
	wb.content[statType] *= value
}

func (wb *weight1Internal) DivideEquals(statType stats.StatType, value float64) {
	wb.content[statType] /= value
}

func (wb *weight1Internal) Equals(other *weight1Internal) bool {
	return wb.content.Equals(&other.content)
}

func (wb *weight1Internal) calcStatScoreRaw(stats *stats.StatBlock) float64 {
	return wb.content.MultiplyForTotalSum2(stats)
}

func (wb *weight1Internal) CalcStatScoreWithBonus(_ *stats.StatBlock, _ *stats.SimTypeMap[float64]) float64 {
	panic("Weight1Basic can't handle bonus map since doesn't have any sim breakdowns")
}

func (wb *weight1Internal) String() string {
	return wb.content.CreateString(6)
}

func (wb *weight1Internal) SeqPair() iter.Seq2[stats.StatType, float64] {
	return wb.content.SeqPair()
}

func (wb *weight1Internal) MarshalJSON() ([]byte, error) {
	return json.Marshal(wb.content)
}

func (wb *weight1Internal) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &wb.content)
}
